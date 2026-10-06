package domain

import (
	"strings"
	"time"
)

type Review struct {
	id                ReviewID
	rating            int
	comment           string
	visitDate         time.Time
	createdAt         time.Time
	isModerated       bool
	isApproved        bool
	moderationComment string
	userID            UserID
	placeID           PlaceID
	imageID           ImageID

	username   string
	userAvatar string
	placeName  string
}

func (r *Review) GetId() ReviewID {
	return r.id
}

func (r *Review) GetRating() int {
	return r.rating
}

func (r *Review) GetComment() string {
	return r.comment
}

func (r *Review) GetVisitDate() time.Time {
	return r.visitDate
}

func (r *Review) GetCreatedAt() time.Time {
	return r.createdAt
}

func (r *Review) IsModerated() bool {
	return r.isModerated
}

func (r *Review) IsApproved() bool {
	return r.isApproved
}

func (r *Review) GetUserID() UserID {
	return r.userID
}

func (r *Review) GetPlaceID() PlaceID {
	return r.placeID
}

func (r *Review) GetModerationComment() string {
	return r.moderationComment
}

func (r *Review) GetImageID() ImageID {
	return r.imageID
}

func (r *Review) GetUsername() string {
	return r.username
}

func (r *Review) GetUserAvatar() string {
	return r.userAvatar
}

func (r *Review) GetPlaceName() string {
	return r.placeName
}

func (r *Review) SetId(id ReviewID) {
	r.id = id
}

func (r *Review) SetRating(rating int) {
	r.rating = rating
}

func (r *Review) SetComment(comment string) {
	r.comment = comment
}

func (r *Review) SetVisitDate(visitDate time.Time) {
	r.visitDate = visitDate
}

func (r *Review) SetCreatedAt(createdAt time.Time) {
	r.createdAt = createdAt
}

func (r *Review) SetModerated(moderated bool) {
	r.isModerated = moderated
}

func (r *Review) SetModerationComment(comment string) {
	r.moderationComment = comment
}

func (r *Review) SetApproved(approved bool) {
	r.isApproved = approved
}

func (r *Review) SetUserID(userID UserID) {
	r.userID = userID
}

func (r *Review) SetPlaceID(placeID PlaceID) {
	r.placeID = placeID
}

func (r *Review) SetImageID(imageID ImageID) {
	r.imageID = imageID
}

func NewReview(
	id ReviewID,
	rating int,
	comment string,
	visitDate time.Time,
	userID UserID,
	placeID PlaceID,
	imageID ImageID,
) (*Review, error) {
	if rating < GetConfig().MinRating || rating > GetConfig().MaxRating {
		return nil, ErrInvalidRating
	}

	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, ErrEmptyComment
	}
	if len(comment) > GetConfig().MaxReviewCommentLength {
		return nil, ErrCommentTooLong
	}

	if visitDate.After(time.Now()) {
		return nil, ErrVisitDateInFuture
	}

	if userID == (UserID{}) {
		return nil, ErrInvalidUserID
	}
	if placeID == (PlaceID{}) {
		return nil, ErrInvalidPlaceID
	}

	return &Review{
		id:                (ReviewID{}),
		rating:            rating,
		comment:           comment,
		visitDate:         visitDate,
		createdAt:         time.Now(),
		isModerated:       false,
		moderationComment: "",
		isApproved:        false,
		userID:            userID,
		placeID:           placeID,
		imageID:           imageID,
	}, nil
}
func NewReviewFromDB(
	id ReviewID,
	rating int,
	comment string,
	visitDate time.Time,
	createdAt time.Time,
	isModerated, isApproved bool,
	moderationComment string,
	userID UserID,
	placeID PlaceID,
	imageID ImageID,
	username string,
) *Review {
	return &Review{
		id:                id,
		rating:            rating,
		comment:           comment,
		visitDate:         visitDate,
		createdAt:         createdAt,
		isModerated:       isModerated,
		isApproved:        isApproved,
		moderationComment: moderationComment,
		userID:            userID,
		placeID:           placeID,
		imageID:           imageID,
		username:          username,
	}
}

func CalculateNewAverageRating(place *Place, newRating int) (newAvgRating float64, newCount int) {
	oldTotal := place.GetAvgRating() * float64(place.GetReviewCnt())
	newTotal := oldTotal + float64(newRating)
	newCount = int(place.GetReviewCnt() + 1)
	if newCount == 0 {
		return 0, 0
	}
	newAvgRating = newTotal / float64(newCount)
	return
}

func RecalculateRatingAfterRemoval(place *Place, removedRating int) (newAvgRating float64, newCount int) {
	if place.GetReviewCnt() <= 1 {
		return 0, 0
	}
	oldTotal := place.GetAvgRating() * float64(place.GetReviewCnt())
	newTotal := oldTotal - float64(removedRating)
	newCount = int(place.GetReviewCnt() - 1)
	newAvgRating = newTotal / float64(newCount)
	return
}
