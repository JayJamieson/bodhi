package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/sobek"
)

type Request struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type Server struct {
	rootDir string
	vm      *sobek.Runtime
}

func NewServer(rootDir string) *Server {
	return &Server{
		rootDir: rootDir,
		vm:      sobek.New(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index"
	}

	filename := strings.TrimPrefix(path, "/") + ".js"
	pagePath := filepath.Join(s.rootDir, filename)

	if _, err := os.Stat(pagePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	jsContent, err := os.ReadFile(pagePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// TODO: create a pool of VMs for each page file or figure out a way to clear the vm
	vm := sobek.New()

	_, jsErr := vm.RunString(string(jsContent))
	if jsErr != nil {
		http.Error(w, fmt.Sprintf("JavaScript error: %v", jsErr), http.StatusInternalServerError)
		return
	}

	var data any

	// handle optional loader function, skip if not defined
	loaderFunc := vm.Get("loader")
	if loaderFunc != nil && !sobek.IsUndefined(loaderFunc) {
		// Read request body
		body, _ := io.ReadAll(r.Body)
		req := Request{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header,
			Body:    string(body),
		}

		callable, ok := sobek.AssertFunction(loaderFunc)
		if !ok {
			http.Error(w, "loader is not a function", http.StatusInternalServerError)
			return
		}

		result, loaderErr := callable(sobek.Undefined(), vm.ToValue(req))
		if loaderErr != nil {
			http.Error(w, fmt.Sprintf("Loader error: %v", loaderErr), http.StatusInternalServerError)
			return
		}
		data = result.Export()
	}

	renderFunc := vm.Get("render")
	if renderFunc == nil || sobek.IsUndefined(renderFunc) {
		http.Error(w, "render function not found", http.StatusInternalServerError)
		return
	}

	renderCallable, ok := sobek.AssertFunction(renderFunc)
	if !ok {
		http.Error(w, "render is not a function", http.StatusInternalServerError)
		return
	}

	var result sobek.Value
	var renderErr error

	if data != nil {
		result, renderErr = renderCallable(sobek.Undefined(), vm.ToValue(data))
	} else {
		result, renderErr = renderCallable(sobek.Undefined())
	}

	if renderErr != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", renderErr), http.StatusInternalServerError)
		return
	}

	infos := listPages(s)

	templateData := make(map[string]any)
	templateData["pages"] = infos

	tmpl, err := template.ParseFiles(filepath.Join(s.rootDir, "index.tmpl"))

	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	child := fmt.Sprintf(`{{define "content"}}%s{{end}}`, result.String())

	_, err = tmpl.Parse(child)

	if err != nil {
		http.Error(w, fmt.Sprintf("Page template error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, templateData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
}

// lists available pages in configured rootDir filtering out index.tmpl, index.js.
// The pages can then be used from index.tmpl for creating navigation links
func listPages(s *Server) []string {
	entries, _ := os.ReadDir(s.rootDir)
	infos := make([]string, 0, len(entries))

	for _, entry := range entries {
		info, _ := entry.Info()
		name := info.Name()
		if name == "index.tmpl" || name == "index.js" {
			continue
		}

		infos = append(infos, strings.TrimSuffix(name, ".js"))
	}
	return infos
}

func main() {
	rootDir := "./pages"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	server := NewServer(rootDir)

	log.Printf("Starting server on :8080, serving from %s", rootDir)
	log.Fatal(http.ListenAndServe(":8080", server))
}
