package models

// RedisInstance represents a Redis instance
type RedisInstance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Capacity  string `json:"capacity"`
}

// CreateRedisRequest validates what the user sends us
type CreateRedisRequest struct {
	Name     string `json:"name" validate:"required"`
	Capacity string `json:"capacity" validate:"required"`
}
