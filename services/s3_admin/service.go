package s3_admin

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultEndpointName = "default"

type S3Service struct {
	hostServerProvider host_servers.HostServerProvider
}

type EndpointSummary struct {
	Name                  string     `json:"name"`
	DisplayName           string     `json:"displayName"`
	Endpoint              string     `json:"endpoint"`
	Provider              string     `json:"provider"`
	IsDefault             bool       `json:"isDefault"`
	UseSSL                bool       `json:"useSsl"`
	Manageable            bool       `json:"manageable"`
	DefaultBucket         string     `json:"defaultBucket,omitempty"`
	BucketCount           int        `json:"bucketCount"`
	MatchedHostServerID   *uuid.UUID `json:"matchedHostServerId,omitempty"`
	MatchedHostServerName string     `json:"matchedHostServerName,omitempty"`
	AvailableStorageBytes *uint64    `json:"availableStorageBytes,omitempty"`
}

type BucketSummary struct {
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	ObjectCount  int       `json:"objectCount"`
	TotalSize    int64     `json:"totalSize"`
	IsDefault    bool      `json:"isDefault"`
	EndpointName string    `json:"endpointName"`
}

type ObjectSummary struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"lastModified"`
	ContentType  string    `json:"contentType,omitempty"`
}

func NewService(hostServerProvider host_servers.HostServerProvider) *S3Service {
	return &S3Service{
		hostServerProvider: hostServerProvider,
	}
}

func (s *S3Service) ListEndpoints(ctx context.Context) ([]EndpointSummary, error) {
	defaultEndpoint, err := s.getDefaultEndpointSummary(ctx)
	if err != nil {
		return nil, err
	}

	return []EndpointSummary{defaultEndpoint}, nil
}

func (s *S3Service) ListBuckets(ctx context.Context, endpointName string) ([]BucketSummary, error) {
	client, cfg, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return nil, err
	}

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	summaries := make([]BucketSummary, 0, len(buckets))
	for _, bucket := range buckets {
		objectCount, totalSize, err := s.getBucketStats(ctx, client, bucket.Name)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, BucketSummary{
			Name:         bucket.Name,
			CreatedAt:    bucket.CreationDate,
			ObjectCount:  objectCount,
			TotalSize:    totalSize,
			IsDefault:    bucket.Name == cfg.defaultBucket,
			EndpointName: endpointName,
		})
	}

	return summaries, nil
}

func (s *S3Service) CreateBucket(ctx context.Context, endpointName string, bucketName string) error {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return err
	}

	if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("failed to create bucket %q: %w", bucketName, err)
	}

	return nil
}

func (s *S3Service) DeleteBucket(ctx context.Context, endpointName string, bucketName string) error {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return err
	}

	if err := client.RemoveBucket(ctx, bucketName); err != nil {
		return fmt.Errorf("failed to delete bucket %q: %w", bucketName, err)
	}

	return nil
}

func (s *S3Service) ListObjects(ctx context.Context, endpointName string, bucketName string, prefix string) ([]ObjectSummary, error) {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return nil, err
	}

	objects := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	})

	summaries := make([]ObjectSummary, 0)
	for object := range objects {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects in bucket %q: %w", bucketName, object.Err)
		}

		stat, err := client.StatObject(ctx, bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to stat object %q: %w", object.Key, err)
		}

		summaries = append(summaries, ObjectSummary{
			Key:          object.Key,
			Size:         object.Size,
			ETag:         strings.Trim(object.ETag, `"`),
			LastModified: object.LastModified,
			ContentType:  stat.ContentType,
		})
	}

	return summaries, nil
}

func (s *S3Service) UploadObject(ctx context.Context, endpointName string, bucketName string, objectKey string, reader io.Reader, size int64, contentType string) error {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return err
	}

	if objectKey == "" {
		return fmt.Errorf("object key is required")
	}

	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}

	if _, err := client.PutObject(ctx, bucketName, objectKey, reader, size, opts); err != nil {
		return fmt.Errorf("failed to upload object %q: %w", objectKey, err)
	}

	return nil
}

func (s *S3Service) DeleteObject(ctx context.Context, endpointName string, bucketName string, objectKey string) error {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return err
	}

	if err := client.RemoveObject(ctx, bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object %q: %w", objectKey, err)
	}

	return nil
}

func (s *S3Service) DownloadObject(ctx context.Context, endpointName string, bucketName string, objectKey string) (*minio.Object, minio.ObjectInfo, error) {
	client, _, err := s.clientForEndpoint(endpointName)
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}

	object, err := client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, fmt.Errorf("failed to get object %q: %w", objectKey, err)
	}

	info, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, minio.ObjectInfo{}, fmt.Errorf("failed to stat object %q: %w", objectKey, err)
	}

	return object, info, nil
}

type endpointConfig struct {
	name          string
	displayName   string
	endpoint      string
	useSSL        bool
	accessKey     string
	secretKey     string
	defaultBucket string
}

