package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken(t *testing.T) {
	// Test that token is generated
	token1, err := GenerateCSRFToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	// Test that tokens are unique
	token2, err := GenerateCSRFToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "Generated tokens should be unique")

	// Test token length (base64 encoded 32 bytes)
	assert.Greater(t, len(token1), 40, "Token should be reasonably long")
}

func TestSetCSRFTokenMiddleware_CreatesToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SetCSRFTokenMiddleware(false, ""))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Check cookie is set
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			found = true
			assert.NotEmpty(t, cookie.Value)
			assert.False(t, cookie.HttpOnly, "CSRF cookie should NOT be HttpOnly so JavaScript can read it")
			break
		}
	}
	assert.True(t, found, "CSRF cookie should be set")

	// Check header is set
	headerToken := w.Header().Get(CSRFHeaderName)
	assert.NotEmpty(t, headerToken, "CSRF header should be set")
}

func TestSetCSRFTokenMiddleware_SkipsIfExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	existingToken := "existing-token-123"

	router := gin.New()
	router.Use(SetCSRFTokenMiddleware(false, ""))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: existingToken})
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Check that no new cookie was set
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 0, "Should not set new cookie when one already exists")
}

func TestCSRFMiddleware_GET_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "GET requests should be allowed without CSRF token")
}

func TestCSRFMiddleware_OPTIONS_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "OPTIONS requests should be allowed without CSRF token")
}

func TestCSRFMiddleware_POST_WithValidToken_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token := "valid-csrf-token"

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: token})
	req.Header.Set(CSRFHeaderName, token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "POST with valid CSRF token should be allowed")
}

func TestCSRFMiddleware_POST_WithMissingCookie_Rejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set(CSRFHeaderName, "some-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code, "POST without CSRF cookie should be rejected")
}

func TestCSRFMiddleware_POST_WithMissingHeader_Rejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "some-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code, "POST without CSRF header should be rejected")
}

func TestCSRFMiddleware_POST_WithMismatchedToken_Rejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFMiddleware(false))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "token-in-cookie"})
	req.Header.Set(CSRFHeaderName, "different-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code, "POST with mismatched CSRF tokens should be rejected")
}

func TestCSRFMiddleware_AllMutationMethods_Validated(t *testing.T) {
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			token := "valid-token"

			router := gin.New()
			router.Use(CSRFMiddleware(false))
			router.Handle(method, "/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			// Test with valid token
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/test", nil)
			req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: token})
			req.Header.Set(CSRFHeaderName, token)
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code, method+" with valid token should be allowed")

			// Test without token
			w = httptest.NewRecorder()
			req, _ = http.NewRequest(method, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, 403, w.Code, method+" without token should be rejected")
		})
	}
}
