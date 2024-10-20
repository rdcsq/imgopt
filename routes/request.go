package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"imgopt/models"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-playground/validator/v10"
)

type ImageOptimizationRequestWrapper struct {
	Config    models.ApiConfig
	Validate  validator.Validate
	S3Clients map[string]*s3.Client
}

func (iow *ImageOptimizationRequestWrapper) Handler(w http.ResponseWriter, r *http.Request) {
	// parse json
	var request models.ImageOptimizationRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		ErrorResponse(w, 400, "Bad Request")
		return
	}

	// validate
	err = iow.Validate.Struct(request)
	if err != nil {
		ErrorResponse(w, 400, err.Error())
		return
	}

	// check if bucket exists
	for _, options := range request.Options {
		bucket, ok := iow.Config.RegisteredS3Buckets[options.BucketId]
		if !ok {
			ErrorResponse(w, 400, "Bucket not found")
			return
		}

		// check if alias is in allowlist
		alias := r.Header.Get("X-Identity")
		allowed := false
		for _, allowedAlias := range bucket.AllowedAliases {
			if alias == allowedAlias {
				allowed = true
				break
			}
		}

		if !allowed {
			ErrorResponse(w, 403, "Forbidden")
			return
		}
	}

	// try to download image
	resp, err := http.Get(request.Url)
	if err != nil {
		ErrorResponse(w, 500, "Failed to download image")
		return
	}
	defer resp.Body.Close()

	if resp.ContentLength > 15*1024*1024 {
		ErrorResponse(w, 400, "Image size exceeds 15 MB")
		return
	}

	// process the image
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		ErrorResponse(w, 500, "Failed to read image data")
		return
	}

	var wg sync.WaitGroup
	results := make([]bool, len(request.Options))

	for i, options := range request.Options {
		wg.Add(1)
		go func(url string, options models.ImageOptimizationRequestOptions) {
			defer wg.Done()
			results[i] = iow.generateImage(imageData, options)
		}(request.Url, options)
	}

	wg.Wait()
	json.NewEncoder(w).Encode(map[string]any{"status": "completed", "results": results})
}

func getWebpOptions(opts map[string]any) (*vips.WebpExportParams, error) {
	ep := vips.NewWebpExportParams()

	if opts != nil {
		formatOptionsBytes, err := json.Marshal(opts)
		if err != nil {
			return nil, err
		}

		var formatOptions models.WebpConfig
		err = json.Unmarshal(formatOptionsBytes, &formatOptions)
		if err != nil {
			return nil, err
		}
		ep.Quality = formatOptions.Quality
		ep.Lossless = formatOptions.Lossless
	}

	return ep, nil
}

func (iow *ImageOptimizationRequestWrapper) generateImage(imageData []byte, options models.ImageOptimizationRequestOptions) bool {
	image, err := vips.NewThumbnailWithSizeFromBuffer(imageData, options.Size, options.Size, vips.InterestingNone, ParseResizeMode(options.ResizeMode))
	if err != nil {
		return false
	}
	image.RemoveMetadata()

	// this should be the only available option
	if options.Format == "webp" {
		opts, err := getWebpOptions(options.FormatOptions)
		if err != nil {
			return false
		}

		imageData, _, err := image.ExportWebp(opts)
		if err != nil {
			return false
		}

		_, err = iow.S3Clients[options.BucketId].PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: &options.BucketId,
			Key:    &options.Key,
			Body:   bytes.NewReader(imageData),
		})

		return err == nil
	}

	log.Println("somehow got to an unsupported format.")
	return false
}

func ParseResizeMode(mode string) vips.Size {
	switch mode {
	case "up":
		return vips.SizeUp
	case "down":
		return vips.SizeDown
	case "force":
		return vips.SizeForce
	default:
		return vips.SizeBoth
	}
}

func ErrorResponse(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": message})
}
