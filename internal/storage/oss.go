package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// OSSConfig contains the minimum configuration needed for private OSS
// uploads. Credentials can be supplied through config.yaml or environment
// variables; environment variables take precedence.
type OSSConfig struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	KeyPrefix       string
}

// Client implements the OSS Object API using the standard signed REST
// request. Keeping the adapter small avoids exposing AccessKey credentials to
// the mini-program and does not require a third-party SDK at runtime.
type Client struct {
	config OSSConfig
	client *http.Client
}

var ossHTTPClient = &http.Client{Timeout: 30 * time.Second}

var (
	ossClientOnce sync.Once
	ossClient     *Client
	ossClientErr  error
)

// InitOSSClient initializes the process-wide OSS client. Configuration is
// loaded once during application startup and the same HTTP client/signing
// configuration is reused by every request.
func InitOSSClient(ctx context.Context) error {
	ossClientOnce.Do(func() {
		ossClient, ossClientErr = newOSSClient(ctx)
	})
	return ossClientErr
}

// OSSClient returns the initialized process-wide OSS client.
func OSSClient() (*Client, error) {
	if ossClientErr != nil {
		return nil, ossClientErr
	}
	if ossClient == nil {
		return nil, fmt.Errorf("oss client is not initialized")
	}
	return ossClient, nil
}

// NewOSSClient is kept as a compatibility wrapper for existing callers. It
// now returns the singleton instead of constructing a client per request.
func NewOSSClient(ctx context.Context) (*Client, error) {
	if err := InitOSSClient(ctx); err != nil {
		return nil, err
	}
	return OSSClient()
}

func newOSSClient(ctx context.Context) (*Client, error) {
	config := OSSConfig{
		Endpoint:        configValue(ctx, "oss.endpoint", "OSS_ENDPOINT"),
		Bucket:          configValue(ctx, "oss.bucket", "OSS_BUCKET"),
		AccessKeyID:     configValue(ctx, "oss.access_key_id", "OSS_ACCESS_KEY_ID"),
		AccessKeySecret: configValue(ctx, "oss.access_key_secret", "OSS_ACCESS_KEY_SECRET"),
		KeyPrefix:       configValue(ctx, "oss.key_prefix", "OSS_KEY_PREFIX"),
	}
	config.Endpoint = strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.KeyPrefix = strings.Trim(config.KeyPrefix, "/")
	if config.Endpoint == "" {
		config.Endpoint = "https://oss-cn-beijing.aliyuncs.com"
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "indu/verification"
	}
	if config.Endpoint == "" || config.Bucket == "" || config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return nil, fmt.Errorf("oss.endpoint, oss.bucket, oss.access_key_id and oss.access_key_secret are required")
	}
	if !strings.HasPrefix(config.Endpoint, "http://") && !strings.HasPrefix(config.Endpoint, "https://") {
		config.Endpoint = "https://" + config.Endpoint
	}
	if strings.HasPrefix(strings.ToLower(config.Endpoint), "http://") {
		return nil, fmt.Errorf("oss.endpoint must use HTTPS")
	}
	return &Client{config: config, client: ossHTTPClient}, nil
}

