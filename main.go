package main

import (
	"context"
	"encoding/json"
	"imgopt/customvalidators"
	"imgopt/middleware"
	"imgopt/models"
	"imgopt/routes"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-playground/validator/v10"
)

func main() {
	config := readConfig()
	log.Println("loaded config file")

	s3Clients := createS3Clients(config)
	log.Println("created s3 clients")

	vips.Startup(nil)
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	defer vips.Shutdown()
	log.Println("loaded lipvips")

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("filename", customvalidators.Filename)

	imageOptimizationRequestWrapper := routes.ImageOptimizationRequestWrapper{
		Config:    config,
		Validate:  *validate,
		S3Clients: s3Clients,
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /heartbeat", routes.Heartbeat)
	router.HandleFunc("POST /r", imageOptimizationRequestWrapper.Handler)

	authMiddleware := middleware.AuthWrapper{AllowedTokens: config.AllowedTokens}
	middlewareStack := middleware.CreateStack(middleware.Logging, middleware.JsonContentType, authMiddleware.Auth)
	server := http.Server{
		Addr:    config.ListenAddress,
		Handler: middlewareStack(router),
	}

	log.Printf("starting server on %s\n", config.ListenAddress)
	err := server.ListenAndServe()
	if err != nil {
		panic("error starting server. perhaps the listening address is already in use.")
	}
}

func readConfig() models.ApiConfig {
	configString, err := os.ReadFile("config.json")
	if err != nil {
		panic("Unable to read config file.")
	}

	var config models.ApiConfig
	err = json.Unmarshal(configString, &config)
	if err != nil {
		panic("Error parsing config file.")
	}

	return config
}

func createS3Clients(config models.ApiConfig) map[string]*s3.Client {
	s3Clients := make(map[string]*s3.Client, len(config.RegisteredS3Buckets))

	for id, bucket := range config.RegisteredS3Buckets {
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(bucket.AccessKey, bucket.SecretKey, "")),
			awsconfig.WithRegion(bucket.Region),
		)
		if err != nil {
			panic("Error creating s3 client")
		}

		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			if bucket.Endpoint != "" {
				o.BaseEndpoint = aws.String(bucket.Endpoint)
			}
			o.UsePathStyle = bucket.UsePathStyle
		})
		s3Clients[id] = client
	}

	return s3Clients
}
