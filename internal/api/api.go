package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/service"
)

var (
	userService    = service.NewUserService()
	bibiService    = service.NewBibiService()
	tagService     = service.NewTagService()
	commentService = service.NewCommentService()
	likeService    = service.NewLikeService()
	systemService  = service.NewSystemService()
	tokenService   = service.NewTokenService()
	feedService    = service.NewFeedService()
)

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RegisterRoutes 注册 API 路由
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.Use(corsMiddleware())
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", handleLogin)
			auth.POST("/register", handleRegister)
		}

		// 用户相关
		user := api.Group("/user")
		{
			user.GET("/me", authMiddleware(), handleGetCurrentUser)
			user.PUT("/me", authMiddleware(), handleUpdateCurrentUser)
			user.GET("/list", handleGetUsers)
		}

		// Token 管理
		tokens := api.Group("/tokens")
		tokens.Use(authMiddleware())
		{
			tokens.GET("", handleGetTokens)
			tokens.POST("", handleCreateToken)
			tokens.DELETE("/:id", handleDeleteToken)
		}

		// Bibi 相关
		bibis := api.Group("/bibis")
		{
			bibis.GET("", handleGetBibis)
			bibis.POST("", authMiddleware(), handleCreateBibi)
			bibis.GET("/:id", handleGetBibi)
			bibis.PUT("/:id", authMiddleware(), handleUpdateBibi)
			bibis.DELETE("/:id", authMiddleware(), handleDeleteBibi)
			bibis.POST("/:id/pin", authMiddleware(), handleTogglePin)
			bibis.GET("/search", handleSearchBibis)
			bibis.POST("/:id/like", handleToggleLike)

			// 评论相关
			bibis.GET("/:id/comments", handleGetComments)
			bibis.POST("/:id/comments", handleCreateComment)
		}

		// 标签相关
		tags := api.Group("/tags")
		{
			tags.GET("", handleGetTags)
			tags.POST("", authMiddleware(), handleCreateTag)
			tags.PUT("/:id", authMiddleware(), handleUpdateTag)
			tags.DELETE("/:id", authMiddleware(), handleDeleteTag)
		}

		// 评论相关
		comments := api.Group("/comments")
		{
			comments.PUT("/:id", authMiddleware(), handleUpdateComment)
			comments.DELETE("/:id", authMiddleware(), handleDeleteComment)
		}

		// 系统设置（仅管理员）
		settings := api.Group("/settings")
		settings.Use(adminMiddleware())
		{
			settings.GET("", handleGetSettings)
			settings.PUT("", handleUpdateSettings)
		}

		// 公开系统设置（任何人可查看注册状态）
		api.GET("/public/settings", handleGetPublicSettings)

		// 广场数据源管理（仅管理员）
		feeds := api.Group("/feeds")
		feeds.Use(adminMiddleware())
		{
			feeds.GET("", handleGetFeedSources)
			feeds.POST("", handleCreateFeedSource)
			feeds.DELETE("/:id", handleDeleteFeedSource)
			feeds.POST("/sync", handleSyncFeedSources)
		}

		// 公开的远程笔记（用于广场）
		api.GET("/public/remote-bibis", handleGetRemoteBibis)
	}
}

// authMiddleware 认证中间件（支持 JWT 和新 Token 格式）
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 先尝试新的 Token 验证
		userID, err := tokenService.ValidateToken(token)
		if err != nil {
			// 如果失败，尝试旧的 JWT 验证（兼容旧 token）
			userID, err = service.ParseToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
				c.Abort()
				return
			}
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// adminMiddleware 管理员认证中间件
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 先尝试新的 Token 验证
		userID, err := tokenService.ValidateToken(token)
		if err != nil {
			// 如果失败，尝试旧的 JWT 验证（兼容旧 token）
			userID, err = service.ParseToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
				c.Abort()
				return
			}
		}

		user, err := userService.GetUserByID(userID)
		if err != nil || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// handleLogin 处理登录请求
