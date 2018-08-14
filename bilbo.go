package main

import (
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"html/template"
	"log"
	"net/http"
	"path"
)

type BilboConfig struct {
	DataDir     string
	HttpListen  string
	Secret      string
	StaticDir   string
	TemplateDir string
}

type Bilbo struct {
	cfg  BilboConfig
	mux  *mux.Router
	repo *git.Repository
	tmpl *template.Template
	hndl http.Handler
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
	b.tmpl = template.Must(template.ParseGlob(path.Join(b.cfg.TemplateDir, "*.html")))

	// Create HTTP routes
	b.mux = mux.NewRouter()
	b.mux.HandleFunc("/", b.HandleIndex).Methods("GET").Name("index")
	b.mux.HandleFunc("/edit/_preview", b.HandleEditPreview).Methods("POST").Name("editPreview")
	b.mux.HandleFunc("/edit/{page:.*}", b.HandleEdit).Methods("GET", "POST").Name("edit")
	b.mux.HandleFunc("/pages/", b.HandlePages).Methods("GET").Name("pagesIndex")
	b.mux.HandleFunc("/pages/{folder:.*}/", b.HandlePages).Methods("GET").Name("pages")
	b.mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(b.cfg.StaticDir))))
	b.mux.HandleFunc("/{page:.+}", b.HandlePage).Methods("GET").Name("page")

	// Apply middlewares
	b.hndl = RecoverMiddleware(b.mux)
	b.hndl = csrf.Protect([]byte(b.cfg.Secret), csrf.Secure(false))(b.hndl)

	log.Printf("Now listening on %s", b.cfg.HttpListen)
	err = http.ListenAndServe(b.cfg.HttpListen, b.hndl)

	return
}

func (b *Bilbo) renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, localData map[string]interface{}) {
	// Global template data
	data := map[string]interface{}{
		"csrfToken":      csrf.Token(r),
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	// Merge the local template data with the global one
	for k, v := range localData {
		data[k] = v
	}

	b.tmpl.ExecuteTemplate(w, templateFile, data)
}
