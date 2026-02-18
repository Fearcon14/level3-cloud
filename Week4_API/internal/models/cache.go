package models

// SetCacheRequest is the body for POST /api/v1/instances/:id/cache.
type SetCacheRequest struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"` // optional; 0 means no expiry
}

// GetCacheResponse is the response for GET /api/v1/instances/:id/cache/:key.
type GetCacheResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
