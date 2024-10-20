package models

type ImageOptimizationRequest struct {
	Url     string                            `json:"url" validate:"url"`
	Options []ImageOptimizationRequestOptions `json:"options" validate:"dive"`
}

type ImageOptimizationRequestOptions struct {
	Width         int            `json:"width" validate:"number"`
	Height        int            `json:"height" validate:"number"`
	Format        string         `json:"format" validate:"oneof=webp"`
	FormatOptions map[string]any `json:"formatOptions" validate:"omitempty"`
	BucketId      string         `json:"bucketId" validate:"required"`
	Key           string         `json:"key" validate:"required,filename"`
}
