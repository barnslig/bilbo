package main

import (
	"net/http"
)

func (b *Bilbo) HandleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/Home", http.StatusSeeOther)
}
