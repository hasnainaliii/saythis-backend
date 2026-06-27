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

type ImageUploader interface {
	Upload(ctx context.Context, file io.Reader, filename string) (secureURL string, err error)
}

type CloudinaryUploader struct {
	cloudName  string
	apiKey     string
	apiSecret  string
	folder     string
	httpClient *http.Client
}

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

func (c *CloudinaryUploader) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	toSign := fmt.Sprintf("folder=%s&timestamp=%s%s", c.folder, timestamp, c.apiSecret)
	h := sha1.New()
	h.Write([]byte(toSign))
	signature := fmt.Sprintf("%x", h.Sum(nil))

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
