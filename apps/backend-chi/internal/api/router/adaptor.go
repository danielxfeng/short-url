package router

import (
	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
)

func LinksToDTO(links []db.Link) []dto.LinkResponse {
	result := make([]dto.LinkResponse, len(links))
	for i, link := range links {
		result[i] = dto.LinkResponse{
			ID:          link.ID,
			Code:        link.Code,
			OriginalUrl: link.OriginalUrl,
			Clicks:      link.Clicks,
			CreatedAt:   link.CreatedAt,
			IsDeleted:   link.DeletedAt != nil,
		}
	}
	return result
}

func UserToDTO(user db.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Provider:    user.Provider,
		ProviderID:  user.ProviderID,
		DisplayName: user.DisplayName,
		ProfilePic:  user.ProfilePic,
	}
}
