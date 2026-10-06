package interfaces

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
)

type PlaceRepository interface {
	Save(ctx context.Context, place *domain.Place) error
	FindByID(ctx context.Context, id domain.PlaceID) (*domain.Place, error)
	FindByCity(ctx context.Context, cityID domain.CityID) ([]*domain.Place, error)
	FindByCategory(ctx context.Context, category string) ([]*domain.Place, error)
	Update(ctx context.Context, place *domain.Place) error
	Delete(ctx context.Context, id domain.PlaceID) error
	UpdateRating(ctx context.Context, id domain.PlaceID, newAvgRating float64, newReviewsCount int) error

	SaveReview(ctx context.Context, review *domain.Review) error
	FindReviewByID(ctx context.Context, id domain.ReviewID) (*domain.Review, error)
	FindReviewsByPlace(ctx context.Context, placeID domain.PlaceID) ([]*domain.Review, error)
	FindReviewsByUserAndPlace(ctx context.Context, userID domain.UserID, placeID domain.PlaceID) ([]*domain.Review, error)
	FindReviewsByUser(ctx context.Context, userID domain.UserID) ([]*domain.Review, error)
	UpdateReview(ctx context.Context, review *domain.Review) error
	DeleteReview(ctx context.Context, id domain.ReviewID) error
	ModerateReview(ctx context.Context, id domain.ReviewID, approved bool, moderationComment string) error

	GetPopularPlacesReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.PlaceReportData, error)
	FindPendingReviews(ctx context.Context, limit, offset int) ([]*domain.Review, error)
}
