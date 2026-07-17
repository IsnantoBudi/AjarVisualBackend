package models

import "time"

type CachedImage struct {
	PromptHash  string    `gorm:"primaryKey;column:prompt_hash;type:varchar(64)" json:"prompt_hash"`
	Prompt      string    `gorm:"column:prompt;type:text;not null" json:"prompt"`
	ImageData   []byte    `gorm:"column:image_data;type:longblob;not null" json:"image_data"`
	ContentType string    `gorm:"column:content_type;type:varchar(50);not null" json:"content_type"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (CachedImage) TableName() string {
	return "cached_images"
}
