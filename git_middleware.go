package main

import (
	"context"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/http"
)

// This middleware adds the to-be-used GIT HEAD to the context.
// Reasons:
// - Performance: Only call expensive .Head() method once per request
// - Commit context: Only one place to set the currently viewed point in GIT history
func (b *Bilbo) GitMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHead := false
		commitHash := plumbing.NewHash(r.URL.Query().Get("commit"))

		if commitHash == plumbing.ZeroHash {
			head, err := b.repo.Head()
			if err != nil {
				return
			}

			isHead = true
			commitHash = head.Hash()
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "GitHead", commitHash)
		ctx = context.WithValue(ctx, "GitIsHead", isHead)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
