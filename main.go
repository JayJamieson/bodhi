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
	rootDir  string
	vm       *sobek.Runtime
	template *template.Template
}

func NewServer(rootDir string) *Server {
	// Load the index.tmpl template
	tmplPath := filepath.Join(rootDir, "index.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}

	return &Server{
		rootDir:  rootDir,
		vm:       sobek.New(),
		template: tmpl,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index"
	}

	// Remove leading slash and add .js extension
	filename := strings.TrimPrefix(path, "/") + ".js"
	filepath := filepath.Join(s.rootDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Read the JavaScript file
	jsContent, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Create a new VM instance for this request
	vm := sobek.New()

	// Execute the JavaScript
	_, jsErr := vm.RunString(string(jsContent))
	if jsErr != nil {
		http.Error(w, fmt.Sprintf("JavaScript error: %v", jsErr), http.StatusInternalServerError)
		return
	}

	var data any

	// Check if loader function exists
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

		// Call loader function as callable
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

	// Get render function
	renderFunc := vm.Get("render")
	if renderFunc == nil || sobek.IsUndefined(renderFunc) {
		http.Error(w, "render function not found", http.StatusInternalServerError)
		return
	}

	// Call render function as callable
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

	// Render the template with the JavaScript render result as content
	templateData := struct {
		Content template.HTML
	}{
		Content: template.HTML(result.String()),
	}

	w.Header().Set("Content-Type", "text/html")
	err = s.template.Execute(w, templateData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	server := NewServer(rootDir)

	log.Printf("Starting server on :8080, serving from %s", rootDir)
	log.Fatal(http.ListenAndServe(":8080", server))
}
