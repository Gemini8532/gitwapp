package api

import (
	"github.com/Gemini8532/gitwapp/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handler) {
	api := r.Group("/api")
	{
		api.POST("/login", h.Login)

		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.GET("/repos", h.GetRepos)
			authorized.POST("/repos", h.CreateRepo)
			authorized.DELETE("/repos/:id", h.DeleteRepo)

			authorized.GET("/repos/:id/status", h.GetRepoStatus)
		}
	}
}
