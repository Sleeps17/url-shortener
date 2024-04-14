package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	redisCache "url-shortener/internal/cache/redis-cache"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/delete"
	"url-shortener/internal/http-server/get"
	"url-shortener/internal/http-server/middleware"
	"url-shortener/internal/http-server/save"
	"url-shortener/internal/http-server/update"
	"url-shortener/internal/logger"
	"url-shortener/internal/storage/mongodb"
)

func main() {
	// TODO: init config
	cfg := config.MustLoad()

	// TODO: init logger
	log := logger.MustSetup(cfg.Env)
	log.Info("logger started")

	// TODO: init cache
	c := redisCache.MustNew(cfg.CacheConfig.ConnectionString, cfg.CacheConfig.DB, cfg.CacheConfig.Timeout, cfg.CacheConfig.Capacity)

	// TODO: init database
	s := mongodb.MustNew(
		cfg.DBConfig.Timeout,
		c,
		cfg.DBConfig.ConnectionString,
		cfg.DBConfig.DBName,
		cfg.DBConfig.CollectionName,
	)
	log.Info("database started")

	// TODO: init server
	router := gin.Default()
	router.Use(middleware.GetCreator())
	a := router.Group("/", gin.BasicAuth(gin.Accounts{
		"pasha": "1234",
		"vova":  "9876",
	}))

	a.POST("/", save.Save(log, s))
	router.GET("/:username/:alias", get.Get(log, s))
	a.DELETE("/", delete.Delete(log, s))
	a.PUT("/", update.Update(log, s))

	srv := &http.Server{
		Addr:         cfg.HttpServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	// TODO: start server
	log.Info("server started")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err.Error())
			return
		}
	}()

	// TODO: graceful shutdown
	done := make(chan os.Signal, 1)
	defer close(done)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Close(ctx); err != nil {
		log.Error("failed to close db", slog.String("error", err.Error()))
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
		return
	}

	log.Info("server stopped")
}
