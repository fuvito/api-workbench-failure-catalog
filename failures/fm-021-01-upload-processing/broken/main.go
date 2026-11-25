package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// BUG: the file record is saved before the file itself, and a storage failure is
// ignored. Empty files are accepted and the client's checksum is never checked.

type record struct {
	ID     string `json:"id"`
	Size   int    `json:"size"`
	Status string `json:"status"`
}

var (
	mu      sync.Mutex
	records = map[string]record{}
	objects = map[string][]byte{} // stands in for object storage
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "file could not be read")
		return
	}
	// BUG: no empty-file check, and X-Content-SHA256 is ignored

	mu.Lock()
	defer mu.Unlock()
	// BUG: reports success before the file is stored
	rec := record{ID: fmt.Sprintf("file-%d", len(records)+1), Size: len(data), Status: "stored"}
	records[rec.ID] = rec

	// Test hook: the object storage write "fails".
	if r.Header.Get("X-Test-Fault") != "storage-write" {
		objects[rec.ID] = data
	}
	writeJSON(w, http.StatusCreated, rec)
}

func list(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]int{"count": len(records)})
}

func content(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	data, ok := objects[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "content not found")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /files", upload)
	mux.HandleFunc("GET /files", list)
	mux.HandleFunc("GET /files/{id}/content", content)

	http.ListenAndServe(":19071", mux)
}
