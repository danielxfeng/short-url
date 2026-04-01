package router

import (
	"testing"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
)

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
