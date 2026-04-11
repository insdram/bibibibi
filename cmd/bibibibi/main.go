package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bibibibi/bibibibi/internal/api"
	"github.com/bibibibi/bibibibi/internal/store"
)

func main() {
	dbPath := os.Getenv("BIBIBIBI_DB_PATH")
	if dbPath == "" {
		dbPath = "bibibibi.db"
	}

	if err := store.InitDB(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check 端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		db := store.GetDB()
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database unavailable"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database ping failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	api.RegisterRoutes(r)

	distPath := os.Getenv("FRONTEND_DIST")
	if distPath == "" {
		distPath = "./dist"
	}

	log.Printf("静态文件路径: %s", distPath)

	// 静态资源
	r.Static("/assets", distPath+"/assets")

	// 根路径返回 index.html
	r.GET("/", func(c *gin.Context) {
		c.File(distPath + "/index.html")
	})

	// 其他路由返回 index.html（SPA）
	r.NoRoute(func(c *gin.Context) {
		c.File(distPath + "/index.html")
	})

	port := os.Getenv("BIBIBIBI_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	go func() {
		log.Printf("bibibibi 服务启动在 http://0.0.0.0:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("强制关闭服务: %v", err)
	}

	log.Println("服务已关闭")
}
