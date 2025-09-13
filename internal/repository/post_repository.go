package repository

import (
	"context"

	"github.com/example/blog-service/internal/models"
	"gorm.io/gorm"
)

type PostRepository struct{ db *gorm.DB }

func NewPostRepository(db *gorm.DB) *PostRepository { return &PostRepository{db: db} }

func (r *PostRepository) Create(ctx context.Context, tx *gorm.DB, p *models.Post) error {
	return tx.WithContext(ctx).Create(p).Error
}

func (r *PostRepository) Update(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"title":   p.Title,
		"content": p.Content,
		"tags":    p.Tags,
	}).Error
}

func (r *PostRepository) GetByID(ctx context.Context, id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil { return nil, err }
	return &post, nil
}

func (r *PostRepository) SearchByTag(ctx context.Context, tag string) ([]models.Post, error) {
	var posts []models.Post
	// tags @> ARRAY[tag]::text[] uses GIN index
	if err := r.db.WithContext(ctx).Where("tags @> ARRAY[?]::text[]", tag).Order("id DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) LogActivity(ctx context.Context, tx *gorm.DB, action string, postID uint) error {
	log := models.ActivityLog{Action: action, PostID: postID}
	return tx.WithContext(ctx).Create(&log).Error
} 