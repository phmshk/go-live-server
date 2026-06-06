package server

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phmshk/go-live-server/internal/utils"
)

type customResponseWriter struct {
	originalRW http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (crw *customResponseWriter) Header() http.Header {
	return crw.originalRW.Header()
}

func (crw *customResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if crw.statusCode == 0 {
		crw.statusCode = http.StatusOK
	}

	return crw.buffer.Write(b)
}

func LiveReloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/live-reload" {
			next.ServeHTTP(w, r)
			return
		}

		fileExt := filepath.Ext(r.URL.Path)

		if fileExt != "" && fileExt != ".html" && fileExt != ".htm" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		buffer := &bytes.Buffer{}
		customRW := &customResponseWriter{
			originalRW: w,
			buffer:     buffer,
			statusCode: 0,
		}

		next.ServeHTTP(customRW, r)

		if customRW.statusCode != http.StatusOK {
			w.WriteHeader(customRW.statusCode)
			w.Write(buffer.Bytes())
			return
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			w.WriteHeader(customRW.statusCode)
			w.Write(buffer.Bytes())
			return
		}

		rawBytes := buffer.Bytes()
		tagIndex := utils.CheckTags(rawBytes)
		if tagIndex == -1 {
			w.WriteHeader(customRW.statusCode)
			w.Write(rawBytes)
			return
		}

		scriptBytes := []byte(liveReloadScript)
		finalBytes := make([]byte, 0, len(rawBytes)+len(scriptBytes))

		finalBytes = append(finalBytes, rawBytes[:tagIndex]...)
		finalBytes = append(finalBytes, scriptBytes...)
		finalBytes = append(finalBytes, rawBytes[tagIndex:]...)

		w.Header().Set("Content-Length", strconv.Itoa(len(finalBytes)))
		w.WriteHeader(customRW.statusCode)
		w.Write(finalBytes)
	})
}
