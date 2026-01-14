package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"openshortpath/server/config"
)

func TestAuthProviderHandler_GetAuthProvider(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		expectedAuth  string
		expectedSignup bool
	}{
		{
			name: "local auth with signup enabled",
			cfg: &config.Config{
				AuthProvider: "local",
				EnableSignup: true,
			},
			expectedAuth:  "local",
			expectedSignup: true,
		},
		{
			name: "local auth with signup disabled",
			cfg: &config.Config{
				AuthProvider: "local",
				EnableSignup: false,
			},
			expectedAuth:  "local",
			expectedSignup: false,
		},
		{
			name: "external_jwt auth (signup should be false)",
			cfg: &config.Config{
				AuthProvider: "external_jwt",
				EnableSignup: false,
			},
			expectedAuth:  "external_jwt",
			expectedSignup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthProviderHandler(tt.cfg)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth-provider", nil)

			handler.GetAuthProvider(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAuth, response["auth_provider"])
			assert.Equal(t, tt.expectedSignup, response["enable_signup"])
		})
	}
}
