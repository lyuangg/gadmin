package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/logger"
	"github.com/lyuangg/gadmin/routes"
	"github.com/lyuangg/gadmin/tasks"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

type slogGinWriter struct {
	logger logger.ILogger
	level  slog.Level
}

func (w *slogGinWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}
	w.logger.Log(context.Background(), w.level, msg)
	return len(p), nil
}

func main() {
	configPath := flag.String("c", "", "配置文件路径 (例如: -c ./config.yml)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Default().ErrorContext(context.Background(), "配置加载失败", "error", err)
		os.Exit(1)
	}

	appInstance := app.NewApp(cfg)
	defer appInstance.Close()

	utils.InitJWT(cfg)
	gin.SetMode(cfg.GinMode)

	gin.DefaultWriter = &slogGinWriter{logger: appInstance.Logger(), level: slog.LevelInfo}
	gin.DefaultErrorWriter = &slogGinWriter{logger: appInstance.Logger(), level: slog.LevelError}

	router := gin.Default()
	routes.SetupRoutes(router, appInstance)

	scanner := routes.NewRouteScanner(router, appInstance)
	if err := scanner.ScanAndImport(); err != nil {
		appInstance.Logger().ErrorContext(context.Background(), "路由扫描失败", "error", err)
	}

	// 每天凌晨清理操作日志，保留条数见配置 operation_log_retain_count
	tasks.StartOperationLogCleanScheduler(appInstance)

	appInstance.Logger().InfoContext(context.Background(), "服务器启动", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		appInstance.Logger().ErrorContext(context.Background(), "服务器启动失败", "error", err)
		os.Exit(1)
	}
}
