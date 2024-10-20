package models

type ApiConfig struct {
	ListenAddress       string              `json:"listenAddress"`
	AllowedTokens       []AllowedToken      `json:"allowedTokens"`
	DefaultFormatConfig DefaultFormatConfig `json:"defaultFormatConfig"`
	RegisteredS3Buckets map[string]S3Bucket `json:"registeredS3Buckets"`
}

type AllowedToken struct {
	Alias string `json:"alias"`
	Token string `json:"token"`
}

type DefaultFormatConfig struct {
	Webp WebpConfig `json:"webp"`
}

type WebpConfig struct {
	Quality  int  `json:"quality"`
	Lossless bool `json:"lossless"`
}

type S3Bucket struct {
	Name           string   `json:"name"`
	Region         string   `json:"region"`
	AccessKey      string   `json:"accessKey"`
	SecretKey      string   `json:"secretKey"`
	Endpoint       string   `json:"endpoint"`
	AllowedAliases []string `json:"allowedAliases"`
	UsePathStyle   bool     `json:"usePathStyle"`
}
