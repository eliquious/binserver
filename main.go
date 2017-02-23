package main

import (
	"context"
	"github.com/eliquious/xrouter"
	"github.com/rs/xlog"
	"github.com/skratchdot/open-golang/open"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var static = http.FileServer(http.Dir("."))

// Handler is the default HTTP handler.
func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logger := xlog.FromContext(ctx)

	// Get file path
	path := strings.Replace(r.URL.Path, "/", "", 1)
	if path == "" {
		path = "index.html"
	}

	// Read asset from memory
	bytes, err := Asset(path)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(r.URL.Path)))
		w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
		w.Write(bytes)
		return
	}
	logger.Error(err)

	// Serve files from .
	static.ServeHTTP(w, r)

	return
}

func main() {
	cfg := xlog.Config{Level: xlog.LevelDebug, Output: xlog.NewConsoleOutput()}
	logger := xlog.New(cfg)

	// Create router and install middleware
	r := xrouter.New()
	r.Use(xlog.NewHandler(cfg))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger := xlog.FromContext(r.Context())
			next.ServeHTTP(w, r)
			logger.Info(xlog.F{
				"duration": time.Now().Sub(start).String(),
			})
		})
	})
	r.Use(xlog.URLHandler("path"))
	r.Use(xlog.MethodHandler("method"))
	r.GET("/", Handler)
	r.StaticRoot(static)

	logger.Info("Serving on port 8080")
	go open.Run("http://localhost:8080")
	if err := http.ListenAndServe(":8080", r.Handler()); err != nil {
		logger.Error(err)
	}
}
