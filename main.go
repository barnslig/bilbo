package main

import (
	"flag"
	"log"
)

var (
	dataDir    = flag.String("d", "data/", "Path of the GIT repository")
	httpListen = flag.String("l", "localhost:8080", "[host]:[port] where the HTTP is listening")
	isDev      = flag.Bool("dev", false, "Run in development mode")
	secret     = flag.String("secret", "changeme", "32 byte long CSRF token secret")
)

func main() {
	flag.Parse()

	_, err := NewBilbo(BilboConfig{
		DataDir:    *dataDir,
		HttpListen: *httpListen,
		IsDev:      *isDev,
		Secret:     *secret,
	})
	if err != nil {
		log.Fatal(err)
	}
}
