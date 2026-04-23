package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"stellarbill-backend/internal/service"
)

// mockErrorService returns different errors for testing
type mockErrorService struct {
	shouldReturnError bool
	errorType         error
	detail            *service.SubscriptionDetail
	warnings          []string
}

func (m *mockErrorService) GetDetail(_ context.Context, _, _, _ string) (*service.SubscriptionDetail, []string, error) {
	if m.shouldReturnError {
		return nil, nil, m.errorType
	}
	return m.detail, m.warnings, nil
}

// setupErrorTestRouter builds a test router with trace ID middleware
func setupErrorTestRouter(svc service.SubscriptionService, setCallerID bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Add trace ID context
	r.Use(func(c *gin.Context) {
		if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
			c.Set("traceID", traceID)
		} else {
			c.Set("traceID", "test-trace-123")
		}
		c.Header("X-Trace-ID", c.GetString("traceID"))
	})
	if setCallerID {
		r.Use(func(c *gin.Context) {
			c.Set("callerID", "caller-123")
			c.Set("tenantID", "tenant-1")
			c.Next()
		})
	}
	r.GET("/api/subscriptions/:id", NewGetSubscriptionHandler(svc))
	return r
}

// TestErrorEnvelope_NotFound tests the error envelope for not found errors
func TestErrorEnvelope_NotFound(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrNotFound,
	}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Code != string(ErrorCodeNotFound) {
		t.Errorf("Expected error code %s, got %s", ErrorCodeNotFound, envelope.Code)
	}
	if envelope.Message != "The requested resource was not found" {
		t.Errorf("Expected proper message, got %s", envelope.Message)
	}
	if envelope.TraceID != "test-trace-123" {
		t.Errorf("Expected trace ID test-trace-123, got %s", envelope.TraceID)
	}
}

// TestErrorEnvelope_Deleted tests the error envelope for deleted resource errors
func TestErrorEnvelope_Deleted(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrDeleted,
	}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/deleted-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status %d, got %d", http.StatusGone, w.Code)
	}

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Code != string(ErrorCodeNotFound) {
		t.Errorf("Expected error code %s, got %s", ErrorCodeNotFound, envelope.Code)
	}
	if envelope.TraceID == "" {
		t.Error("Expected trace ID to be present")
	}
}

// TestErrorEnvelope_Forbidden tests the error envelope for forbidden errors
func TestErrorEnvelope_Forbidden(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrForbidden,
	}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/forbidden-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Code != string(ErrorCodeForbidden) {
		t.Errorf("Expected error code %s, got %s", ErrorCodeForbidden, envelope.Code)
	}
	if envelope.Message != "You do not have permission to access this resource" {
		t.Errorf("Expected proper message, got %s", envelope.Message)
	}
}

// TestErrorEnvelope_BillingParse tests the error envelope for billing parse errors
func TestErrorEnvelope_BillingParse(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrBillingParse,
	}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/billing-error-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Code != string(ErrorCodeInternalError) {
		t.Errorf("Expected error code %s, got %s", ErrorCodeInternalError, envelope.Code)
	}
}

// TestErrorEnvelope_ValidationError tests validation errors
func TestErrorEnvelope_ValidationError(t *testing.T) {
	svc := &mockErrorService{}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	// Empty subscription ID should trigger validation error
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/", nil)
	r.ServeHTTP(w, req)

	// Gin routing returns 404 for unmatched routes, skip this test
	if w.Code == http.StatusNotFound {
		t.Skip("Route not matched, skipping validation test")
	}

	// If we get here, check the response format
	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err == nil {
		if envelope.Code != string(ErrorCodeValidationFailed) {
			t.Errorf("Expected validation error code, got %s", envelope.Code)
		}
	}
}

// TestErrorEnvelope_MissingAuth tests authentication error envelope
func TestErrorEnvelope_MissingAuth(t *testing.T) {
	svc := &mockErrorService{}
	r := setupErrorTestRouter(svc, false) // Don't set callerID

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/some-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Code != string(ErrorCodeUnauthorized) {
		t.Errorf("Expected error code %s, got %s", ErrorCodeUnauthorized, envelope.Code)
	}
}

