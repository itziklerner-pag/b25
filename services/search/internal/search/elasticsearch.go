package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/search/internal/config"
	"github.com/yourorg/b25/services/search/pkg/models"
)

// ElasticsearchClient wraps the Elasticsearch client
type ElasticsearchClient struct {
	client *elasticsearch.Client
	config *config.ElasticsearchConfig
	logger *zap.Logger
}

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient(cfg *config.ElasticsearchConfig, logger *zap.Logger) (*ElasticsearchClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		APIKey:    cfg.APIKey,
		MaxRetries: cfg.MaxRetries,
		RetryBackoff: func(i int) time.Duration {
			return cfg.RetryBackoff * time.Duration(i)
		},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch ping failed: %s", res.String())
	}

	logger.Info("Connected to Elasticsearch", zap.Strings("addresses", cfg.Addresses))

	return &ElasticsearchClient{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// Search performs a search query
func (es *ElasticsearchClient) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	startTime := time.Now()

	// Build query
	query := es.buildQuery(req)

	// Serialize query
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	// Determine indices to search
	indices := es.getIndices(req.Index)

	// Execute search
	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(indices...),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s - %s", res.Status(), string(body))
	}

	// Parse response
	var esResp elasticsearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to API response
	response := &models.SearchResponse{
		Query:     req.Query,
		TotalHits: esResp.Hits.Total.Value,
		MaxScore:  esResp.Hits.MaxScore,
		Results:   make([]models.SearchResult, 0, len(esResp.Hits.Hits)),
		Took:      esResp.Took,
		From:      req.From,
		Size:      req.Size,
	}

	for _, hit := range esResp.Hits.Hits {
		result := models.SearchResult{
			Index:      hit.Index,
			ID:         hit.ID,
			Score:      hit.Score,
			Source:     hit.Source,
			Highlights: hit.Highlight,
		}
		response.Results = append(response.Results, result)
	}

	// Add facets if requested
	if len(req.Facets) > 0 && esResp.Aggregations != nil {
		response.Facets = es.parseFacets(esResp.Aggregations)
	}

	es.logger.Debug("Search completed",
		zap.String("query", req.Query),
		zap.Int64("total_hits", response.TotalHits),
		zap.Duration("latency", time.Since(startTime)),
	)

	return response, nil
}

// Autocomplete performs autocomplete suggestions
func (es *ElasticsearchClient) Autocomplete(ctx context.Context, req *models.AutocompleteRequest) (*models.AutocompleteResponse, error) {
	startTime := time.Now()

	// Build autocomplete query
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"autocomplete": map[string]interface{}{
				"prefix": req.Query,
				"completion": map[string]interface{}{
					"field": es.getAutocompleteField(req.Field),
					"size":  req.Size,
					"skip_duplicates": true,
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	indices := es.getIndices(req.Index)

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(indices...),
		es.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("autocomplete request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("autocomplete error: %s - %s", res.Status(), string(body))
	}

	var esResp elasticsearchSuggestResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	suggestions := make([]string, 0)
	if suggest, ok := esResp.Suggest["autocomplete"]; ok && len(suggest) > 0 {
		for _, option := range suggest[0].Options {
			suggestions = append(suggestions, option.Text)
		}
	}

	response := &models.AutocompleteResponse{
		Query:       req.Query,
		Suggestions: suggestions,
		Took:        esResp.Took,
	}

	es.logger.Debug("Autocomplete completed",
		zap.String("query", req.Query),
		zap.Int("suggestions", len(suggestions)),
		zap.Duration("latency", time.Since(startTime)),
	)

	return response, nil
}

// Index indexes a single document
func (es *ElasticsearchClient) Index(ctx context.Context, req *models.IndexRequest) (*models.IndexResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Document); err != nil {
		return nil, fmt.Errorf("failed to encode document: %w", err)
	}

	indexName := es.getIndexName(req.Index)

	var res *esapi.Response
	var err error

	if req.ID != "" {
		res, err = es.client.Index(
			indexName,
			&buf,
			es.client.Index.WithContext(ctx),
			es.client.Index.WithDocumentID(req.ID),
			es.client.Index.WithRefresh("true"),
		)
	} else {
		res, err = es.client.Index(
			indexName,
			&buf,
			es.client.Index.WithContext(ctx),
			es.client.Index.WithRefresh("true"),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("index request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return &models.IndexResponse{
			Success: false,
			Error:   fmt.Sprintf("index error: %s - %s", res.Status(), string(body)),
		}, nil
	}

	var esResp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	id := ""
	if idVal, ok := esResp["_id"].(string); ok {
		id = idVal
	}

	return &models.IndexResponse{
		Success: true,
		ID:      id,
		Index:   indexName,
	}, nil
}

// BulkIndex indexes multiple documents
func (es *ElasticsearchClient) BulkIndex(ctx context.Context, req *models.BulkIndexRequest) (*models.BulkIndexResponse, error) {
	startTime := time.Now()

	var buf bytes.Buffer
	for _, doc := range req.Documents {
		indexName := es.getIndexName(doc.Index)

		// Action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		if doc.ID != "" {
			action["index"].(map[string]interface{})["_id"] = doc.ID
		}

		if err := json.NewEncoder(&buf).Encode(action); err != nil {
			return nil, fmt.Errorf("failed to encode action: %w", err)
		}

		// Document line
		if err := json.NewEncoder(&buf).Encode(doc.Document); err != nil {
			return nil, fmt.Errorf("failed to encode document: %w", err)
		}
	}

	res, err := es.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		es.client.Bulk.WithContext(ctx),
		es.client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bulk error: %s - %s", res.Status(), string(body))
	}

	var esResp elasticsearchBulkResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := &models.BulkIndexResponse{
		Success: !esResp.Errors,
		Indexed: 0,
		Failed:  0,
		Took:    esResp.Took,
		Errors:  make([]models.IndexResponse, 0),
	}

	for _, item := range esResp.Items {
		if indexResult, ok := item["index"]; ok {
			if indexResult.Error.Type != "" {
				response.Failed++
				response.Errors = append(response.Errors, models.IndexResponse{
					Success: false,
					ID:      indexResult.ID,
					Index:   indexResult.Index,
					Error:   fmt.Sprintf("%s: %s", indexResult.Error.Type, indexResult.Error.Reason),
				})
			} else {
				response.Indexed++
			}
		}
	}

	es.logger.Info("Bulk indexing completed",
		zap.Int("indexed", response.Indexed),
		zap.Int("failed", response.Failed),
		zap.Duration("latency", time.Since(startTime)),
	)

	return response, nil
}

