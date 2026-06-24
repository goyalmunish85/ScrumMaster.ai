package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// client represents a single client's rate limiter and last seen time.
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
}

// RateLimiter manages rate limiting for incoming HTTP requests based on IP.
type RateLimiter struct {
	clients map[string]*client
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
}

// NewRateLimiter creates a new RateLimiter middleware.
// r is the rate limit (events per second) and b is the burst size.
func NewRateLimiter(r float64, b int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rate:    rate.Limit(r),
		burst:   b,
	}

	// Start a background goroutine to clean up stale client entries every minute
	go rl.cleanupStaleClients()

	return rl
}

// cleanupStaleClients removes clients that haven't been seen in the last 3 minutes.
func (rl *RateLimiter) cleanupStaleClients() {
	for {
		time.Sleep(time.Minute)

		rl.mu.RLock()
		var staleIPs []string
		for ip, c := range rl.clients {
			c.mu.Lock()
			lastSeen := c.lastSeen
			c.mu.Unlock()
			if time.Since(lastSeen) > 3*time.Minute {
				staleIPs = append(staleIPs, ip)
			}
		}
		rl.mu.RUnlock()

		if len(staleIPs) > 0 {
			rl.mu.Lock()
			for _, ip := range staleIPs {
				delete(rl.clients, ip)
			}
			rl.mu.Unlock()
		}
	}
}

// Handler returns the HTTP middleware function.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("RateLimiter: unable to parse RemoteAddr %s: %v", r.RemoteAddr, err)
			ip = r.RemoteAddr
		}

		rl.mu.RLock()
		c, found := rl.clients[ip]
		rl.mu.RUnlock()

		if !found {
			rl.mu.Lock()
			c, found = rl.clients[ip]
			if !found {
				c = &client{
					limiter:  rate.NewLimiter(rl.rate, rl.burst),
					lastSeen: time.Now(),
				}
				rl.clients[ip] = c
			}
			rl.mu.Unlock()
		}

		c.mu.Lock()
		c.lastSeen = time.Now()
		c.mu.Unlock()

		if !c.limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
