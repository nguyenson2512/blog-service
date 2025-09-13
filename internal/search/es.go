package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/example/blog-service/internal/config"
)

type Elastic struct {
	Client *elasticsearch.Client
	Index  string
}

func NewElastic(cfg *config.Config) (*Elastic, error) {
	cfgES := elasticsearch.Config{
		Addresses: []string{cfg.ElasticAddr},
	}
	if cfg.ElasticUsername != "" {
		cfgES.Username = cfg.ElasticUsername
		cfgES.Password = cfg.ElasticPassword
	}
	client, err := elasticsearch.NewClient(cfgES)
	if err != nil {
		return nil, err
	}
	return &Elastic{Client: client, Index: "posts"}, nil
}

func (e *Elastic) EnsurePostsIndex(ctx context.Context) error {
	res, err := e.Client.Indices.Exists([]string{e.Index})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK { return nil }

	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title":   map[string]string{"type": "text"},
				"content": map[string]string{"type": "text"},
				"tags":    map[string]string{"type": "keyword"},
			},
		},
	}
	b, _ := json.Marshal(mapping)
	createRes, err := e.Client.Indices.Create(e.Index, e.Client.Indices.Create.WithBody(bytes.NewReader(b)))
	if err != nil { return err }
	defer createRes.Body.Close()
	if createRes.IsError() { return fmt.Errorf("failed to create index: %s", createRes.String()) }
	return nil
}

func (e *Elastic) IndexPost(ctx context.Context, id uint, doc map[string]interface{}) error {
	b, _ := json.Marshal(doc)
	req := esapi.IndexRequest{Index: e.Index, DocumentID: fmt.Sprintf("%d", id), Body: bytes.NewReader(b), Refresh: "true"}
	res, err := req.Do(ctx, e.Client)
	if err != nil { return err }
	defer res.Body.Close()
	if res.IsError() { return fmt.Errorf("index error: %s", res.String()) }
	return nil
}

func (e *Elastic) SearchPosts(ctx context.Context, query string) ([]map[string]interface{}, error) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "content"},
			},
		},
	}
	b, _ := json.Marshal(body)
	res, err := e.Client.Search(e.Client.Search.WithContext(ctx), e.Client.Search.WithIndex(e.Index), e.Client.Search.WithBody(strings.NewReader(string(b))), e.Client.Search.WithTrackTotalHits(true), e.Client.Search.WithTimeout(10*time.Second))
	if err != nil { return nil, err }
	defer res.Body.Close()
	if res.IsError() { return nil, fmt.Errorf("search error: %s", res.String()) }
	var parsed map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil { return nil, err }
	var results []map[string]interface{}
	hits := parsed["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, h := range hits {
		src := h.(map[string]interface{})["_source"].(map[string]interface{})
		results = append(results, src)
	}
	return results, nil
}

func (e *Elastic) FindRelatedPosts(ctx context.Context, postID uint, tags []string, limit int) ([]map[string]interface{}, error) {
	if len(tags) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Build should clauses for each tag
	var shouldClauses []map[string]interface{}
	for _, tag := range tags {
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"tags": tag,
			},
		})
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": shouldClauses,
				"must_not": map[string]interface{}{
					"term": map[string]interface{}{
						"id": postID,
					},
				},
				"minimum_should_match": 1,
			},
		},
		"size": limit,
	}

	b, _ := json.Marshal(body)
	res, err := e.Client.Search(
		e.Client.Search.WithContext(ctx),
		e.Client.Search.WithIndex(e.Index),
		e.Client.Search.WithBody(strings.NewReader(string(b))),
		e.Client.Search.WithTrackTotalHits(true),
		e.Client.Search.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("related posts search error: %s", res.String())
	}

	var parsed map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	hits := parsed["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, h := range hits {
		src := h.(map[string]interface{})["_source"].(map[string]interface{})
		results = append(results, src)
	}
	return results, nil
}
