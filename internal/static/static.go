package static

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"xboard-go/pkg/response"
)

//go:embed dist/*
var distFS embed.FS

var distSub fs.FS

func init() {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to get sub filesystem for embedded dist: " + err.Error())
	}
	distSub = sub
}

// SPAHandler returns a Gin handler that serves the embedded frontend SPA.
// Static assets (JS/CSS/images) are served directly from the embed.FS;
// any unmatched path falls back to index.html so React Router can handle it.
// API paths that don't match a registered route get a JSON 404.
func SPAHandler() gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(distSub))

	return func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")

		// API and health-check paths get a JSON 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/healthz" {
			response.NotFound(c, "resource not found")
			return
		}

		// If the file exists in the embedded FS, serve it directly
		if f, err := distSub.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: rewrite to index.html for client-side routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// Dist returns the embedded frontend filesystem (for advanced use).
func Dist() fs.FS {
	return distSub
}
