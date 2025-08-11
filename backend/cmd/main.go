package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	_ "wbL0/docs"
	"wbL0/internal/config"
	"wbL0/internal/db/postgres"
	redisClient "wbL0/internal/db/redis"
	"wbL0/internal/http/handler/orderHandler"
	"wbL0/internal/http/middleware"
	"wbL0/internal/http/routes"
	"wbL0/internal/kafka/consumer"
	"wbL0/internal/lib/logger"
	"wbL0/internal/metrics"
	orderRepoPostgres2 "wbL0/internal/repository/postgres/orderRepoPostgres"
	orderRepoRedis2 "wbL0/internal/repository/redis/orderRepoRedis"
	"wbL0/internal/service/orderService"
)

func main() {
	cfg := config.MustLoad()

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool := postgres.MustLoad(rootCtx, cfg)
	rdb := redisClient.NewRedisClient(rootCtx, cfg)

	log := logger.SetupLogger(cfg.App.Level)

	orderRepoPostgres := orderRepoPostgres2.NewPostgresRepository(dbPool, log)
	orderRepoRedis := orderRepoRedis2.NewRedisRepo(rdb, log)

	orderService := orderService.NewOrderService(orderRepoPostgres, orderRepoRedis, log, cfg.Redis.TTL)

	orderHandler := orderHandler.NewOrderHandler(orderService, log)

	metrics.Init()

	r := gin.Default()
	r.Use(metrics.MetricsMiddleware())
	r.Use(middleware.TimeoutMiddleware(cfg.Server.Timeout))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	routes.InitRoutes(r, *orderHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer wg.Done()
		if err := consumer.ConsumeMessage(ctx, cfg, orderService, log); err != nil {
			log.Error("Kafka consumer failed", "error", err)
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		log.Info("Listening and serving HTTP on ", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	cancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	} else {
		log.Info("Server exiting with graceful shutdown")
	}

	dbPool.Close()
	if err := rdb.Close(); err != nil {
		log.Error("Failed to close Redis connection", "error", err)
	}
}
