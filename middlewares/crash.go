package middlewares

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

//RecoverWrap to send log about the crash
func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				var err error
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				crash, ctx := opentracing.StartSpanFromContext(req.Context(), "crash")
				req = req.WithContext(ctx)
				defer crash.Finish()
				crash.LogFields(log.String("crash", err.Error()), log.String("stack", string(debug.Stack())))
				crash.SetTag("error", true)
				crash.SetTag("crash", true)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, req)
	})
}
