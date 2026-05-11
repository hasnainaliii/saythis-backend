package usecase

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ImageUploader is the interface used by UserUseCase to upload avatar images.
// Keeping it as an interface makes the uploader swappable in tests.
type ImageUploader interface {
	Upload(ctx context.Context, file io.Reader, filename string) (secureURL string, err error)
}

// CloudinaryUploader uploads images to Cloudinary using the signed upload API.
// Images are stored in the "saythis" folder inside the configured cloud.
type CloudinaryUploader struct {
	cloudName  string
	apiKey     string
	apiSecret  string
	folder     string
	httpClient *http.Client
}

// NewCloudinaryUploader parses the CLOUDINARY_URL and returns a configured uploader.
// Expected format: cloudinary://api_key:api_secret@cloud_name
func NewCloudinaryUploader(cloudinaryURL string) (*CloudinaryUploader, error) {
	u, err := url.Parse(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("parse cloudinary url: %w", err)
	}
	apiKey := u.User.Username()
	apiSecret, _ := u.User.Password()
	cloudName := u.Host

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary url must be cloudinary://api_key:api_secret@cloud_name")
	}

	return &CloudinaryUploader{
		cloudName:  cloudName,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		folder:     "saythis",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// MustNewCloudinaryUploader is like NewCloudinaryUploader but panics on error.
// Use at startup where a misconfigured URL is an unrecoverable fault.
func MustNewCloudinaryUploader(cloudinaryURL string) *CloudinaryUploader {
	u, err := NewCloudinaryUploader(cloudinaryURL)
	if err != nil {
		panic("cloudinary: " + err.Error())
	}
	return u
}

type cloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Upload sends the file to Cloudinary via the signed upload API and returns its secure URL.
// Signature algorithm: SHA-1(folder={folder}&timestamp={ts}{api_secret})
func (c *CloudinaryUploader) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Parameters must be sorted alphabetically (f before t).
	toSign := fmt.Sprintf("folder=%s&timestamp=%s%s", c.folder, timestamp, c.apiSecret)
	h := sha1.New()
	h.Write([]byte(toSign))
	signature := fmt.Sprintf("%x", h.Sum(nil))

	// Build the multipart body.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("cloudinary: create file field: %w", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		return "", fmt.Errorf("cloudinary: write file data: %w", err)
	}
	_ = mw.WriteField("api_key", c.apiKey)
	_ = mw.WriteField("timestamp", timestamp)
	_ = mw.WriteField("signature", signature)
	_ = mw.WriteField("folder", c.folder)
	mw.Close()

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", c.cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &body)
	if err != nil {
		return "", fmt.Errorf("cloudinary: build request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudinary: http request: %w", err)
	}
	defer resp.Body.Close()

	var result cloudinaryUploadResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("cloudinary: decode response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("cloudinary: %s", result.Error.Message)
	}

	return result.SecureURL, nil
}