func (s *S3Service) clientForEndpoint(endpointName string) (*minio.Client, endpointConfig, error) {
	cfg, err := s.resolveEndpoint(endpointName)
	if err != nil {
		return nil, endpointConfig{}, err
	}

	client, err := minio.New(cfg.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.accessKey, cfg.secretKey, ""),
		Secure: cfg.useSSL,
	})
	if err != nil {
		return nil, endpointConfig{}, fmt.Errorf("failed to create s3 client: %w", err)
	}

	return client, cfg, nil
}

func (s *S3Service) resolveEndpoint(endpointName string) (endpointConfig, error) {
	if endpointName != "" && endpointName != defaultEndpointName {
		return endpointConfig{}, fmt.Errorf("unknown s3 endpoint %q", endpointName)
	}

	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("S3_KEYID"))
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET"))
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return endpointConfig{}, fmt.Errorf("s3 endpoint credentials are not fully configured")
	}

	useSSL, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("S3_USESSL")))

	return endpointConfig{
		name:          defaultEndpointName,
		displayName:   "Default S3 Endpoint",
		endpoint:      trimScheme(endpoint),
		useSSL:        useSSL,
		accessKey:     accessKey,
		secretKey:     secretKey,
		defaultBucket: strings.TrimSpace(os.Getenv("S3_DEFAULT_BUCKET")),
	}, nil
}

func (s *S3Service) getDefaultEndpointSummary(ctx context.Context) (EndpointSummary, error) {
	client, cfg, err := s.clientForEndpoint(defaultEndpointName)
	if err != nil {
		return EndpointSummary{}, err
	}

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return EndpointSummary{}, fmt.Errorf("failed to inspect default s3 endpoint: %w", err)
	}

	summary := EndpointSummary{
		Name:          cfg.name,
		DisplayName:   cfg.displayName,
		Endpoint:      cfg.endpoint,
		Provider:      inferProvider(cfg.endpoint),
		IsDefault:     true,
		UseSSL:        cfg.useSSL,
		Manageable:    true,
		DefaultBucket: cfg.defaultBucket,
		BucketCount:   len(buckets),
	}

	if s.hostServerProvider != nil {
		if hostServerID, hostServerName := s.matchHostServer(ctx, cfg.endpoint); hostServerID != nil {
			summary.MatchedHostServerID = hostServerID
			summary.MatchedHostServerName = hostServerName
		}
	}

	return summary, nil
}

func (s *S3Service) getBucketStats(ctx context.Context, client *minio.Client, bucketName string) (int, int64, error) {
	objects := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})
	var count int
	var totalSize int64

	for object := range objects {
		if object.Err != nil {
			return 0, 0, fmt.Errorf("failed to inspect bucket %q: %w", bucketName, object.Err)
		}
		count++
		totalSize += object.Size
	}

	return count, totalSize, nil
}

func (s *S3Service) matchHostServer(ctx context.Context, endpoint string) (*uuid.UUID, string) {
	host := hostFromEndpoint(endpoint)
	if host == "" {
		return nil, ""
	}

	server, err := s.hostServerProvider.GetHostServerByHostname(ctx, host)
	if err == nil && server != nil {
		id := server.ID
		return &id, server.Hostname
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		server, err := s.hostServerProvider.GetHostServerByIP(ctx, ip)
		if err == nil && server != nil {
			id := server.ID
			return &id, server.Hostname
		}
	}

	if resolved := resolveIP(host); resolved.IsValid() {
		server, err := s.hostServerProvider.GetHostServerByIP(ctx, resolved)
		if err == nil && server != nil {
			id := server.ID
			return &id, server.Hostname
		}
	}

	return nil, ""
}

func inferProvider(endpoint string) string {
	lower := strings.ToLower(endpoint)
	switch {
	case strings.Contains(lower, "minio"):
		return "MinIO"
	case strings.Contains(lower, "garage"):
		return "Garage"
	default:
		return "S3-Compatible"
	}
}

func trimScheme(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return strings.TrimSuffix(endpoint, "/")
}

func hostFromEndpoint(endpoint string) string {
	trimmed := trimScheme(endpoint)
	if strings.Contains(trimmed, "/") {
		trimmed = strings.SplitN(trimmed, "/", 2)[0]
	}

	host, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return host
	}

	if strings.Count(trimmed, ":") == 0 {
		return trimmed
	}

	return trimmed
}

func resolveIP(host string) netip.Addr {
	ips, err := net.LookupIP(host)
	if err != nil {
		return netip.Addr{}
	}

	for _, ip := range ips {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			return addr
		}
	}

	return netip.Addr{}
}

func SafeObjectKey(objectKey string) string {
	objectKey = strings.TrimSpace(objectKey)
	objectKey = strings.TrimPrefix(path.Clean("/"+objectKey), "/")
	if objectKey == "." {
		return ""
	}
	return objectKey
}

func BucketNameFromURLValue(value string) string {
	return strings.TrimSpace(value)
}

func ObjectKeyFromURLValue(value string) string {
	decoded, err := url.QueryUnescape(value)
	if err == nil {
		value = decoded
	}
	return SafeObjectKey(value)
}
