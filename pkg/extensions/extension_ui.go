//go:build search && ui
// +build search,ui

package extensions

import (
	"embed"
	"os"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"zotregistry.dev/zot/pkg/api/config"
	zcommon "zotregistry.dev/zot/pkg/common"
	"zotregistry.dev/zot/pkg/log"
)

// content is our static web server content.
//
//go:embed build/*
var content embed.FS

type uiHandler struct {
	log log.Logger
	fs fs.FS
}

func (uih uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf, _ := fs.ReadFile(uih.fs, "index.html")

	_, err := w.Write(buf)
	if err != nil {
		uih.log.Error().Err(err).Msg("failed to serve index.html")
	}
}

func addUISecurityHeaders(h http.Handler) http.HandlerFunc { //nolint:varnamelen
	return func(w http.ResponseWriter, r *http.Request) {
		permissionsPolicy := "microphone=(), geolocation=(), battery=(), camera=(), autoplay=(), gyroscope=(), payment=()"
		w.Header().Set("Permissions-Policy", permissionsPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		cspDirectives := []string{
			"default-src 'none'",
			"script-src 'self' 'unsafe-inline'",
			"style-src 'self' 'unsafe-inline'",
			"font-src 'self'",
			"connect-src 'self'",
			"img-src 'self'",
			"manifest-src 'self'",
			"base-uri 'self'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(cspDirectives, "; "))

		h.ServeHTTP(w, r)
	}
}

func SetupUIRoutes(conf *config.Config, router *mux.Router,
	log log.Logger,
) {
	if !conf.IsUIEnabled() {
		log.Info().Msg("skip enabling the ui route as the config prerequisites are not met")

		return
	}

	log.Info().Msg("setting up ui routes")

	var fsys fs.FS
	fsys, _ = fs.Sub(content, "build")

	uih := uiHandler{log: log, fs: fsys}

	contentPath := conf.Extensions.UI.ContentPath
	if contentPath != "" {
		log.Info().Msg("serving local build for the ui")

		fsys = os.DirFS(contentPath)
		uih.fs = fsys
	}

	// See https://go-review.googlesource.com/c/go/+/482635/2/src/net/http/fs.go
	// See https://github.com/golang/go/issues/59469
	// In go 1.20.4 they decided to allow any method in the FileServer handler.
	// In order to be consistent with the status codes returned when the UI is disabled
	// we need to be explicit about the methods we allow on UI routes.
	// If we don't add this, all unmatched http methods on any urls would match the UI routes.
	allowedMethods := zcommon.AllowedMethods(http.MethodGet)

	router.PathPrefix("/login").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(uih))
	router.PathPrefix("/home").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(uih))
	router.PathPrefix("/explore").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(uih))
	router.PathPrefix("/image").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(uih))
	router.PathPrefix("/user").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(uih))
	router.PathPrefix("/").Methods(allowedMethods...).
		Handler(addUISecurityHeaders(http.FileServer(http.FS(fsys))))

	log.Info().Msg("finished setting up ui routes")
}
