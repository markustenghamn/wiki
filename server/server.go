package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/markustenghamn/wiki/db"
	"github.com/russross/blackfriday"
)

// Server is the wiki server
type Server struct {
	logger   *log.Logger
	db       *db.DB
	username string
	password string
}

// New creates a new wiki server
func New(logger *log.Logger, db *db.DB) *Server {
	fmt.Println(os.Getenv("WIKI_USERNAME"))
	fmt.Println(os.Getenv("WIKI_PASSWORD"))
	return &Server{
		logger:   logger,
		db:       db,
		username: os.Getenv("WIKI_USERNAME"),
		password: os.Getenv("WIKI_PASSWORD")}
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	if path := r.URL.Path; len(r.URL.Path) > 1 {
		target := strings.TrimSuffix(path, "/")

		if target == "/home" {
			target = "/"
		}

		http.Redirect(w, r, target, 302)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "wiki")

	s.logger.Println(r.Method, r.URL.String())

	switch {
	case r.Method == http.MethodPost:
		if s.basicAuth(w, r) {
			s.save(w, r)
		}
	case r.URL.Path == "/favicon.ico":
		s.favicon(w, r)
	case r.URL.Path == "/home":
		s.redirect(w, r)
	case strings.HasSuffix(r.URL.Path, "/edit"):
		if s.basicAuth(w, r) {
			s.edit(w, r)
		}
	case strings.HasSuffix(r.URL.Path, "/") && len(r.URL.Path) > 1:
		s.redirect(w, r)
	default:
		s.show(w, r)
	}
}

func (s *Server) getPageName(r *http.Request) []byte {
	name := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/edit"), "/")

	if name == "" {
		return []byte("home")
	}

	return []byte(name)
}

type data map[string]interface{}

// bytesAsHTML returns the template bytes as HTML
func bytesAsHTML(b []byte) template.HTML {
	return template.HTML(string(b))
}

// parsedMarkdown returns provided bytes parsed as Markdown
func parsedMarkdown(b []byte) []byte {
	return blackfriday.MarkdownCommon(b)
}

func (s *Server) basicAuth(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	username, password, authOK := r.BasicAuth()
	if authOK == false {
		http.Error(w, "Not authorized", 401)
		return false
	}
	fmt.Println(username, "server", s.username, password, "serverp", s.password)
	if username != s.username || password != s.password {
		http.Error(w, "Not authorized", 401)
		return false
	}
	return true
}
