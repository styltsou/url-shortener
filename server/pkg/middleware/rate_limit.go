package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

type RateLimitConfig struct {
	Requests       int
	Window         time.Duration
	TrustedProxies []netip.Prefix
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
			key := clientIP(r, cfg.TrustedProxies)
			// Public redirects must not exhaust the management API's bucket.
			if strings.HasPrefix(r.URL.Path, "/api/v1/") {
				key = "api:" + key
			} else {
				key = "redirect:" + key
			}

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

func clientIP(r *http.Request, trustedProxies []netip.Prefix) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return r.RemoteAddr
	}
	peer = peer.Unmap()
	if !isTrustedProxy(peer, trustedProxies) {
		return peer.String()
	}

	forwarded := strings.Join(r.Header.Values("X-Forwarded-For"), ",")
	if forwarded != "" {
		// Walk from the connection toward the client, stopping at the first
		// untrusted hop. The leftmost value is not automatically trustworthy.
		parts := strings.Split(forwarded, ",")
		current := peer
		for i := len(parts) - 1; i >= 0 && isTrustedProxy(current, trustedProxies); i-- {
			next, parseErr := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if parseErr != nil {
				return peer.String()
			}
			current = next.Unmap()
		}
		return current.String()
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if ip, parseErr := netip.ParseAddr(realIP); parseErr == nil {
			return ip.Unmap().String()
		}
	}
	return peer.String()
}

func isTrustedProxy(ip netip.Addr, proxies []netip.Prefix) bool {
	for _, proxy := range proxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}
