package main

import (
	"flag"
	"log"
)

var (
	dataDir     = flag.String("d", "data/", "Path of the GIT repository")
	httpListen  = flag.String("l", "localhost:8080", "[host]:[port] where the HTTP is listening")
	staticDir   = flag.String("s", "static/", "Path to the static files directory")
	templateDir = flag.String("t", "templates/", "Path to the template directory")
)

func main() {
	flag.Parse()

	_, err := NewBilbo(BilboConfig{
		DataDir:     *dataDir,
		HttpListen:  *httpListen,
		StaticDir:   *staticDir,
		TemplateDir: *templateDir,
	})
	if err != nil {
		log.Fatal(err)
	}
}
