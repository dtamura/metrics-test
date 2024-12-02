package main

import (
	"encoding/json"
	"net/http"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
