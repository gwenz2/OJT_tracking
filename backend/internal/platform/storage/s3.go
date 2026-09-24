// Package storage is the private object-store adapter used only for evidence
// media. It deliberately exposes just the operations the evidence workflow
// needs — Put, PresignGet, Delete, Health — not a generic filesystem.
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/skycode/ojt-management/backend/internal/config"
)

type Store struct {
	client  *minio.Client
	presign *minio.Client
	bucket  string
	ttl     time.Duration
}

func NewS3(cfg config.S3Config) (*Store, error) {
	creds := credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, "")
	client, err := minio.New(cfg.Endpoint, &minio.Options{Creds: creds, Secure: cfg.UseSSL})
	if err != nil {
		return nil, fmt.Errorf("s3 client: %w", err)
	}
	presign := client
	if cfg.PublicEndpoint != "" {
		// Presigning is offline signature work — this client never dials the
		// public endpoint; it exists only so generated URLs carry the host the
		// browser will actually use (and that the edge proxies back to MinIO).
		// Region is pinned so minio-go skips the bucket-location probe — the
		// public host is unreachable from this process by design.
		presign, err = minio.New(cfg.PublicEndpoint, &minio.Options{
			Creds: creds, Secure: cfg.PublicSSL, Region: "us-east-1",
		})
		if err != nil {
			return nil, fmt.Errorf("s3 presign client: %w", err)
		}
	}
	return &Store{client: client, presign: presign, bucket: cfg.Bucket, ttl: cfg.PresignTTL}, nil
}

// EnsureBucket creates the configured private bucket if it does not exist.
// Called once at startup in development; harmless if it already exists.
func (s *Store) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	return nil
}

// Put stores object bytes under a generated key.
func (s *Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Get returns a readable stream for an object. Callers must close it.
func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return obj, nil
}

// PresignGet returns a short-lived authorized read URL for a private object.
func (s *Store) PresignGet(ctx context.Context, key string) (string, error) {
	u, err := s.presign.PresignedGetObject(ctx, s.bucket, key, s.ttl, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign %q: %w", key, err)
	}
	return u.String(), nil
}

// Delete removes an object (used for orphan cleanup after DB failures).
func (s *Store) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

// Health performs a bounded dependency check without leaking details.
func (s *Store) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
