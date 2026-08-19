package repository

import (
	"context"
	"net/http"
	"path"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/r2"
	"github.com/insanelyharsh/web-portfolio/internal/config"
)

type MediaRepository interface {
	// GetObject fetches an object from R2 by key (resolved against the
	// configured base path). The caller owns the returned response body
	// and must close it.
	GetObject(ctx context.Context, key string) (*http.Response, error)
}

type MediaRepositoryImpl struct {
	client *cloudflare.Client
	cfg    config.R2Config
}

func NewMediaRepository(client *cloudflare.Client, cfg config.R2Config) MediaRepository {
	return &MediaRepositoryImpl{
		client: client,
		cfg:    cfg,
	}
}

func (m *MediaRepositoryImpl) GetObject(ctx context.Context, key string) (*http.Response, error) {
	fullKey := path.Join(m.cfg.BasePath, key)

	return m.client.R2.Buckets.Objects.Get(ctx, m.cfg.Bucket, fullKey, r2.BucketObjectGetParams{
		AccountID: cloudflare.F(m.cfg.AccountID),
	})
}