func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	token, user, err := userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// handleGetTokens 获取用户的 Token 列表
func handleGetTokens(c *gin.Context) {
	userID := c.GetUint("userID")

	tokens, err := tokenService.GetTokensByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 Token 列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// handleCreateToken 创建新 Token
func handleCreateToken(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Description     string `json:"description"`
		ExpiresInHours   *int   `json:"expires_in_hours"` // nil 或 0 表示不过期
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var expiresInHours *int
	if req.ExpiresInHours != nil && *req.ExpiresInHours > 0 {
		expiresInHours = req.ExpiresInHours
	}

	token, err := tokenService.CreateToken(userID, req.Description, expiresInHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// handleDeleteToken 删除 Token
func handleDeleteToken(c *gin.Context) {
	userID := c.GetUint("userID")

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 Token ID"})
		return
	}

	if err := tokenService.DeleteToken(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleRegister 处理注册请求
func handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := userService.Register(req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// handleGetCurrentUser 获取当前用户信息
func handleGetCurrentUser(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// handleGetUsers 获取所有用户
func handleGetUsers(c *gin.Context) {
	users, err := userService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// handleUpdateCurrentUser 更新当前用户信息
func handleUpdateCurrentUser(c *gin.Context) {
	userID := c.GetUint("userID")
	var req struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Website  string `json:"website"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := userService.UpdateUser(userID, req.Username, req.Nickname, req.Email, req.Website, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// handleGetBibis 处理获取笔记列表请求
func handleGetBibis(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	visibility := c.Query("visibility")
	creatorIDStr := c.Query("creator_id")

	var creatorID *uint
	if creatorIDStr != "" {
		if id, err := strconv.ParseUint(creatorIDStr, 10, 32); err == nil {
			cid := uint(id)
			creatorID = &cid
		}
	}

	bibis, total, err := bibiService.GetBibis(page, pageSize, visibility, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取笔记列表失败"})
		return
	}

	// 确保 bibis 不是 nil（JSON 序列化 nil 切片会变成 null）
	if bibis == nil {
		bibis = []model.Bibi{}
	}

	// 如果是获取广场（没有指定 creator_id），合并远程笔记
	if creatorID == nil {
		remoteBibis, err := feedService.GetAllRemoteBibis()
		if err != nil {
			// 如果获取远程笔记失败，仍然返回本地笔记
			c.JSON(http.StatusOK, gin.H{
				"bibis":     bibis,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
			return
		}
		// 合并总数
		total += int64(len(remoteBibis))
		// 将远程笔记转换为统一格式
		allBibis := make([]map[string]interface{}, 0, len(bibis)+len(remoteBibis))
		for _, b := range bibis {
			allBibis = append(allBibis, map[string]interface{}{
				"id":            b.ID,
				"content":       b.Content,
				"visibility":    b.Visibility,
				"pinned":       b.Pinned,
				"like_count":    b.LikeCount,
				"comment_count": b.CommentCount,
				"liked":        false,
				"created_at":    b.CreatedAt.Format(time.RFC3339),
				"creator":        b.Creator,
				"tags":          b.Tags,
				"comments":      b.Comments,
				"is_remote":     false,
			})
		}
		for _, rb := range remoteBibis {
			allBibis = append(allBibis, map[string]interface{}{
				"id":            rb.ID,
				"content":       rb.Content,
				"visibility":    "PUBLIC",
				"pinned":       false,
				"like_count":    rb.LikeCount,
				"comment_count": rb.CommentCount,
				"liked":        false,
				"created_at":    rb.CreatedAt,
				"creator": map[string]interface{}{
					"id":       0,
					"username": rb.Creator.Username,
					"nickname": rb.Creator.Nickname,
					"avatar":   rb.Creator.Avatar,
				},
				"tags":       []interface{}{},
				"comments":   rb.Comments,
				"is_remote":  true,
				"source_url": rb.SourceURL,
			})
		}

		// 按时间倒序排列
		sort.Slice(allBibis, func(i, j int) bool {
			return allBibis[i]["created_at"].(string) > allBibis[j]["created_at"].(string)
		})

		c.JSON(http.StatusOK, gin.H{
			"bibis":     allBibis,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bibis":     bibis,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handleCreateBibi 处理创建笔记请求
func handleCreateBibi(c *gin.Context) {
	var req struct {
		Content    string `json:"content" binding:"required"`
		Visibility string `json:"visibility"`
		TagIDs     []uint `json:"tag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID := c.GetUint("userID")
	if req.Visibility == "" {
		req.Visibility = "PUBLIC"
	}

	bibi, err := bibiService.CreateBibi(userID, req.Content, req.Visibility, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建笔记失败"})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleGetBibi 处理获取笔记详情请求
func handleGetBibi(c *gin.Context) {
	id := c.Param("id")

	bibi, err := bibiService.GetBibiByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleUpdateBibi 处理更新笔记请求
func handleUpdateBibi(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetUint("userID")

	var req struct {
		Content    string `json:"content" binding:"required"`
		Visibility string `json:"visibility"`
		TagIDs     []uint `json:"tag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.Visibility == "" {
		req.Visibility = "PUBLIC"
	}

	bibi, err := bibiService.UpdateBibi(id, req.Content, req.Visibility, req.TagIDs, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleDeleteBibi 处理删除笔记请求
func handleDeleteBibi(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetUint("userID")

	if err := bibiService.DeleteBibi(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleTogglePin 处理切换置顶状态请求
func handleTogglePin(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetUint("userID")

	bibi, err := bibiService.TogglePin(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleToggleLike 处理切换点赞状态请求
func handleToggleLike(c *gin.Context) {
	id := c.Param("id")

	if err := likeService.ToggleLike(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"liked": true})
}

// handleSearchBibis 处理搜索笔记请求
func handleSearchBibis(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	bibis, total, err := bibiService.SearchBibis(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索笔记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bibis": bibis,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// handleGetTags 处理获取标签列表请求
func handleGetTags(c *gin.Context) {
	creatorID, _ := strconv.ParseUint(c.Query("creator_id"), 10, 32)

	tags, err := tagService.GetTags(uint(creatorID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取标签列表失败"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// handleCreateTag 处理创建标签请求
func handleCreateTag(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID := c.GetUint("userID")
	tag, err := tagService.CreateTag(req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// handleUpdateTag 处理更新标签请求
func handleUpdateTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的标签 ID"})
		return
	}

	userID := c.GetUint("userID")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	tag, err := tagService.UpdateTag(uint(id), req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// handleDeleteTag 处理删除标签请求
func handleDeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的标签 ID"})
		return
	}

	userID := c.GetUint("userID")

	if err := tagService.DeleteTag(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleGetComments 处理获取评论列表请求
func handleGetComments(c *gin.Context) {
	bibiID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := commentService.GetCommentsByBibiID(bibiID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// handleCreateComment 处理创建评论请求
func handleCreateComment(c *gin.Context) {
	bibiID := c.Param("id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Website  string `json:"website"`
		Content  string `json:"content" binding:"required"`
		ParentID uint   `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	comment, err := commentService.CreateComment(bibiID, req.Name, req.Email, req.Website, req.Content, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// handleUpdateComment 处理更新评论请求
func handleUpdateComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论 ID"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Email   string `json:"email" binding:"required"`
		Website string `json:"website"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	comment, err := commentService.UpdateComment(uint(id), req.Name, req.Email, req.Website, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新评论失败"})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// handleDeleteComment 处理删除评论请求
func handleDeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论 ID"})
		return
	}

	userID := c.GetUint("userID")

	if err := commentService.DeleteComment(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleGetSettings 获取系统设置
func handleGetSettings(c *gin.Context) {
	registrationEnabled, _ := systemService.GetSetting("registration_enabled")
	gravatarSource, _ := systemService.GetSetting("gravatar_source")
	if gravatarSource == "" {
		gravatarSource = "https://www.gravatar.com/avatar/"
	}
	c.JSON(http.StatusOK, gin.H{
		"registration_enabled": registrationEnabled == "true",
		"gravatar_source":      gravatarSource,
	})
}

// handleUpdateSettings 更新系统设置
func handleUpdateSettings(c *gin.Context) {
	var req struct {
		RegistrationEnabled *bool   `json:"registration_enabled"`
		GravatarSource      *string `json:"gravatar_source"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.RegistrationEnabled != nil {
		value := "false"
		if *req.RegistrationEnabled {
			value = "true"
		}
		if err := systemService.UpdateSetting("registration_enabled", value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
			return
		}
	}

	if req.GravatarSource != nil && *req.GravatarSource != "" {
		if err := systemService.UpdateSetting("gravatar_source", *req.GravatarSource); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// handleGetPublicSettings 获取公开系统设置（无需登录）
func handleGetPublicSettings(c *gin.Context) {
	registrationEnabled, _ := systemService.GetSetting("registration_enabled")
	gravatarSource, _ := systemService.GetSetting("gravatar_source")
	if gravatarSource == "" {
		gravatarSource = "https://www.gravatar.com/avatar/"
	}
	c.JSON(http.StatusOK, gin.H{
		"registration_enabled": registrationEnabled == "true",
		"gravatar_source":      gravatarSource,
	})
}

// handleGetFeedSources 获取所有广场数据源
func handleGetFeedSources(c *gin.Context) {
	sources, err := feedService.GetFeedSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取数据源失败"})
		return
	}
	c.JSON(http.StatusOK, sources)
}

// handleCreateFeedSource 创建广场数据源
func handleCreateFeedSource(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	source, err := feedService.CreateFeedSource(req.Name, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建数据源失败"})
		return
	}

	// 立即获取一次数据（同步执行，返回错误给前端）
	if _, err := feedService.FetchBibisFromSource(source.URL); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"source": source,
			"warning": fmt.Sprintf("数据源创建成功但同步失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, source)
}

// handleDeleteFeedSource 删除广场数据源
func handleDeleteFeedSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的数据源 ID"})
		return
	}

	if err := feedService.DeleteFeedSource(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据源失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleSyncFeedSources 同步所有广场数据源（远程笔记不再存储，此接口仅返回成功）
func handleSyncFeedSources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "同步成功"})
}

// handleGetRemoteBibis 获取远程笔记（直接调用远程API）
func handleGetRemoteBibis(c *gin.Context) {
	bibis, err := feedService.GetAllRemoteBibis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取远程笔记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bibis": bibis,
		"total": len(bibis),
	})
}
