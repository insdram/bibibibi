package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bibibibi/bibibibi/internal/service"
)

var (
	userService    = service.NewUserService()
	bibiService    = service.NewBibiService()
	tagService     = service.NewTagService()
	commentService = service.NewCommentService()
	likeService    = service.NewLikeService()
	systemService  = service.NewSystemService()
)

// RegisterRoutes 注册 API 路由
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
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
			comments.PUT("/:id", handleUpdateComment)
			comments.DELETE("/:id", handleDeleteComment)
		}

		// 系统设置（仅管理员）
		settings := api.Group("/settings")
		settings.Use(adminMiddleware())
		{
			settings.GET("", handleGetSettings)
			settings.PUT("", handleUpdateSettings)
		}
	}
}

// authMiddleware JWT 认证中间件
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

		userID, err := userService.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
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

		userID, err := userService.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
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

	bibis, total, err := bibiService.GetBibis(page, pageSize, visibility)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取笔记列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bibis": bibis,
		"total": total,
		"page":  page,
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

	bibi, err := bibiService.GetBibiByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleUpdateBibi 处理更新笔记请求
func handleUpdateBibi(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

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

	bibi, err := bibiService.UpdateBibi(uint(id), req.Content, req.Visibility, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新笔记失败"})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleDeleteBibi 处理删除笔记请求
func handleDeleteBibi(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

	if err := bibiService.DeleteBibi(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除笔记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleTogglePin 处理切换置顶状态请求
func handleTogglePin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

	bibi, err := bibiService.TogglePin(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换置顶状态失败"})
		return
	}

	c.JSON(http.StatusOK, bibi)
}

// handleToggleLike 处理切换点赞状态请求
func handleToggleLike(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

	userID := c.GetUint("userID")

	liked, err := likeService.ToggleLike(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"liked": liked})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败"})
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

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	tag, err := tagService.UpdateTag(uint(id), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
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

	if err := tagService.DeleteTag(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleGetComments 处理获取评论列表请求
func handleGetComments(c *gin.Context) {
	bibiID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := commentService.GetCommentsByBibiID(uint(bibiID), page, pageSize)
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
	bibiID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的笔记 ID"})
		return
	}

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

	comment, err := commentService.CreateComment(uint(bibiID), req.Name, req.Email, req.Website, req.Content, req.ParentID)
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

	if err := commentService.DeleteComment(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// handleGetSettings 获取系统设置
func handleGetSettings(c *gin.Context) {
	registrationEnabled, _ := systemService.GetSetting("registration_enabled")
	c.JSON(http.StatusOK, gin.H{
		"registration_enabled": registrationEnabled == "true",
	})
}

// handleUpdateSettings 更新系统设置
func handleUpdateSettings(c *gin.Context) {
	var req struct {
		RegistrationEnabled *bool `json:"registration_enabled"`
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

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
