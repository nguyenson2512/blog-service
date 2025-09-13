package http

import (
	"github.com/gin-gonic/gin"

	"github.com/example/blog-service/internal/cache"
	"github.com/example/blog-service/internal/config"
	"github.com/example/blog-service/internal/db"
	"github.com/example/blog-service/internal/search"
	"github.com/example/blog-service/internal/transport/http/handlers"
)

type Router = *gin.Engine

func NewRouter(cfg *config.Config, database *db.Database, cache *cache.RedisClient, es *search.Elastic) Router {
	if mode := gin.Mode(); mode == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	h := handlers.NewPostHandler(database, cache, es)

	r.POST("/posts", h.CreatePost)
	r.GET("/posts/:id", h.GetPost)
	r.PUT("/posts/:id", h.UpdatePost)
	r.GET("/posts/search-by-tag", h.SearchByTag)
	r.GET("/posts/search", h.Search)

	return r
} 