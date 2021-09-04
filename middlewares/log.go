package middlewares

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func JaeggerLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//tracer := opentracing.GlobalTracer()
		//clientSpan := tracer.StartSpan(r.URL.Path)
		clientSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "log")

		err := clientSpan.Tracer().Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(w.Header()))
		if err != nil {
			clientSpan.LogFields(log.String("error", err.Error()))
		}
		defer clientSpan.Finish()

		for k, v := range r.Header {
			clientSpan.SetTag(k, v)
		}
		clientSpan.
			SetTag("Remote", r.RemoteAddr).
			SetTag("Method", r.Method).
			SetTag("Path", r.URL.RawPath)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
