package model

// Error represents an HTTP error status
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
