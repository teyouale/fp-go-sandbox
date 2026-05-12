// Package main implements a lightweight HTTP server that is API-compatible
// with Codapi's /v1/exec endpoint. Instead of spawning Docker containers,
// it runs Go code directly on the host using `go run`.
//
// This is designed to run on Fly.io where Docker-in-Docker is not available.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// --- Codapi-compatible request/response types ---

type ExecRequest struct {
	Sandbox string            `json:"sandbox"`
	Command string            `json:"command"`
	Files   map[string]string `json:"files"`
}

type ExecResponse struct {
	ID       string `json:"id"`
	OK       bool   `json:"ok"`
	Duration int    `json:"duration"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// --- Configuration ---

const (
	maxTimeout    = 10 * time.Second
	maxOutputSize = 8192
	maxPoolSize   = 4
	playgroundDir = "/home/sandbox/playground"
	samplesDir    = "/home/sandbox/samples"
	listenAddr    = ":1313"
)

// semaphore limits concurrent executions
var sem = make(chan struct{}, maxPoolSize)

// counter for unique request IDs
var (
	counterMu sync.Mutex
	counter   int64
)

func nextID(sandbox, command string) string {
	counterMu.Lock()
	counter++
	id := counter
	counterMu.Unlock()
	return fmt.Sprintf("%s_%s_%d", sandbox, command, id)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/exec", handleExec)
	mux.HandleFunc("/v1/samples", handleListSamples)
	mux.HandleFunc("/v1/samples/", handleRunSample)
	mux.HandleFunc("/health", handleHealth)

	// CORS middleware for browser-based playgrounds
	handler := corsMiddleware(mux)

	log.Printf("🚀 fp-go sandbox runner listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

// --- Sample endpoints ---

type SampleInfo struct {
	Name string `json:"name"`
	URL  string `json:"run_url"`
}

func handleListSamples(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(samplesDir)
	if err != nil {
		writeJSON(w, http.StatusOK, []SampleInfo{})
		return
	}

	var samples []SampleInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".go")
		samples = append(samples, SampleInfo{
			Name: name,
			URL:  fmt.Sprintf("/v1/samples/%s", name),
		})
	}
	writeJSON(w, http.StatusOK, samples)
}

func handleRunSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract sample name from URL: /v1/samples/{name}
	name := strings.TrimPrefix(r.URL.Path, "/v1/samples/")
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		handleListSamples(w, r)
		return
	}

	// Read the sample file
	filePath := filepath.Join(samplesDir, name+".go")
	code, err := os.ReadFile(filePath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ExecResponse{
			OK:     false,
			Stderr: fmt.Sprintf("sample not found: %s", name),
		})
		return
	}

	// Acquire a slot
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	default:
		writeJSON(w, http.StatusServiceUnavailable, ExecResponse{
			OK: false, Stderr: "busy: try again later",
		})
		return
	}

	id := nextID("sample", name)
	resp := executeGo(id, string(code))
	writeJSON(w, http.StatusOK, resp)
}

func handleExec(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Sandbox != "go" {
		resp := ExecResponse{
			OK:     false,
			Stderr: fmt.Sprintf("unknown sandbox: %s", req.Sandbox),
		}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	// Extract code from files
	code := extractCode(req.Files)
	if strings.TrimSpace(code) == "" {
		resp := ExecResponse{OK: false, Stderr: "empty request"}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	id := nextID(req.Sandbox, req.Command)

	// Try to acquire a slot (non-blocking)
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	default:
		resp := ExecResponse{ID: id, OK: false, Stderr: "busy: try again later"}
		writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}

	// Execute the code
	resp := executeGo(id, code)
	writeJSON(w, http.StatusOK, resp)
}

func extractCode(files map[string]string) string {
	// Codapi sends code with key "" (empty string) for single-file submissions
	if code, ok := files[""]; ok {
		return code
	}
	// Or with key "main.go"
	if code, ok := files["main.go"]; ok {
		return code
	}
	// Fall back to first file
	for _, code := range files {
		return code
	}
	return ""
}

func executeGo(id, code string) ExecResponse {
	start := time.Now()

	// Create a temp directory for this execution
	tmpDir, err := os.MkdirTemp("", "codapi-*")
	if err != nil {
		return ExecResponse{
			ID: id, OK: false,
			Stderr:   "internal error: failed to create temp dir",
			Duration: int(time.Since(start).Milliseconds()),
		}
	}
	defer os.RemoveAll(tmpDir)

	// Write the user's code
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(code), 0644); err != nil {
		return ExecResponse{
			ID: id, OK: false,
			Stderr:   "internal error: failed to write code",
			Duration: int(time.Since(start).Milliseconds()),
		}
	}

	// Copy go.mod and go.sum from the pre-warmed playground
	copyFile(filepath.Join(playgroundDir, "go.mod"), filepath.Join(tmpDir, "go.mod"))
	copyFile(filepath.Join(playgroundDir, "go.sum"), filepath.Join(tmpDir, "go.sum"))

	// Run the code with timeout
	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", mainFile)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"GOPROXY=off",
		"GOFLAGS=-mod=mod",
		"CGO_ENABLED=0",
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &limitedWriter{w: &stdout, limit: maxOutputSize}
	cmd.Stderr = &limitedWriter{w: &stderr, limit: maxOutputSize}

	err = cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	stdoutStr := strings.TrimSpace(stdout.String())
	stderrStr := strings.TrimSpace(stderr.String())

	if ctx.Err() == context.DeadlineExceeded {
		return ExecResponse{
			ID: id, OK: false,
			Stdout:   stdoutStr,
			Stderr:   "code execution timeout",
			Duration: duration,
		}
	}

	if err != nil {
		// Compilation or runtime error — this is a user code problem, not ours
		combined := stdoutStr
		if stderrStr != "" {
			if combined != "" {
				combined = combined + "\n" + stderrStr
			} else {
				combined = stderrStr
			}
		}
		return ExecResponse{
			ID: id, OK: false,
			Stderr:   combined,
			Duration: duration,
		}
	}

	return ExecResponse{
		ID: id, OK: true,
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		Duration: duration,
	}
}

// copyFile copies src to dst. Errors are silently ignored.
func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	_ = os.WriteFile(dst, data, 0644)
}

// limitedWriter limits the number of bytes written.
type limitedWriter struct {
	w       *strings.Builder
	limit   int
	written int
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.limit - lw.written
	if remaining <= 0 {
		return len(p), nil // silently discard
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	n, err := lw.w.Write(p)
	lw.written += n
	return n, err
}

// corsMiddleware adds CORS headers for browser-based playgrounds.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
