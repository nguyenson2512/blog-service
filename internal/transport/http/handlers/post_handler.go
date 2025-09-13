package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/blog-service/internal/cache"
	"github.com/example/blog-service/internal/db"
	"github.com/example/blog-service/internal/search"
	"github.com/example/blog-service/internal/service"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(database *db.Database, cache *cache.RedisClient, es *search.Elastic) *PostHandler {
	return &PostHandler{service: service.NewPostService(database, cache, es)}
}

type createReq struct {
	Title   string   `json:"title" binding:"required,min=1"`
	Content string   `json:"content" binding:"required,min=1"`
	Tags    []string `json:"tags"`
}

type updateReq struct {
	Title   string   `json:"title" binding:"required,min=1"`
	Content string   `json:"content" binding:"required,min=1"`
	Tags    []string `json:"tags"`
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post, err := h.service.CreatePost(c.Request.Context(), service.CreatePostInput{Title: req.Title, Content: req.Content, Tags: req.Tags})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	
	// Check if related posts are requested
	includeRelated := c.Query("include_related") == "true"
	
	if includeRelated {
		postWithRelated, err := h.service.GetPostWithRelated(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, postWithRelated)
	} else {
		post, err := h.service.GetPost(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, post)
	}
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post, err := h.service.UpdatePost(c.Request.Context(), uint(id), service.UpdatePostInput{Title: req.Title, Content: req.Content, Tags: req.Tags})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) SearchByTag(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag is required"})
		return
	}
	posts, err := h.service.SearchByTag(c.Request.Context(), tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}
	res, err := h.service.SearchES(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
