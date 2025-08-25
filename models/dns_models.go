package models

type Zone struct {
	Name    string           `json:"name" binding:"required"`
	TTL     uint32           `json:"ttl,omitempty"`
	Records []ResourceRecord `json:"records,omitempty"`
}

type ResourceRecord struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Class  string            `json:"class"`
	TTL    uint32            `json:"ttl"`
	Fields map[string]string `json:"fields"`
}

type Record struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   int    `json:"ttl"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

type ZoneListResponse struct {
	Success bool     `json:"success"`
	Zones   []string `json:"zones"`
	Count   int      `json:"count"`
	Error   string   `json:"error,omitempty"`
}
