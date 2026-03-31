package router

import (
	"testing"
	"time"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
)

func TestLinksToDTO(t *testing.T) {
	now := time.Now().UTC()
	deletedAt := now.Add(2 * time.Hour)

	links := []db.Link{
		{
			ID:          1,
			UserID:      10,
			Code:        "abc123",
			OriginalUrl: "https://example.com/a",
			Clicks:      3,
			CreatedAt:   now,
			DeletedAt:   nil,
		},
		{
			ID:          2,
			UserID:      11,
			Code:        "def456",
			OriginalUrl: "https://example.com/b",
			Clicks:      9,
			CreatedAt:   now.Add(time.Minute),
			DeletedAt:   &deletedAt,
		},
	}

	got := LinksToDTO(links)

	if len(got) != len(links) {
		t.Fatalf("expected %d items, got %d", len(links), len(got))
	}

	if got[0].ID != links[0].ID || got[0].Code != links[0].Code || got[0].OriginalUrl != links[0].OriginalUrl || got[0].Clicks != links[0].Clicks {
		t.Fatalf("unexpected first item mapping: %+v", got[0])
	}
	if got[0].CreatedAt != links[0].CreatedAt {
		t.Fatalf("created_at mismatch: got %v want %v", got[0].CreatedAt, links[0].CreatedAt)
	}
	if got[0].IsDeleted {
		t.Fatalf("expected active link to have is_deleted=false")
	}

	if !got[1].IsDeleted {
		t.Fatalf("expected deleted link to have is_deleted=true")
	}
}

func TestLinksToDTOEmpty(t *testing.T) {
	got := LinksToDTO(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

func TestUserToDTO(t *testing.T) {
	name := "tester"
	pic := "https://example.com/pic.png"

	user := db.User{
		ID:          7,
		Provider:    db.ProviderEnumGITHUB,
		ProviderID:  "12345",
		DisplayName: &name,
		ProfilePic:  &pic,
	}

	got := UserToDTO(user)

	if got.ID != user.ID {
		t.Fatalf("id mismatch: got %d want %d", got.ID, user.ID)
	}
	if got.Provider != user.Provider {
		t.Fatalf("provider mismatch: got %q want %q", got.Provider, user.Provider)
	}
	if got.ProviderID != user.ProviderID {
		t.Fatalf("provider id mismatch: got %q want %q", got.ProviderID, user.ProviderID)
	}
	if got.DisplayName == nil || *got.DisplayName != *user.DisplayName {
		t.Fatalf("display name mismatch: got %v want %v", got.DisplayName, user.DisplayName)
	}
	if got.ProfilePic == nil || *got.ProfilePic != *user.ProfilePic {
		t.Fatalf("profile pic mismatch: got %v want %v", got.ProfilePic, user.ProfilePic)
	}
}
