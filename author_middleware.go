package main

import (
	"context"
	"net/http"
)

type GitAuthor struct {
	Name  string
	Email string
}

func (b *Bilbo) AuthorMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		author := GitAuthor{
			Name:  "Anonymous",
			Email: "anon@anon.com",
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, key("GitAuthor"), author)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
