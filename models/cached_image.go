package models

import "time"

type CachedImage struct {
	PromptHash  string     `json:"prompt_hash"`
	Prompt      string     `json:"prompt"`
	ImageData   []byte     `json:"image_data"`
	ContentType string     `json:"content_type"`
	CreatedAt   *time.Time `json:"created_at"`
}