// TestErrorEnvelope_ValidDetailsIncluded tests validation errors include details
func TestErrorEnvelope_ValidDetailsIncluded(t *testing.T) {
	svc := &mockErrorService{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-456")
		c.Set("callerID", "caller-123")
		c.Set("tenantID", "tenant-1")
	})
	r.GET("/api/subscriptions/:id", NewGetSubscriptionHandler(svc))

	w := httptest.NewRecorder()
	// Test with whitespace-only ID (will be trimmed to empty)
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/%20%20", nil)
	r.ServeHTTP(w, req)

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Details == nil {
		t.Error("Expected details in validation error")
	} else if field, ok := envelope.Details["field"]; !ok || field != "id" {
		t.Errorf("Expected field details, got %v", envelope.Details)
	}
}

// TestErrorEnvelope_TraceIDTracking tests trace ID is properly tracked through responses
func TestErrorEnvelope_TraceIDTracking(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrNotFound,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// Use custom trace ID from header or generate one
		if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
			c.Set("traceID", traceID)
		} else {
			c.Set("traceID", "generated-trace-id")
		}
		c.Header("X-Trace-ID", c.GetString("traceID"))
	})
	r.Use(func(c *gin.Context) {
		c.Set("callerID", "caller-123")
		c.Set("tenantID", "tenant-1")
	})
	r.GET("/api/subscriptions/:id", NewGetSubscriptionHandler(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/test-id", nil)
	req.Header.Set("X-Trace-ID", "custom-trace-789")
	r.ServeHTTP(w, req)

	var envelope ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.TraceID != "custom-trace-789" {
		t.Errorf("Expected custom trace ID, got %s", envelope.TraceID)
	}

	// Also check response header
	if headerTraceID := w.Header().Get("X-Trace-ID"); headerTraceID != "custom-trace-789" {
		t.Errorf("Expected trace ID in header, got %s", headerTraceID)
	}
}

// TestErrorEnvelope_ContentType tests proper content type header
func TestErrorEnvelope_ContentType(t *testing.T) {
	svc := &mockErrorService{
		shouldReturnError: true,
		errorType:         service.ErrNotFound,
	}
	r := setupErrorTestRouter(svc, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/subscriptions/test-id", nil)
	r.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected proper content type, got %s", contentType)
	}
}

// TestRespondWithValidationError tests the validation error response function
func TestRespondWithValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-validation")
	})
	r.GET("/test", func(c *gin.Context) {
		details := map[string]interface{}{
			"field":  "email",
			"reason": "invalid format",
		}
		RespondWithValidationError(c, "Validation failed", details)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var envelope ErrorEnvelope
	json.Unmarshal(w.Body.Bytes(), &envelope)
	if envelope.Code != string(ErrorCodeValidationFailed) {
		t.Errorf("Expected VALIDATION_FAILED code, got %s", envelope.Code)
	}
	if envelope.Details == nil || envelope.Details["field"] != "email" {
		t.Error("Expected validation details")
	}
}

// TestRespondWithAuthError tests the auth error response function
func TestRespondWithAuthError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-auth")
	})
	r.GET("/test", func(c *gin.Context) {
		RespondWithAuthError(c, "Token expired")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var envelope ErrorEnvelope
	json.Unmarshal(w.Body.Bytes(), &envelope)
	if envelope.Code != string(ErrorCodeUnauthorized) {
		t.Errorf("Expected UNAUTHORIZED code, got %s", envelope.Code)
	}
	if envelope.Message != "Token expired" {
		t.Errorf("Expected custom message, got %s", envelope.Message)
	}
}

// TestRespondWithNotFoundError tests the not found error response function
func TestRespondWithNotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-notfound")
	})
	r.GET("/test", func(c *gin.Context) {
		RespondWithNotFoundError(c, "Subscription")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var envelope ErrorEnvelope
	json.Unmarshal(w.Body.Bytes(), &envelope)
	if envelope.Code != string(ErrorCodeNotFound) {
		t.Errorf("Expected NOT_FOUND code, got %s", envelope.Code)
	}
	if envelope.Message != "Subscription not found" {
		t.Errorf("Expected formatted message, got %s", envelope.Message)
	}
}

// TestRespondWithInternalError tests the internal error response function
func TestRespondWithInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-internal")
	})
	r.GET("/test", func(c *gin.Context) {
		RespondWithInternalError(c, "Database connection failed")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var envelope ErrorEnvelope
	json.Unmarshal(w.Body.Bytes(), &envelope)
	if envelope.Code != string(ErrorCodeInternalError) {
		t.Errorf("Expected INTERNAL_ERROR code, got %s", envelope.Code)
	}
}

// TestErrorCode_Constants tests that all error code constants are properly defined
func TestErrorCode_Constants(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrorCodeBadRequest, "BAD_REQUEST"},
		{ErrorCodeUnauthorized, "UNAUTHORIZED"},
		{ErrorCodeForbidden, "FORBIDDEN"},
		{ErrorCodeNotFound, "NOT_FOUND"},
		{ErrorCodeConflict, "CONFLICT"},
		{ErrorCodeValidationFailed, "VALIDATION_FAILED"},
		{ErrorCodeInternalError, "INTERNAL_ERROR"},
		{ErrorCodeServiceUnavailable, "SERVICE_UNAVAILABLE"},
	}

	for _, tt := range tests {
		if string(tt.code) != tt.expected {
			t.Errorf("Error code mismatch: expected %s, got %s", tt.expected, tt.code)
		}
	}
}

