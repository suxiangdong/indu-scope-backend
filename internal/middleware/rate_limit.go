package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

type rateLimitBucket struct {
	started time.Time
	count   int
}

var rateLimitState = struct {
	sync.Mutex
	buckets map[string]rateLimitBucket
}{buckets: make(map[string]rateLimitBucket)}

// RateLimit applies a small process-local fixed-window limiter. It is used on
// authentication and expensive verification endpoints as a first line of
// defense against accidental retries and credential/OCR abuse.
func RateLimit(window time.Duration, limit int) ghttp.HandlerFunc {
	if window <= 0 {
		window = time.Minute
	}
	if limit <= 0 {
		limit = 1
	}
	return func(r *ghttp.Request) {
		remote := r.GetRemoteIp()
		if remote == "" {
			remote = r.GetClientIp()
		}
		key := r.Method + ":" + r.URL.Path + ":" + remote
		now := time.Now()

		rateLimitState.Lock()
		for bucketKey, bucket := range rateLimitState.buckets {
			if now.Sub(bucket.started) >= window {
				delete(rateLimitState.buckets, bucketKey)
			}
		}
		if len(rateLimitState.buckets) > 10000 {
			for bucketKey := range rateLimitState.buckets {
				delete(rateLimitState.buckets, bucketKey)
				if len(rateLimitState.buckets) <= 5000 {
					break
				}
			}
		}
		bucket := rateLimitState.buckets[key]
		if bucket.started.IsZero() || now.Sub(bucket.started) >= window {
			bucket = rateLimitBucket{started: now}
		}
		bucket.count++
		rateLimitState.buckets[key] = bucket
		allowed := bucket.count <= limit
		rateLimitState.Unlock()

		if !allowed {
			r.Response.Status = http.StatusTooManyRequests
			r.SetError(gerror.NewCode(gcode.CodeServerBusy, "too many requests, please try again later"))
			return
		}
		r.Middleware.Next()
	}
}