// CreateIndices creates all required indices
func (es *ElasticsearchClient) CreateIndices(ctx context.Context) error {
	for indexType, indexCfg := range es.config.Indices {
		mapping := es.getIndexMapping(indexType)

		settings := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   indexCfg.Shards,
				"number_of_replicas": indexCfg.Replicas,
			},
			"mappings": mapping,
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(settings); err != nil {
			return fmt.Errorf("failed to encode settings: %w", err)
		}

		res, err := es.client.Indices.Create(
			indexCfg.Name,
			es.client.Indices.Create.WithContext(ctx),
			es.client.Indices.Create.WithBody(&buf),
		)
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexCfg.Name, err)
		}
		defer res.Body.Close()

		if res.IsError() {
			// Ignore if index already exists
			if !strings.Contains(res.String(), "resource_already_exists_exception") {
				body, _ := io.ReadAll(res.Body)
				return fmt.Errorf("create index error: %s - %s", res.Status(), string(body))
			}
			es.logger.Info("Index already exists", zap.String("index", indexCfg.Name))
		} else {
			es.logger.Info("Created index", zap.String("index", indexCfg.Name))
		}
	}

	return nil
}

// Health checks Elasticsearch health
func (es *ElasticsearchClient) Health(ctx context.Context) (*models.ComponentHealth, error) {
	startTime := time.Now()

	res, err := es.client.Cluster.Health(
		es.client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return &models.ComponentHealth{
			Status: "unhealthy",
			Error:  err.Error(),
		}, nil
	}
	defer res.Body.Close()

	latency := time.Since(startTime)

	if res.IsError() {
		return &models.ComponentHealth{
			Status:  "unhealthy",
			Latency: latency.String(),
			Error:   res.String(),
		}, nil
	}

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return &models.ComponentHealth{
			Status:  "degraded",
			Latency: latency.String(),
			Error:   "failed to parse health response",
		}, nil
	}

	status := "healthy"
	if clusterStatus, ok := health["status"].(string); ok {
		if clusterStatus == "yellow" {
			status = "degraded"
		} else if clusterStatus == "red" {
			status = "unhealthy"
		}
	}

	return &models.ComponentHealth{
		Status:  status,
		Latency: latency.String(),
	}, nil
}

// Helper functions

