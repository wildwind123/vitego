package vitego

import (
	"context"
	"net/http"
	"os"
)

func (vGP *ViteGoParams) MuxHandler(ctx context.Context, pageHandler func(http.ResponseWriter, *http.Request), m *http.ServeMux) {
	m.HandleFunc(vGP.BasePath, vGP.Handler(ctx, pageHandler))
}

func (vGP *ViteGoParams) Handler(ctx context.Context, pageHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	fs := http.FileServer(http.Dir(vGP.DistPath))
	handler := http.StripPrefix(vGP.BasePath, fs)

	return func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path[len(vGP.BasePath):]

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
