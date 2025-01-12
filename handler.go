package vitego

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

func (vGP *ViteGoParams) MuxHandler(ctx context.Context, pageHandler func(http.ResponseWriter, *http.Request), m *http.ServeMux) {
	m.HandleFunc(fmt.Sprintf("/%s", vGP.BasePath), vGP.Handler(ctx, pageHandler))
}

func (vGP *ViteGoParams) Handler(ctx context.Context, pageHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	basePath := fmt.Sprintf("/%s", vGP.BasePath)

	fs := http.FileServer(http.Dir(vGP.DistPath))
	handler := http.StripPrefix(basePath, fs)

	return func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path[len(basePath):]

		// Construct the full path to the file
		filePath := vGP.DistPath + path

		// Check if the file exists
		fileInfo, err := os.Stat(filePath)
		if err == nil && !fileInfo.IsDir() {
			// File exists, serve it
			handler.ServeHTTP(w, r)
			return
		}

		r = r.WithContext(ctx)
		pageHandler(w, r)
	}
}
