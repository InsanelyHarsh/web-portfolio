package config

import (
	"errors"
	"os"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type R2Config struct {
	AccountID string
	Bucket    string
	BasePath  string
}

func InitCloudflareR2() (*cloudflare.Client, R2Config, error) {
	token := os.Getenv("CLOUDFLARE_R2_TOKEN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucket := os.Getenv("CLOUDFLARE_R2_BUCKET")

	if token == "" || accountID == "" || bucket == "" {
		return nil, R2Config{}, errors.New("CLOUDFLARE_R2_TOKEN, CLOUDFLARE_ACCOUNT_ID, and CLOUDFLARE_R2_BUCKET must be set")
	}

	client := cloudflare.NewClient(option.WithAPIToken(token))
	cfg := R2Config{
		AccountID: accountID,
		Bucket:    bucket,
		BasePath:  os.Getenv("CLOUDFLARE_R2_BASE_PATH"),
	}
	return client, cfg, nil
}