func (c *Client) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	objectKey = strings.TrimLeft(objectKey, "/")
	if objectKey == "" {
		return "", fmt.Errorf("oss object key is empty")
	}
	if size < 0 {
		return "", fmt.Errorf("oss object size is invalid")
	}
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectURL, canonicalResource, err := c.objectURL(objectKey)
	if err != nil {
		return "", err
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	// Business licenses are sensitive identity documents. Force every object
	// uploaded by this adapter to remain private, regardless of the bucket's
	// default ACL.
	canonicalizedHeaders := "x-oss-object-acl:private\n"
	stringToSign := "PUT\n\n" + contentType + "\n" + date + "\n" + canonicalizedHeaders + canonicalResource
	mac := hmac.New(sha1.New, []byte(c.config.AccessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, reader)
	if err != nil {
		return "", fmt.Errorf("create oss request: %w", err)
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Date", date)
	req.Header.Set("x-oss-object-acl", "private")
	req.Header.Set("Authorization", "OSS "+c.config.AccessKeyID+":"+signature)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload object to oss: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("oss upload failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return strings.Trim(resp.Header.Get("ETag"), `"`), nil
}

// Delete removes a private object. It is used to roll back an OSS upload when
// the corresponding database metadata cannot be saved.
func (c *Client) Delete(ctx context.Context, objectKey string) error {
	return c.deleteObject(ctx, c.config.Endpoint, c.config.Bucket, objectKey)
}

// DeleteWithTimeout performs cleanup independently of a cancelled upload
// request. It prevents a client disconnect from leaving a successfully
// uploaded object behind forever.
func DeleteWithTimeout(parent context.Context, client *Client, objectKey string) error {
	if client == nil {
		return fmt.Errorf("oss client is nil")
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.Delete(cleanupCtx, objectKey)
}

// DeleteFor removes an object from a historical endpoint and bucket recorded
// with the file metadata. This keeps cleanup safe after an OSS migration.
func (c *Client) DeleteFor(ctx context.Context, endpointHost, bucket, objectKey string) error {
	return c.deleteObject(ctx, endpointHost, bucket, objectKey)
}

func (c *Client) deleteObject(ctx context.Context, endpointHost, bucket, objectKey string) error {
	endpointHost = strings.TrimRight(strings.TrimSpace(endpointHost), "/")
	bucket = strings.TrimSpace(bucket)
	if endpointHost == "" || bucket == "" {
		return fmt.Errorf("oss endpoint and bucket are required")
	}
	if !strings.HasPrefix(strings.ToLower(endpointHost), "http://") && !strings.HasPrefix(strings.ToLower(endpointHost), "https://") {
		endpointHost = "https://" + endpointHost
	}
	if strings.HasPrefix(strings.ToLower(endpointHost), "http://") {
		return fmt.Errorf("oss endpoint must use HTTPS")
	}
	endpoint, err := url.Parse(endpointHost)
	if err != nil || endpoint.Host == "" {
		return fmt.Errorf("oss endpoint is invalid")
	}
	key := strings.TrimLeft(objectKey, "/")
	if key == "" {
		return fmt.Errorf("oss object key is empty")
	}
	host := endpoint.Host
	if !strings.HasPrefix(strings.ToLower(host), strings.ToLower(bucket)+".") {
		host = bucket + "." + host
	}
	objectURL := (&url.URL{Scheme: endpoint.Scheme, Host: host, Path: "/" + key}).String()
	canonicalResource := "/" + bucket + "/" + key
	date := time.Now().UTC().Format(http.TimeFormat)
	stringToSign := "DELETE\n\n\n" + date + "\n" + canonicalResource
	mac := hmac.New(sha1.New, []byte(c.config.AccessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, objectURL, nil)
	if err != nil {
		return fmt.Errorf("create oss delete request: %w", err)
	}
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", "OSS "+c.config.AccessKeyID+":"+signature)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete object from oss: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("oss delete failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) objectURL(objectKey string) (string, string, error) {
	endpoint, err := url.Parse(c.config.Endpoint)
	if err != nil || endpoint.Host == "" {
		return "", "", fmt.Errorf("oss.endpoint is invalid")
	}
	host := endpoint.Host
	if !strings.HasPrefix(strings.ToLower(host), strings.ToLower(c.config.Bucket)+".") {
		host = c.config.Bucket + "." + host
	}
	pathValue := "/" + objectKey
	objectURL := (&url.URL{Scheme: endpoint.Scheme, Host: host, Path: pathValue}).String()
	return objectURL, "/" + c.config.Bucket + pathValue, nil
}

func (c *Client) ObjectKey(prefix, name string) string {
	parts := []string{c.config.KeyPrefix, strings.Trim(prefix, "/"), strings.TrimLeft(name, "/")}
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return strings.Join(result, "/")
}

func (c *Client) Bucket() string {
	return c.config.Bucket
}

// EndpointHost returns the host portion used to build OSS object addresses,
// for example oss-cn-beijing.aliyuncs.com.
func (c *Client) EndpointHost() string {
	endpoint, err := url.Parse(c.config.Endpoint)
	if err != nil {
		return strings.TrimPrefix(strings.TrimPrefix(c.config.Endpoint, "https://"), "http://")
	}
	return endpoint.Host
}

// SignedURLFor creates a short-lived private GET URL for an OSS object. The
// endpoint and bucket are accepted explicitly so historical objects can still
// be opened when the active deployment points at another bucket.
func (c *Client) SignedURLFor(endpointHost, bucket, objectKey string, lifetime time.Duration) (string, error) {
	endpointHost = strings.TrimSpace(endpointHost)
	bucket = strings.TrimSpace(bucket)
	objectKey = strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if endpointHost == "" || bucket == "" || objectKey == "" {
		return "", fmt.Errorf("oss endpoint, bucket and object key are required")
	}
	if lifetime <= 0 {
		lifetime = 10 * time.Minute
	}
	if !strings.HasPrefix(endpointHost, "http://") && !strings.HasPrefix(endpointHost, "https://") {
		endpointHost = "https://" + endpointHost
	}
	if strings.HasPrefix(strings.ToLower(endpointHost), "http://") {
		return "", fmt.Errorf("oss endpoint must use HTTPS")
	}
	endpoint, err := url.Parse(strings.TrimRight(endpointHost, "/"))
	if err != nil || endpoint.Host == "" {
		return "", fmt.Errorf("oss endpoint is invalid")
	}
	expires := time.Now().Add(lifetime).Unix()
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := "GET\n\n\n" + strconv.FormatInt(expires, 10) + "\n" + canonicalResource
	mac := hmac.New(sha1.New, []byte(c.config.AccessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	query := url.Values{}
	query.Set("Expires", strconv.FormatInt(expires, 10))
	query.Set("OSSAccessKeyId", c.config.AccessKeyID)
	query.Set("Signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	host := bucket + "." + endpoint.Host
	return (&url.URL{Scheme: endpoint.Scheme, Host: host, Path: "/" + objectKey, RawQuery: query.Encode()}).String(), nil
}

func configValue(ctx context.Context, key, envKey string) string {
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return value
	}
	value, err := g.Cfg().Get(ctx, key)
	if err != nil || value == nil {
		return ""
	}
	return strings.TrimSpace(value.String())
}