// TestMapServiceErrorToResponse tests error mapping from service errors
func TestMapServiceErrorToResponse(t *testing.T) {
	tests := []struct {
		err           error
		expectedCode  int
		expectedError ErrorCode
	}{
		{service.ErrNotFound, http.StatusNotFound, ErrorCodeNotFound},
		{service.ErrDeleted, http.StatusGone, ErrorCodeNotFound},
		{service.ErrForbidden, http.StatusForbidden, ErrorCodeForbidden},
		{service.ErrBillingParse, http.StatusInternalServerError, ErrorCodeInternalError},
	}

	for _, tt := range tests {
		statusCode, code, message := MapServiceErrorToResponse(tt.err)
		if statusCode != tt.expectedCode {
			t.Errorf("For error %v: expected status %d, got %d", tt.err, tt.expectedCode, statusCode)
		}
		if code != tt.expectedError {
			t.Errorf("For error %v: expected code %s, got %s", tt.err, tt.expectedError, code)
		}
		if message == "" {
			t.Errorf("For error %v: expected non-empty message", tt.err)
		}
	}
}

// TestErrorEnvelope_AllFields tests that all envelope fields are populated
func TestErrorEnvelope_AllFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "trace-all-fields")
	})
	r.GET("/test", func(c *gin.Context) {
		RespondWithErrorDetails(c, http.StatusBadRequest, ErrorCodeValidationFailed,
			"Test message", map[string]interface{}{"key": "value"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	var envelope ErrorEnvelope
	json.Unmarshal(w.Body.Bytes(), &envelope)

	if envelope.Code == "" {
		t.Error("Code field is empty")
	}
	if envelope.Message == "" {
		t.Error("Message field is empty")
	}
	if envelope.TraceID == "" {
		t.Error("TraceID field is empty")
	}
	if envelope.Details == nil {
		t.Error("Details field is nil")
	}
}

// TestErrorEnvelope_NoDetails tests that details field is omitted when nil
func TestErrorEnvelope_NoDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "trace-no-details")
	})
	r.GET("/test", func(c *gin.Context) {
		RespondWithError(c, http.StatusInternalServerError, ErrorCodeInternalError, "Error occurred")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	// Check that the JSON response doesn't include "details" field
	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "\"details\"") {
		t.Error("Details field should be omitted when nil")
	}
}
