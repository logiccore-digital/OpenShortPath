package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"openshortpath/server/constants"
)

func TestRequireScope_NoScopesInContext(t *testing.T) {
	// No scopes in context means JWT authentication (backward compatibility)
	handler := RequireScope("shorten_url")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)

	handler(c)

	// Should allow request (JWT tokens bypass scope checks)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_ValidScope(t *testing.T) {
	handler := RequireScope("shorten_url")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	
	// Set scopes in context (from API key)
	scopes := []string{"shorten_url", "read_urls"}
	c.Set(constants.ContextKeyScopes, scopes)

	handler(c)

	// Should allow request
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_MissingScope(t *testing.T) {
	handler := RequireScope("write_urls")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	
	// Set scopes in context (from API key) - missing write_urls
	scopes := []string{"shorten_url", "read_urls"}
	c.Set(constants.ContextKeyScopes, scopes)

	handler(c)

	// Should deny request
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireScope_EmptyScopes(t *testing.T) {
	handler := RequireScope("shorten_url")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	
	// Set empty scopes in context
	scopes := []string{}
	c.Set(constants.ContextKeyScopes, scopes)

	handler(c)

	// Should deny request
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireScope_InvalidScopesType(t *testing.T) {
	handler := RequireScope("shorten_url")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	
	// Set invalid type in context
	c.Set(constants.ContextKeyScopes, "not-a-slice")

	handler(c)

	// Should deny request
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireScope_MultipleScopes(t *testing.T) {
	testCases := []struct {
		name          string
		requiredScope string
		apiKeyScopes  []string
		shouldAllow   bool
	}{
		{"Has required scope", "shorten_url", []string{"shorten_url", "read_urls"}, true},
		{"Has required scope (single)", "shorten_url", []string{"shorten_url"}, true},
		{"Missing required scope", "write_urls", []string{"shorten_url", "read_urls"}, false},
		{"Has all scopes", "read_urls", []string{"shorten_url", "read_urls", "write_urls"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := RequireScope(tc.requiredScope)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
			c.Set(constants.ContextKeyScopes, tc.apiKeyScopes)

			handler(c)

			if tc.shouldAllow {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusForbidden, w.Code)
			}
		})
	}
}

