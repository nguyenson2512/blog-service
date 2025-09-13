package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/blog-service/internal/cache"
	"github.com/example/blog-service/internal/config"
	"github.com/example/blog-service/internal/db"
	"github.com/example/blog-service/internal/models"
	"github.com/example/blog-service/internal/search"
	"github.com/example/blog-service/internal/transport/http"
)

type Application struct {
	Config *config.Config
	DB     *db.Database
	Cache  *cache.RedisClient
	Search *search.Elastic
	Router http.Router
}

func Initialize() (*Application, error) {
	cfg := config.Load()

	database, err := db.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	if err := database.AutoMigrate(&models.Post{}, &models.ActivityLog{}); err != nil {
		return nil, fmt.Errorf("db migrate: %w", err)
	}
	if err := database.EnsureGINIndexOnTags(); err != nil {
		return nil, fmt.Errorf("ensure GIN index: %w", err)
	}

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	es, err := search.NewElastic(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := es.EnsurePostsIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure ES index: %w", err)
	}

	r := http.NewRouter(cfg, database, redisClient, es)

	return &Application{
		Config: cfg,
		DB:     database,
		Cache:  redisClient,
		Search: es,
		Router: r,
	}, nil
}

func (a *Application) Close() {
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}
	if a.Cache != nil {
		_ = a.Cache.Close()
	}
} 