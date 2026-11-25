package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// The file is checked, then stored, and only then recorded and reported as
// stored. A storage failure returns 502 and leaves no record.

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
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "file is empty")
		return
	}
	if want := r.Header.Get("X-Content-SHA256"); want != "" {
		sum := sha256.Sum256(data)
		if !strings.EqualFold(want, hex.EncodeToString(sum[:])) {
			writeError(w, http.StatusUnprocessableEntity, "checksum does not match")
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()
	rec := record{ID: fmt.Sprintf("file-%d", len(records)+1), Size: len(data), Status: "stored"}

	// Test hook: the object storage write "fails".
	if r.Header.Get("X-Test-Fault") == "storage-write" {
		writeError(w, http.StatusBadGateway, "file could not be stored")
		return
	}
	objects[rec.ID] = data
	records[rec.ID] = rec // record it only once the file is stored
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
