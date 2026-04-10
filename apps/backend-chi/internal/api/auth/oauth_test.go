package auth

import (
	"testing"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
)

func TestNewGoogleOauth2Helper_UsesBackendCallbackURLs(t *testing.T) {
	cfg := &dep.Config{
		BackendPublicURL:   "https://api.example.com",
		GoogleClientID:     "google-id",
		GoogleClientSecret: "google-secret",
		GithubClientID:     "github-id",
		GithubClientSecret: "github-secret",
	}

	helper := NewGoogleOauth2Helper(cfg)

	googleCfg, ok := helper.GetConfigForProvider("google")
	if !ok {
		t.Fatalf("expected google oauth config")
	}
	if googleCfg.Config.RedirectURL != "https://api.example.com/api/v1/user/auth/google/callback" {
		t.Fatalf("expected google callback redirect url, got %q", googleCfg.Config.RedirectURL)
	}

	githubCfg, ok := helper.Configs[models.ProviderEnumGITHUB]
	if !ok {
		t.Fatalf("expected github oauth config")
	}
	if githubCfg.Config.RedirectURL != "https://api.example.com/api/v1/user/auth/github/callback" {
		t.Fatalf("expected github callback redirect url, got %q", githubCfg.Config.RedirectURL)
	}
}
