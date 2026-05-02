package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"saythis-backend/internal/response"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter tracks a token-bucket rate limiter per client IP address.
// It is safe for concurrent use and automatically evicts stale entries to
// prevent unbounded memory growth.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rateVal  rate.Limit
	burst    int
}

// NewRateLimiter creates a RateLimiter.
//   - r   – sustained request rate (use rate.Every(d) or rate.Limit(n))
//   - burst – maximum number of requests allowed in a sudden spike
//
// Example: NewRateLimiter(rate.Every(time.Second), 10) allows 1 req/s
// sustained, with bursts up to 10 requests.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rateVal:  r,
		burst:    burst,
	}
	// Background goroutine: sweep the map every minute and remove IPs that
	// haven't made a request in the last 3 minutes.
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rateVal, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit is the http.Handler middleware. Wrap your router with it via Chain.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Fallback: use the raw RemoteAddr (covers unit-test scenarios)
			ip = r.RemoteAddr
		}

		if !rl.getVisitor(ip).Allow() {
			response.Error(w, http.StatusTooManyRequests, "too many requests — please slow down")
			return
		}

		next.ServeHTTP(w, r)
	})
}
