// Package server serves the rendered report on a local http port.
package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
	"github.com/seanmcn/whatsapp-analyse/internal/parser"
	"github.com/seanmcn/whatsapp-analyse/internal/render"
)

const (
	maxUploadBytes = 50 << 20 // 50 MiB
	sessionTTL     = time.Hour
)

type session struct {
	msgs    []parser.Message
	authors []string // distinct authors, most-frequent first
	created time.Time
}

type sessionStore struct {
	mu sync.Mutex
	m  map[string]*session
}

func newSessionStore() *sessionStore { return &sessionStore{m: map[string]*session{}} }

func (s *sessionStore) put(sess *session) string {
	var b [16]byte
	rand.Read(b[:])
	id := hex.EncodeToString(b[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	// opportunistic cleanup
	cutoff := time.Now().Add(-sessionTTL)
	for k, v := range s.m {
		if v.created.Before(cutoff) {
			delete(s.m, k)
		}
	}
	s.m[id] = sess
	return id
}

func (s *sessionStore) get(id string) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[id]
}

func distinctAuthors(msgs []parser.Message) []string {
	counts := map[string]int{}
	for _, m := range msgs {
		counts[m.Author]++
	}
	type ac struct {
		name string
		n    int
	}
	list := make([]ac, 0, len(counts))
	for k, v := range counts {
		list = append(list, ac{k, v})
	}
	// simple sort
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].n > list[j-1].n; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	out := make([]string, len(list))
	for i, a := range list {
		out[i] = a.name
	}
	return out
}

func Serve(addr string, s analyse.Stats) error {
	html, err := render.HTML(s)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(render.StaticFS()))))
	fmt.Printf("Serving on http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// ServeUpload runs the server in upload mode: the user provides a chat
// export via a web form and the report is rendered on demand.
func ServeUpload(addr string, gap time.Duration) error {
	store := newSessionStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(uploadFormHTML))
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			http.Error(w, "upload too large or malformed: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("chat")
		if err != nil {
			http.Error(w, "missing 'chat' file field", http.StatusBadRequest)
			return
		}
		defer file.Close()
		buf, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "read upload: "+err.Error(), http.StatusBadRequest)
			return
		}
		var msgs []parser.Message
		if strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
			msgs, err = parser.ParseZipReader(bytes.NewReader(buf), int64(len(buf)))
		} else {
			msgs, err = parser.Parse(bytes.NewReader(buf))
		}
		if err != nil {
			http.Error(w, "parse: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(msgs) == 0 {
			http.Error(w, "no messages parsed — check the file format", http.StatusBadRequest)
			return
		}
		id := store.put(&session{msgs: msgs, authors: distinctAuthors(msgs), created: time.Now()})
		http.Redirect(w, r, "/report?id="+id, http.StatusSeeOther)
	})
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		sess := store.get(id)
		if sess == nil {
			http.Error(w, "session not found or expired — please re-upload", http.StatusNotFound)
			return
		}
		me := r.URL.Query().Get("me")
		them := r.URL.Query().Get("them")
		stats := analyse.Run(sess.msgs, me, them, gap)
		htmlBytes, err := render.HTML(stats)
		if err != nil {
			http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
			return
		}
		htmlBytes = injectPerspectiveBar(htmlBytes, id, sess.authors, stats.Participants[0], stats.Participants[1])
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlBytes)
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(render.StaticFS()))))
	fmt.Printf("Serving upload form on http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func injectPerspectiveBar(doc []byte, id string, authors []string, me, them string) []byte {
	swapURL := "/report?id=" + url.QueryEscape(id) + "&me=" + url.QueryEscape(them) + "&them=" + url.QueryEscape(me)
	var b strings.Builder
	b.WriteString(`<div style="background:#161b22;border:1px solid #30363d;border-radius:14px;padding:.75rem 1.25rem;margin:0 0 18px 0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#e6edf3;display:flex;gap:1.25rem;align-items:center;flex-wrap:wrap;font-size:.9rem;">`)
	b.WriteString(`<span style="color:#8b949e;">Showing <b style="color:#e6edf3;font-weight:600;">` + html.EscapeString(me) + `</b> → <b style="color:#e6edf3;font-weight:600;">` + html.EscapeString(them) + `</b></span>`)
	b.WriteString(`<a href="` + swapURL + `" style="background:#238636;color:#fff;text-decoration:none;padding:.4rem .9rem;border-radius:6px;font-weight:500;">⇄ Swap perspective</a>`)
	b.WriteString(`<a href="/" style="color:#8b949e;text-decoration:none;margin-left:auto;">Upload another</a>`)
	b.WriteString(`</div>`)
	_ = authors
	bar := []byte(b.String())
	if i := bytes.Index(doc, []byte("<body>")); i >= 0 {
		out := make([]byte, 0, len(doc)+len(bar))
		out = append(out, doc[:i+len("<body>")]...)
		out = append(out, bar...)
		out = append(out, doc[i+len("<body>"):]...)
		return out
	}
	return append(bar, doc...)
}

const uploadFormHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Chinwag — WhatsApp chat analyser</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background:#0e1116; color:#e6edf3; display:flex; align-items:center; justify-content:center; min-height:100vh; margin:0; }
  .card { background:#161b22; border:1px solid #30363d; border-radius:12px; padding:2.5rem; max-width:480px; width:90%; box-shadow:0 8px 32px rgba(0,0,0,0.4); }
  h1 { margin:0 0 .5rem; font-size:1.6rem; }
  p { color:#8b949e; line-height:1.5; }
  input[type=file] { display:block; width:100%; margin:1.25rem 0; padding:.75rem; background:#0d1117; border:1px dashed #30363d; border-radius:8px; color:#e6edf3; }
  button { background:#238636; color:#fff; border:0; padding:.75rem 1.25rem; border-radius:8px; font-size:1rem; cursor:pointer; width:100%; }
  button:hover { background:#2ea043; }
  small { color:#6e7681; }
</style>
</head>
<body>
<div class="card">
  <h1>Chinwag</h1>
  <p>Upload a WhatsApp chat export (<code>.txt</code> or <code>.zip</code>) to see your stats. Files are processed in memory and never stored.</p>
  <form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="chat" accept=".txt,.zip" required>
    <button type="submit">Analyse</button>
  </form>
  <p><small>To export: open a chat in WhatsApp → menu → Export chat → Without media.</small></p>
</div>
</body>
</html>`
