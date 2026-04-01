package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	stateStore "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/statestore"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"golang.org/x/oauth2"
)

const (
	GoogleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	GoogleTokenURL    = "https://oauth2.googleapis.com/token"
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	GithubAuthURL     = "https://github.com/login/oauth/authorize"
	GithubTokenURL    = "https://github.com/login/oauth/access_token"
	GithubUserInfoURL = "https://api.github.com/user"
)

type OauthHandler interface {
	GetConfigForProvider(provider string) (*OauthConfig, bool)
	GetOauthAuthURL(opt *oauth2.Config, store stateStore.StateStore) string
	ExchangeCodeAndGetClient(ctx context.Context, opt *oauth2.Config, code string, verifier string) (*http.Client, error)
}

type OauthConfig struct {
	Config      oauth2.Config
	GetUserInfo func(client *http.Client) (*db.UpsertUserParams, error)
}

type GoogleOauth2Helper struct {
	Configs map[db.ProviderEnum]OauthConfig
}

func NewGoogleOauth2Helper(cfg *dep.Config) *GoogleOauth2Helper {
	configs := make(map[db.ProviderEnum]OauthConfig)

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		configs[db.ProviderEnumGOOGLE] = OauthConfig{
			Config: oauth2.Config{
				ClientID:     cfg.GoogleClientID,
				ClientSecret: cfg.GoogleClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  GoogleAuthURL,
					TokenURL: GoogleTokenURL,
				},
				RedirectURL: cfg.FrontendRedirectURL,
				Scopes:      []string{"profile", "openid"},
			},
			GetUserInfo: getAndParseGoogleUserInfo,
		}
	}

	if cfg.GithubClientID != "" && cfg.GithubClientSecret != "" {
		configs[db.ProviderEnumGITHUB] = OauthConfig{
			Config: oauth2.Config{
				ClientID:     cfg.GithubClientID,
				ClientSecret: cfg.GithubClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  GithubAuthURL,
					TokenURL: GithubTokenURL,
				},
				RedirectURL: cfg.FrontendRedirectURL,
				Scopes:      []string{"read:user"},
			},
			GetUserInfo: getAndParseGithubUserInfo,
		}
	}

	return &GoogleOauth2Helper{Configs: configs}
}

func (h *GoogleOauth2Helper) GetConfigForProvider(provider string) (*OauthConfig, bool) {
	provider = strings.TrimSpace(strings.ToUpper(provider))
	config, ok := h.Configs[db.ProviderEnum(provider)]
	return &config, ok
}

func (h *GoogleOauth2Helper) GetOauthAuthURL(opt *oauth2.Config, store stateStore.StateStore) string {
	varifier := oauth2.GenerateVerifier()
	state := oauth2.GenerateVerifier()
	store.Add(state, varifier)

	return opt.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(varifier))
}

func (h *GoogleOauth2Helper) ExchangeCodeAndGetClient(ctx context.Context, opt *oauth2.Config, code string, verifier string) (*http.Client, error) {
	token, err := opt.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}

	client := opt.Client(ctx, token)
	return client, nil
}

func getAndParseUserInfo(client *http.Client, userInfoURL string) (map[string]interface{}, error) {
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getRequiredStringField(data map[string]interface{}, key string) (string, error) {
	raw, ok := data[key]
	if !ok {
		return "", fmt.Errorf("missing %s in user info response", key)
	}

	switch v := raw.(type) {
	case string:
		if v == "" {
			return "", fmt.Errorf("empty %s in user info response", key)
		}
		return v, nil
	case float64:
		return fmt.Sprintf("%.0f", v), nil
	default:
		return "", fmt.Errorf("invalid %s type %T in user info response", key, raw)
	}
}

func getAndParseGoogleUserInfo(client *http.Client) (*db.UpsertUserParams, error) {
	data, err := getAndParseUserInfo(client, GoogleUserInfoURL)

	if err != nil {
		return nil, err
	}

	providerID, err := getRequiredStringField(data, "sub")
	if err != nil {
		return nil, err
	}

	var displayName *string
	if name, ok := data["name"].(string); ok && name != "" {
		displayName = &name
	} else {
		displayName = nil
	}

	var profilePic string
	if pic, ok := data["picture"].(string); ok && pic != "" {
		profilePic = pic
	} else {
		profilePic = ""
	}

	return &db.UpsertUserParams{
		Provider:    db.ProviderEnumGOOGLE,
		ProviderID:  providerID,
		DisplayName: displayName,
		ProfilePic:  &profilePic,
	}, nil
}

func getAndParseGithubUserInfo(client *http.Client) (*db.UpsertUserParams, error) {
	data, err := getAndParseUserInfo(client, GithubUserInfoURL)

	if err != nil {
		return nil, err
	}

	providerID, err := getRequiredStringField(data, "id")
	if err != nil {
		return nil, err
	}

	var displayName *string
	if name, ok := data["name"].(string); ok && name != "" {
		displayName = &name
	} else {
		displayName = nil
	}

	var profilePic *string
	if pic, ok := data["avatar_url"].(string); ok && pic != "" {
		profilePic = &pic
	} else {
		profilePic = nil
	}

	return &db.UpsertUserParams{
		Provider:    db.ProviderEnumGITHUB,
		ProviderID:  providerID,
		DisplayName: displayName,
		ProfilePic:  profilePic,
	}, nil
}
