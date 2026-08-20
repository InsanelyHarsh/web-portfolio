package config

import (
	"errors"
	"os"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

// R2Config holds the bucket-level settings needed to address objects in
// Cloudflare R2. These aren't part of the SDK client itself since
// cloudflare-go's R2 object methods take the account/bucket per call.
type R2Config struct {
	AccountID string
	Bucket    string
	// BasePath is an optional key prefix within the bucket that requested
	// media keys are resolved against (e.g. "blog-media").
	BasePath string
}

// InitCloudflareR2 builds a Cloudflare API client and the R2 bucket config
// needed to fetch objects from it. CLOUDFLARE_R2_TOKEN, CLOUDFLARE_ACCOUNT_ID,
// and CLOUDFLARE_R2_BUCKET are required; CLOUDFLARE_R2_BASE_PATH is optional.
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
