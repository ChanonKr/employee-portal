package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"employee-portal/pkg/httputil"
)

// KeyedLimiter keeps one token bucket per key (IP or user). Idle buckets are
// swept so memory does not grow without bound.
type KeyedLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	limit     rate.Limit
	burst     int
	idleTTL   time.Duration
	lastSweep time.Time
}

type bucket struct {
	lim  *rate.Limiter
	seen time.Time
}

// NewKeyedLimiter allows `n` requests per `per` for each key.
func NewKeyedLimiter(n int, per time.Duration) *KeyedLimiter {
	return &KeyedLimiter{
		buckets: make(map[string]*bucket),
		limit:   rate.Every(per / time.Duration(n)),
		burst:   n,
		idleTTL: 2 * per,
	}
}

// Allow reports whether the request may proceed and, if not, how long to wait.
func (k *KeyedLimiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()
	k.mu.Lock()
	defer k.mu.Unlock()

	if now.Sub(k.lastSweep) > k.idleTTL {
		for key, b := range k.buckets {
			if now.Sub(b.seen) > k.idleTTL {
				delete(k.buckets, key)
			}
		}
		k.lastSweep = now
	}

	b, ok := k.buckets[key]
	if !ok {
		b = &bucket{lim: rate.NewLimiter(k.limit, k.burst)}
		k.buckets[key] = b
	}
	b.seen = now

	r := b.lim.ReserveN(now, 1)
	if delay := r.DelayFrom(now); delay > 0 {
		r.CancelAt(now)
		return false, delay
	}
	return true, 0
}

// RateLimit rejects with 429 when the key's bucket is empty.
func RateLimit(l *KeyedLimiter, key func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, wait := l.Allow(key(c))
		if !ok {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			httputil.Abort(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many attempts. Try again in a minute.")
			return
		}
		c.Next()
	}
}

func ByIP(c *gin.Context) string { return c.ClientIP() }

func ByUser(c *gin.Context) string {
	return strconv.FormatInt(ActorFrom(c).UserID, 10)
}
