package main

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

// This middleware adds the to-be-used GIT HEAD to the context.
// Reasons:
// - Performance: Only call expensive .Head() method once per request
// - Commit context: Only one place to set the currently viewed point in GIT history
func (b *Bilbo) GitMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHead := false
		commitHash := plumbing.NewHash(r.URL.Query().Get("commit"))

		head, err := b.repo.Head()
		if err != nil {
			return
		}

		if commitHash == plumbing.ZeroHash {
			commitHash = head.Hash()
		}

		if commitHash == head.Hash() {
			isHead = true
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, key("GitHead"), commitHash)
		ctx = context.WithValue(ctx, key("GitIsHead"), isHead)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
