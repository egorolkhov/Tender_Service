package middleware

import (
	"avito.go/pkg/logger"
	"compress/gzip"
	"errors"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
	"time"
)

func Middleware(h http.HandlerFunc) http.HandlerFunc {
	foo := func(w http.ResponseWriter, r *http.Request) {
		Time := time.Now()
		Duration := time.Since(Time)
		logger.Log.Info(
			"INFO",
			zap.String("method", r.Method),
			zap.String("time", Duration.String()),
			zap.String("URI", r.RequestURI),
		)

		cookie, err := r.Cookie("default")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				value, err := BuildToken(secretKey)
				if err != nil {
					log.Println("error building token", err)
				}
				cookie = &http.Cookie{
					Name:  "default",
					Value: value,
				}
				http.SetCookie(w, cookie)
			default:
				log.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
		w.Header().Set("Authorization", cookie.Value)

		if !strings.Contains(r.Header.Get("Accept-Encoding"), `gzip`) {
			h.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer gz.Close()
		cw := &CompressWrite{w, gz}
		cw.Header().Set("Content-Encoding", "gzip")

		if !strings.Contains(r.Header.Get("Content-Encoding"), `gzip`) {
			h.ServeHTTP(cw, r)
			return
		}

		gzR, err := gzip.NewReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cr := &CompressRead{r.Body, gzR}

		r.Body = cr

		h.ServeHTTP(cw, r)
	}
	return foo
}
