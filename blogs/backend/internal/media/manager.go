package media

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/insanelyharsh/web-portfolio/internal/media/repository"
)

// ErrNotFound is returned when a requested media object doesn't exist, so
// callers (e.g. the HTTP layer) can distinguish "not found" from other
// failures via errors.Is instead of matching on an error string.
var ErrNotFound = errors.New("not found")

type MediaManager struct {
	repo repository.MediaRepository
}

func NewMediaManager(repo repository.MediaRepository) *MediaManager {
	return &MediaManager{
		repo: repo,
	}
}

func (m *MediaManager) GetObject(ctx context.Context, key string) (*http.Response, error) {
	res, err := m.repo.GetObject(ctx, key)
	if err != nil {
		var apiErr *cloudflare.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return res, nil
}
