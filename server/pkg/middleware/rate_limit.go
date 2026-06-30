package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type rateLimitBucket struct {
	windowStart time.Time
	count       int
}

func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Requests <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}

	var mu sync.Mutex
	buckets := make(map[string]rateLimitBucket)
	lastCleanup := time.Now()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			key := clientIP(r)

			mu.Lock()
			if now.Sub(lastCleanup) > cfg.Window {
				for ip, bucket := range buckets {
					if now.Sub(bucket.windowStart) > cfg.Window {
						delete(buckets, ip)
					}
				}
				lastCleanup = now
			}

			bucket := buckets[key]
			if bucket.windowStart.IsZero() || now.Sub(bucket.windowStart) >= cfg.Window {
				bucket = rateLimitBucket{windowStart: now}
			}
			bucket.count++
			buckets[key] = bucket

			remaining := cfg.Requests - bucket.count
			resetAfter := cfg.Window - now.Sub(bucket.windowStart)
			limited := bucket.count > cfg.Requests
			mu.Unlock()

			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(now.Add(resetAfter).Unix(), 10))

			if limited {
				w.Header().Set("Retry-After", strconv.Itoa(int(resetAfter.Seconds())+1))
				render.Status(r, http.StatusTooManyRequests)
				render.JSON(w, r, dto.ErrorResponse{
					Error: dto.ErrorObject{
						Code:   apperrors.CodeRateLimited,
						Title:  "Rate limit exceeded",
						Detail: "Too many requests. Please retry after the rate limit window resets.",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
