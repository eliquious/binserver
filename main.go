package main

import (
	"flag"
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

var debug = flag.Bool("debug", false, "Enables debug logging")

// StaticRoot serves the files in the current directory
var StaticRoot = http.FileServer(http.Dir("."))

// Handler is the default HTTP handler.
func Handler(w http.ResponseWriter, r *http.Request) {
	logger := xlog.FromRequest(r)

	// Get file path
	path := strings.Replace(r.URL.Path, "/", "", 1)
	if path == "" {
		path = "index.html"
	}

	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(path, "js") {
		w.Header().Set("Content-Type", "application/x-javascript")
	} else {
		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	}

	// Read asset from memory
	bytes, err := Asset(path)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
		w.Write(bytes)
		return
	}
	logger.Debug("Loading from disk...")

	// Serve files from .
	StaticRoot.ServeHTTP(w, r)

	return
}

func main() {
	flag.Parse()

	level := xlog.LevelError
	if *debug {
		level = xlog.LevelInfo
	}

	cfg := xlog.Config{Level: level, Output: xlog.NewConsoleOutput()}
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
	r.NotFound(http.HandlerFunc(Handler))

	logger.Info("Serving on port 8080")
	go open.Run("http://localhost:8080")
	if err := http.ListenAndServe(":8080", r.Handler()); err != nil {
		logger.Error(err)
	}
}
