package service

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/example/blog-service/internal/cache"
	"github.com/example/blog-service/internal/db"
	"github.com/example/blog-service/internal/models"
	"github.com/example/blog-service/internal/repository"
	"github.com/example/blog-service/internal/search"
)

type PostService struct {
	db     *db.Database
	cache  *cache.RedisClient
	es     *search.Elastic
	repo   *repository.PostRepository
}

func NewPostService(database *db.Database, cache *cache.RedisClient, es *search.Elastic) *PostService {
	return &PostService{
		db:    database,
		cache: cache,
		es:   es,
		repo: repository.NewPostRepository(database.Gorm),
	}
}

type CreatePostInput struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type UpdatePostInput struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (s *PostService) CreatePost(ctx context.Context, in CreatePostInput) (*models.Post, error) {
	post := &models.Post{Title: in.Title, Content: in.Content, Tags: pq.StringArray(in.Tags)}
	var created *models.Post
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(ctx, tx, post); err != nil { return err }
		if err := s.repo.LogActivity(ctx, tx, "new_post", post.ID); err != nil { return err }
		created = post
		return nil
	})
	if err != nil { return nil, err }
	_ = s.es.IndexPost(ctx, created.ID, map[string]interface{}{ "id": created.ID, "title": created.Title, "content": created.Content })
	return created, nil
}

func (s *PostService) GetPost(ctx context.Context, id uint) (*models.Post, error) {
	key := fmt.Sprintf("post:%d", id)
	var post models.Post
	if found, err := s.cache.GetJSON(ctx, key, &post); err == nil && found {
		return &post, nil
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	_ = s.cache.SetJSON(ctx, key, p)
	return p, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id uint, in UpdatePostInput) (*models.Post, error) {
	post := &models.Post{ID: id, Title: in.Title, Content: in.Content, Tags: pq.StringArray(in.Tags)}
	if err := s.repo.Update(ctx, post); err != nil { return nil, err }
	_ = s.cache.Del(ctx, fmt.Sprintf("post:%d", id))
	_ = s.es.IndexPost(ctx, id, map[string]interface{}{ "id": id, "title": post.Title, "content": post.Content })
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) SearchByTag(ctx context.Context, tag string) ([]models.Post, error) {
	return s.repo.SearchByTag(ctx, tag)
}

func (s *PostService) SearchES(ctx context.Context, q string) ([]map[string]interface{}, error) {
	return s.es.SearchPosts(ctx, q)
} 