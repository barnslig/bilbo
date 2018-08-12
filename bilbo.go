package main

import (
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
	b.mux.HandleFunc("/pages/", b.HandlePages).Methods("GET").Name("pagesIndex")
	b.mux.HandleFunc("/pages/{folder:.*}/", b.HandlePages).Methods("GET").Name("pages")
	b.mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(b.cfg.StaticDir))))
	b.mux.HandleFunc("/{page:.+}", b.HandlePage).Methods("GET").Name("page")

	// Apply middlewares
	b.hndl = RecoverMiddleware(b.mux)

	log.Printf("Now listening on %s", b.cfg.HttpListen)
	err = http.ListenAndServe(b.cfg.HttpListen, b.hndl)

	return
}

func (b *Bilbo) renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, data map[string]interface{}) {
	b.tmpl.ExecuteTemplate(w, templateFile, data)
}
