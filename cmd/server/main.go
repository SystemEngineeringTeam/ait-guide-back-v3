// Package main はサーバーのエントリポイントを提供する。
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/config"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/handler"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/infra/loader"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/infra/postgres"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/usecase"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// DB接続
	pool, err := postgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("connected to database")

	// マイグレーション
	if err := postgres.Migrate(ctx, pool, "db/migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations completed")

	// シードデータ投入
	seeder := loader.NewCSVSeeder(pool)
	if err := seeder.SeedAll(ctx, cfg.SeedDir); err != nil {
		log.Fatalf("failed to seed data: %v", err)
	}
	log.Println("seed data loaded")

	// DI: リポジトリ
	buildingRepo := postgres.NewBuildingRepository(pool)
	nodeRepo := postgres.NewNodeRepository(pool)
	roomRepo := postgres.NewRoomRepository(pool)
	routeRepo := postgres.NewRouteRepository(pool)

	// DI: ユースケース
	buildingUC := usecase.NewBuildingUsecase(buildingRepo)
	nodeUC := usecase.NewNodeUsecase(nodeRepo)
	roomUC := usecase.NewRoomUsecase(roomRepo)
	routeUC := usecase.NewRouteUsecase(nodeRepo, routeRepo)

	// DI: ハンドラー
	healthH := handler.NewHealthHandler()
	buildingH := handler.NewBuildingHandler(buildingUC)
	nodeH := handler.NewNodeHandler(nodeUC)
	roomH := handler.NewRoomHandler(roomUC)
	routeH := handler.NewRouteHandler(routeUC)

	// ルーター
	router := handler.Router(healthH, buildingH, roomH, nodeH, routeH)

	// サーバー起動
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Printf("server starting on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
