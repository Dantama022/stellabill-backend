package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegister_HealthAndCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin header to be set")
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("payload.status = %v, want %q", payload["status"], "ok")
	}
}

func TestRegister_CORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine)

	req := httptest.NewRequest(http.MethodOptions, "http://localhost:8080/api/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("expected Access-Control-Allow-Methods header to be set")
	}
}

func TestRegister_GetSubscriptionShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/subscriptions/sub_123", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if payload["id"] != "sub_123" {
		t.Fatalf("payload.id = %v, want %q", payload["id"], "sub_123")
	}
	if _, ok := payload["plan_id"]; !ok {
		t.Fatalf("expected payload.plan_id to be present")
	}
	if _, ok := payload["customer"]; !ok {
		t.Fatalf("expected payload.customer to be present")
	}
}
