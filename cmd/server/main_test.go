package main

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServerAddr(t *testing.T) {
	if got := serverAddr("8080", ""); got != ":8080" {
		t.Fatalf("serverAddr = %q, want %q", got, ":8080")
	}
	if got := serverAddr("8080", "9090"); got != ":9090" {
		t.Fatalf("serverAddr = %q, want %q", got, ":9090")
	}
}

func TestConfigureGinMode(t *testing.T) {
	configureGinMode("production")
	if gin.Mode() != gin.ReleaseMode {
		t.Fatalf("gin.Mode = %q, want %q", gin.Mode(), gin.ReleaseMode)
	}

	configureGinMode("development")
	if gin.Mode() != gin.DebugMode {
		t.Fatalf("gin.Mode = %q, want %q", gin.Mode(), gin.DebugMode)
	}
}

func TestNewRouterRegistersExpectedRoutes(t *testing.T) {
	router := newRouter()
	routes := router.Routes()

	if len(routes) == 0 {
		t.Fatal("expected router to register routes")
	}
}

func TestRunUsesConfiguredPort(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("PORT", "9191")

	previous := runServer
	defer func() { runServer = previous }()

	called := false
	runServer = func(router *gin.Engine, addr string) error {
		called = true
		if addr != ":9191" {
			t.Fatalf("addr = %q, want %q", addr, ":9191")
		}
		if router == nil {
			t.Fatal("expected router to be initialized")
		}
		return nil
	}

	if err := run(); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !called {
		t.Fatal("expected runServer to be called")
	}
	if gin.Mode() != gin.ReleaseMode {
		t.Fatalf("gin.Mode = %q, want %q", gin.Mode(), gin.ReleaseMode)
	}
}

func TestRunPropagatesServerError(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("PORT", "8080")

	previous := runServer
	defer func() { runServer = previous }()

	wantErr := errors.New("listen failed")
	runServer = func(_ *gin.Engine, _ string) error {
		return wantErr
	}

	if err := run(); !errors.Is(err, wantErr) {
		t.Fatalf("run error = %v, want %v", err, wantErr)
	}
}
