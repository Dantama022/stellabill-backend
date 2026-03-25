package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/config"
	"stellarbill-backend/internal/routes"
)

var runServer = func(router *gin.Engine, addr string) error {
	return router.Run(addr)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()
	configureGinMode(cfg.Env)

	router := newRouter()

	addr := serverAddr(cfg.Port, os.Getenv("PORT"))
	log.Printf("Stellarbill backend listening on %s", addr)
	return runServer(router, addr)
}

func configureGinMode(env string) {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
		return
	}
	gin.SetMode(gin.DebugMode)
}

func newRouter() *gin.Engine {
	router := gin.Default()
	routes.Register(router)
	return router
}

func serverAddr(cfgPort, envPort string) string {
	if envPort != "" {
		return ":" + envPort
	}
	return ":" + cfgPort
}
