package webserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"opskit/internal/state"
)

func Serve(paths state.Paths, listenAddr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(paths.UIDir))))
	mux.Handle("/state/", http.StripPrefix("/state/", cacheControl(http.FileServer(http.Dir(paths.StateDir)))))
	mux.Handle("/reports/", http.StripPrefix("/reports/", http.FileServer(http.Dir(paths.ReportsDir))))
	mux.Handle("/evidence/", http.StripPrefix("/evidence/", http.FileServer(http.Dir(paths.EvidenceDir))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	if err := ensureIndex(paths.UIDir); err != nil {
		return err
	}
	if listenAddr == "" {
		listenAddr = ":18080"
	}
	fmt.Printf("opskit web serving %s at http://127.0.0.1%s/ui/\n", paths.Root, listenAddr)
	return http.ListenAndServe(listenAddr, mux)
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func ensureIndex(uiDir string) error {
	p := filepath.Join(uiDir, "index.html")
	if _, err := os.Stat(p); err != nil {
		return fmt.Errorf("ui missing: %s", p)
	}
	return nil
}
