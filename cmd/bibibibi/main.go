package main

import (
	"log"
	"os"

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
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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

	log.Printf("bibibibi 服务启动在 http://0.0.0.0:%s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
