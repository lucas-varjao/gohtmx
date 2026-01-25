// backend/internal/middleware/rate_limit_test.go

package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// Tests only verify observable behavior via HTTP. No inspection of limiter.ips/mu or assert.Same.

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Allow Request Under Limit", func(t *testing.T) {
		ipLimiter := NewIPRateLimiter(10, 10, time.Minute)
		handlerCalled := false

		r := gin.New()
		r.Use(RateLimitMiddleware(ipLimiter))
		r.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.10")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Block Request Over Limit", func(t *testing.T) {
		ipLimiter := NewIPRateLimiter(0.1, 1, time.Minute)
		handlerCalled := false

		r := gin.New()
		r.Use(RateLimitMiddleware(ipLimiter))
		r.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		})

		ip := "192.168.1.20"
		limiter := ipLimiter.GetLimiter(ip)
		limiter.Allow() // exhaust the single token

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", ip)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "limite de requisições excedido")
	})

	t.Run("Multiple Requests From Same IP", func(t *testing.T) {
		ipLimiter := NewIPRateLimiter(rate.Every(30*time.Second), 2, time.Minute)
		r := gin.New()
		r.Use(RateLimitMiddleware(ipLimiter))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		ip := "192.168.1.30"
		makeRequest := func() (int, string) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", ip)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			return w.Code, w.Body.String()
		}

		code1, _ := makeRequest()
		assert.Equal(t, http.StatusOK, code1)

		code2, _ := makeRequest()
		assert.Equal(t, http.StatusOK, code2)

		code3, body3 := makeRequest()
		assert.Equal(t, http.StatusTooManyRequests, code3)
		assert.Contains(t, body3, "limite de requisições excedido")
	})

	t.Run("Different IPs Have Separate Rate Limits", func(t *testing.T) {
		ipLimiter := NewIPRateLimiter(0.1, 1, time.Minute)
		r := gin.New()
		r.Use(RateLimitMiddleware(ipLimiter))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		ip1 := "192.168.1.50"
		ip2 := "192.168.1.51"
		makeRequest := func(ip string) (int, string) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", ip)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			return w.Code, w.Body.String()
		}

		code1, _ := makeRequest(ip1)
		assert.Equal(t, http.StatusOK, code1)

		code2, body2 := makeRequest(ip1)
		assert.Equal(t, http.StatusTooManyRequests, code2)
		assert.Contains(t, body2, "limite de requisições excedido")

		code3, _ := makeRequest(ip2)
		assert.Equal(t, http.StatusOK, code3)
	})

	t.Run("Concurrent Requests From Same IP", func(t *testing.T) {
		ipLimiter := NewIPRateLimiter(10, 5, time.Minute)
		r := gin.New()
		r.Use(RateLimitMiddleware(ipLimiter))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		ip := "192.168.1.40"
		var wg sync.WaitGroup
		requestCount := 10
		results := make([]bool, requestCount)

		for i := 0; i < requestCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", ip)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				results[index] = w.Code == http.StatusOK
			}(i)
		}

		wg.Wait()

		allowed := 0
		for _, ok := range results {
			if ok {
				allowed++
			}
		}
		// Burst is 5; at most 5 requests allowed, non-deterministic under concurrency
		assert.LessOrEqual(t, allowed, 5)
	})
}
