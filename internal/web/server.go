package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/superciccio/drop-lang/internal/ui"
)

type RouteHandler func(params map[string]string, body interface{}) (interface{}, int, error)

type route struct {
	method  string
	pattern string   // e.g. "/users/:id"
	parts   []string // split path segments
	handler RouteHandler
}

type Server struct {
	port   int
	routes []route
	srv    *http.Server
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

func (s *Server) AddRoute(method, pattern string, handler RouteHandler) {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	s.routes = append(s.routes, route{
		method:  strings.ToUpper(method),
		pattern: pattern,
		parts:   parts,
		handler: handler,
	})
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf(":%d", s.port)

	// Check if port is available before starting
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use — is another Drop app running?", s.port)
	}

	s.srv = &http.Server{Handler: mux}

	// Graceful shutdown on SIGINT/SIGTERM
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-done
		fmt.Println("\nShutting down...")
		s.srv.Shutdown(context.Background())
	}()

	fmt.Printf("Drop server running on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")
	if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	if s.srv != nil {
		return s.srv.Shutdown(context.Background())
	}
	return nil
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Parse the body early so we can check for _method override
	var body interface{}
	var formData url.Values
	if r.Body != nil && (method == "POST" || method == "PUT") {
		contentType := r.Header.Get("Content-Type")
		data, err := io.ReadAll(r.Body)
		if err == nil && len(data) > 0 {
			if strings.Contains(contentType, "application/json") {
				var jsonBody interface{}
				if err := json.Unmarshal(data, &jsonBody); err == nil {
					body = jsonBody
				}
			} else if strings.Contains(contentType, "form") {
				formData, _ = url.ParseQuery(string(data))
				// Method override from forms (for DELETE/PUT via HTML forms)
				if override := formData.Get("_method"); override != "" {
					method = strings.ToUpper(override)
				}
				// Convert form data to a map (excluding _method)
				m := make(map[string]interface{})
				for k, v := range formData {
					if k != "_method" {
						if len(v) == 1 {
							m[k] = v[0]
						} else {
							iface := make([]interface{}, len(v))
							for i, s := range v {
								iface[i] = s
							}
							m[k] = iface
						}
					}
				}
				if len(m) > 0 {
					body = m
				}
			} else {
				// Try JSON, fall back to string
				var jsonBody interface{}
				if err := json.Unmarshal(data, &jsonBody); err == nil {
					body = jsonBody
				} else {
					body = string(data)
				}
			}
		}
	}

	isBrowser := strings.Contains(r.Header.Get("Accept"), "text/html")

	for _, rt := range s.routes {
		if rt.method != method {
			continue
		}
		params, ok := matchRoute(rt.parts, path)
		if !ok {
			continue
		}

		result, status, err := rt.handler(params, body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// For browser form submissions that mutate data, redirect back
		if isBrowser && formData != nil && (method == "POST" || method == "DELETE" || method == "PUT") {
			referer := r.Header.Get("Referer")
			if referer == "" {
				referer = "/"
			}
			http.Redirect(w, r, referer, http.StatusSeeOther)
			return
		}

		// Auto-render for browser requests (skip /api/ paths)
		if isBrowser && !strings.HasPrefix(path, "/api/") {
			writeAutoRendered(w, result, status)
			return
		}

		writeResponse(w, result, status)
		return
	}

	http.NotFound(w, r)
}

func matchRoute(patternParts []string, path string) (map[string]string, bool) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) == 1 && patternParts[0] == "" && len(pathParts) == 1 && pathParts[0] == "" {
		return map[string]string{}, true
	}

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

func writeResponse(w http.ResponseWriter, data interface{}, status int) {
	if status == 0 {
		status = 200
	}
	switch v := data.(type) {
	case string:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprint(w, v)
	default:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(data)
	}
}

func writeAutoRendered(w http.ResponseWriter, data interface{}, status int) {
	if status == 0 {
		status = 200
	}

	// If it's already full HTML (from a page block), serve as-is
	if s, ok := data.(string); ok && strings.HasPrefix(s, "<!DOCTYPE html>") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprint(w, s)
		return
	}

	// Auto-render data to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, ui.AutoRender(data))
}