func (es *ElasticsearchClient) buildQuery(req *models.SearchRequest) map[string]interface{} {
	query := map[string]interface{}{
		"from": req.From,
		"size": req.Size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  req.Query,
							"fields": []string{"*"},
							"type":   "best_fields",
						},
					},
				},
			},
		},
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

	// Add filters
	if len(req.Filters) > 0 {
		filters := make([]interface{}, 0)
		for field, value := range req.Filters {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{
					field: value,
				},
			})
		}
		boolQuery["filter"] = filters
	}

	// Add date range filter
	if req.DateRange != nil {
		rangeFilter := map[string]interface{}{
			"range": map[string]interface{}{
				req.DateRange.Field: map[string]interface{}{},
			},
		}
		rangeMap := rangeFilter["range"].(map[string]interface{})[req.DateRange.Field].(map[string]interface{})
		if req.DateRange.From != nil {
			rangeMap["gte"] = req.DateRange.From.Format(time.RFC3339)
		}
		if req.DateRange.To != nil {
			rangeMap["lte"] = req.DateRange.To.Format(time.RFC3339)
		}

		if filters, ok := boolQuery["filter"].([]interface{}); ok {
			boolQuery["filter"] = append(filters, rangeFilter)
		} else {
			boolQuery["filter"] = []interface{}{rangeFilter}
		}
	}

	// Add minimum score
	if req.MinScore != nil {
		query["min_score"] = *req.MinScore
	}

	// Add sorting
	if len(req.Sort) > 0 {
		sort := make([]interface{}, 0)
		for _, s := range req.Sort {
			sort = append(sort, map[string]interface{}{
				s.Field: map[string]interface{}{
					"order": s.Order,
				},
			})
		}
		query["sort"] = sort
	}

	// Add highlighting
	if req.Highlight {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"*": map[string]interface{}{},
			},
		}
	}

	// Add facets (aggregations)
	if len(req.Facets) > 0 {
		aggs := make(map[string]interface{})
		for _, field := range req.Facets {
			aggs[field] = map[string]interface{}{
				"terms": map[string]interface{}{
					"field": field + ".keyword",
					"size":  100,
				},
			}
		}
		query["aggs"] = aggs
	}

	return query
}

func (es *ElasticsearchClient) getIndices(index string) []string {
	if index == "" {
		// Search all indices
		indices := make([]string, 0, len(es.config.Indices))
		for _, cfg := range es.config.Indices {
			indices = append(indices, cfg.Name)
		}
		return indices
	}

	// Search specific index
	if cfg, ok := es.config.Indices[index]; ok {
		return []string{cfg.Name}
	}

	return []string{index}
}

func (es *ElasticsearchClient) getIndexName(indexType string) string {
	if cfg, ok := es.config.Indices[indexType]; ok {
		return cfg.Name
	}
	return indexType
}

func (es *ElasticsearchClient) getAutocompleteField(field string) string {
	if field == "" {
		return "suggest"
	}
	return field
}

func (es *ElasticsearchClient) parseFacets(aggs map[string]interface{}) map[string][]models.Facet {
	facets := make(map[string][]models.Facet)

	for field, aggData := range aggs {
		if aggMap, ok := aggData.(map[string]interface{}); ok {
			if buckets, ok := aggMap["buckets"].([]interface{}); ok {
				fieldFacets := make([]models.Facet, 0, len(buckets))
				for _, bucket := range buckets {
					if b, ok := bucket.(map[string]interface{}); ok {
						facet := models.Facet{
							Key:   b["key"],
							Count: int64(b["doc_count"].(float64)),
						}
						fieldFacets = append(fieldFacets, facet)
					}
				}
				facets[field] = fieldFacets
			}
		}
	}

	return facets
}

func (es *ElasticsearchClient) getIndexMapping(indexType string) map[string]interface{} {
	// Return appropriate mapping based on index type
	// This is a basic mapping; in production, you'd have more specific mappings
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"type": "date",
			},
			"suggest": map[string]interface{}{
				"type": "completion",
			},
		},
	}
}

// Elasticsearch response types

type elasticsearchResponse struct {
	Took         int64                  `json:"took"`
	TimedOut     bool                   `json:"timed_out"`
	Hits         hitsResponse           `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

type hitsResponse struct {
	Total    totalHits   `json:"total"`
	MaxScore float64     `json:"max_score"`
	Hits     []hitResult `json:"hits"`
}

type totalHits struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"`
}

type hitResult struct {
	Index     string                 `json:"_index"`
	ID        string                 `json:"_id"`
	Score     float64                `json:"_score"`
	Source    map[string]interface{} `json:"_source"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

type elasticsearchSuggestResponse struct {
	Took    int64                             `json:"took"`
	Suggest map[string][]suggestionResult     `json:"suggest"`
}

type suggestionResult struct {
	Text    string           `json:"text"`
	Offset  int              `json:"offset"`
	Length  int              `json:"length"`
	Options []suggestionOption `json:"options"`
}

type suggestionOption struct {
	Text  string  `json:"text"`
	Score float64 `json:"_score"`
}

type elasticsearchBulkResponse struct {
	Took   int64                      `json:"took"`
	Errors bool                       `json:"errors"`
	Items  []map[string]bulkItemResult `json:"items"`
}

type bulkItemResult struct {
	Index   string          `json:"_index"`
	ID      string          `json:"_id"`
	Version int             `json:"_version"`
	Result  string          `json:"result"`
	Status  int             `json:"status"`
	Error   bulkItemError   `json:"error"`
}

type bulkItemError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}
