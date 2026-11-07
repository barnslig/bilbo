package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/olivere/vite"
	"zgo.at/zcache/v2"
)

//go:embed all:assets/dist
var assets embed.FS

//go:embed all:templates
var templates embed.FS

type key string

type BilboConfig struct {
	DataDir    string
	HttpListen string
	IsDev      bool
	Secret     string
}

type Bilbo struct {
	assetFs fs.FS
	cfg     BilboConfig
	hndl    http.Handler
	mux     *mux.Router
	repo    *git.Repository
	tmpl    *template.Template
	tmplFs  fs.FS

	pageCache            *zcache.Cache[string, *Page]
	pageHistoryCache     *zcache.Cache[string, []*object.Commit]
	folderStructureCache *zcache.Cache[string, *FolderStructure]
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

	if b.cfg.IsDev {
		b.assetFs = os.DirFS("assets/src/")
	} else {
		b.assetFs, _ = fs.Sub(assets, "assets/dist")
	}

	if b.cfg.IsDev {
		b.tmplFs = os.DirFS("templates")
	} else {
		b.tmplFs, _ = fs.Sub(templates, "templates")
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

	b.tmpl = template.Must(template.New("").Funcs(tmplFuncMap).ParseFS(b.tmplFs, "*.html"))

	// Create caches
	b.pageCache = zcache.New[string, *Page](time.Hour, 10*time.Minute)
	b.pageHistoryCache = zcache.New[string, []*object.Commit](time.Hour, 10*time.Minute)
	b.folderStructureCache = zcache.New[string, *FolderStructure](time.Hour, 10*time.Minute)

	// Create HTTP routes
	b.mux = mux.NewRouter()
	b.mux.HandleFunc("/", b.HandleIndex).Methods("GET").Name("index")

	b.mux.PathPrefix("/assets/").Handler(http.FileServer(http.FS(b.assetFs)))
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
	b.hndl = b.AuthorMiddleware(b.hndl)
	b.hndl = csrf.Protect([]byte(b.cfg.Secret), csrf.Secure(false))(b.hndl)

	log.Printf("Now listening on http://%s", b.cfg.HttpListen)
	err = http.ListenAndServe(b.cfg.HttpListen, b.hndl)

	return
}

type hash map[string]interface{}

func (b *Bilbo) renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, localData hash) {
	viteFragment, err := vite.HTMLFragment(vite.Config{
		FS:           b.assetFs,
		IsDev:        b.cfg.IsDev,
		ViteEntry:    "src/app.js",
		ViteTemplate: vite.Vanilla,
	})

	if err != nil {
		panic(err)
	}

	// Global template data
	data := hash{
		"csrfToken":      csrf.Token(r),
		csrf.TemplateTag: csrf.TemplateField(r),
		"gitIsHead":      r.Context().Value(key("GitIsHead")).(bool),
		"Vite":           viteFragment,
	}

	// Merge the local template data with the global one
	for k, v := range localData {
		data[k] = v
	}

	b.tmpl.ExecuteTemplate(w, templateFile, data)
}
