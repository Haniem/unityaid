package http

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	buckets := map[string]*rateBucket{}
	var mu sync.Mutex

	return func(c *gin.Context) {
		key := clientKey(c)
		now := time.Now()

		mu.Lock()
		for bucketKey, bucket := range buckets {
			if now.Sub(bucket.lastSeen) > window*4 {
				delete(buckets, bucketKey)
			}
		}
		bucket := buckets[key]
		if bucket == nil || now.After(bucket.resetAt) {
			bucket = &rateBucket{resetAt: now.Add(window)}
			buckets[key] = bucket
		}
		bucket.count++
		bucket.lastSeen = now
		allowed := bucket.count <= limit
		retryAfter := int(time.Until(bucket.resetAt).Seconds())
		mu.Unlock()

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited", "message": "Too many requests"})
			return
		}
		c.Next()
	}
}

func clientKey(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		host, _, err := net.SplitHostPort(forwarded)
		if err == nil {
			return host
		}
		return forwarded
	}
	return c.ClientIP()
}
