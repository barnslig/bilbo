package main

import (
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"gopkg.in/src-d/go-git.v4"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

type BilboConfig struct {
	DataDir     string
	HttpListen  string
	Secret      string
	StaticDir   string
	TemplateDir string
}

type Bilbo struct {
	cache *cache.Cache
	cfg   BilboConfig
	hndl  http.Handler
	mux   *mux.Router
	repo  *git.Repository
	tmpl  *template.Template
}

func NewBilbo(cfg BilboConfig) (b *Bilbo, err error) {
	b = &Bilbo{cfg: cfg}

	// Open or create the GIT repository
	b.repo, err = git.PlainOpen(b.cfg.DataDir)
	if err == git.ErrRepositoryNotExists {
		b.repo, err = git.PlainInit(b.cfg.DataDir, false)
	}
	if err != nil {
		return
	}

	// Precompile templates
	tmplFuncMap := template.FuncMap{
		"route": func(viewName string, args ...interface{}) string {
			var strArgs []string
			for _, arg := range args {
				if arg == nil {
					arg = ""
				}
				strArgs = append(strArgs, fmt.Sprintf("%v", arg))
			}

			route, err := b.mux.Get(viewName).URL(strArgs...)
			if err != nil {
				return "/"
			}

			routeStr := route.String()
			hasTrailingSlash := routeStr[len(routeStr)-1:] == "/"
			cleanRouteStr := path.Clean(routeStr)

			if hasTrailingSlash && cleanRouteStr != "/" {
				return cleanRouteStr + "/"
			}

			return cleanRouteStr
		},
	}
	b.tmpl = template.Must(template.New("").Funcs(tmplFuncMap).ParseGlob(path.Join(b.cfg.TemplateDir, "*.html")))

	// Create cache
	b.cache = cache.New(time.Hour, 10*time.Minute)

	// Create HTTP routes
	b.mux = mux.NewRouter()
	b.mux.HandleFunc("/", b.HandleIndex).Methods("GET").Name("index")

	b.mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(b.cfg.StaticDir))))
	b.mux.HandleFunc("/{special:favicon|favicon.ico}", http.NotFound)

	b.mux.HandleFunc("/edit/_new/{folder:.*}", b.HandleEditNew).Methods("GET", "POST").Name("pages#new")
	b.mux.HandleFunc("/edit/_preview", b.HandleEditPreview).Methods("POST").Name("pages#preview")
	b.mux.HandleFunc("/edit/_rename/{page:.+}", b.HandleEditRename).Methods("GET", "POST").Name("pages#rename")
	b.mux.HandleFunc("/edit/_delete/{page:.+}", b.HandleEditDelete).Methods("GET", "POST").Name("pages#destroy")
	b.mux.HandleFunc("/edit/{page:.+}", b.HandleEdit).Methods("GET", "POST").Name("pages#edit")

	b.mux.HandleFunc("/history/{page:.+}", b.HandleHistory).Methods("GET").Name("pages#history")

	b.mux.HandleFunc("/pages/{folder:.*}", b.HandlePages).Methods("GET").Name("pages#index")
	b.mux.HandleFunc("/{page:.+}", b.HandlePage).Methods("GET").Name("pages#show")

	// Apply middlewares
	b.hndl = RecoverMiddleware(b.mux)
	b.hndl = b.GitMiddleware(b.hndl)
	b.hndl = csrf.Protect([]byte(b.cfg.Secret), csrf.Secure(false))(b.hndl)

	log.Printf("Now listening on http://%s", b.cfg.HttpListen)
	err = http.ListenAndServe(b.cfg.HttpListen, b.hndl)

	return
}

type hash map[string]interface{}

func (b *Bilbo) renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, localData hash) {
	// Global template data
	data := hash{
		"csrfToken":      csrf.Token(r),
		csrf.TemplateTag: csrf.TemplateField(r),
		"gitIsHead":      r.Context().Value("GitIsHead").(bool),
	}

	// Merge the local template data with the global one
	for k, v := range localData {
		data[k] = v
	}

	b.tmpl.ExecuteTemplate(w, templateFile, data)
}
