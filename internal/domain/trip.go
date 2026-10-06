package domain

import (
	"strings"
	"time"
)

const (
	TripStatusDraft    = "DRAFT"
	TripStatusPlanned  = "PLANNED"
	TripStatusOngoing  = "ONGOING"
	TripStatusFinished = "FINISHED"
	TripStatusCanceled = "CANCELED"
)

type Trip struct {
	id        TripID
	title     string
	startDate time.Time
	endDate   time.Time
	budget    float64
	status    string
	notes     string
	userID    UserID
	imageID   ImageID
	createdAt time.Time
}

func (t *Trip) GetId() TripID {
	return t.id
}

func (t *Trip) GetTitle() string {
	return t.title
}

func (t *Trip) GetStartDate() time.Time {
	return t.startDate
}

func (t *Trip) GetEndDate() time.Time {
	return t.endDate
}

func (t *Trip) GetBudget() float64 {
	return t.budget
}

func (t *Trip) GetStatus() string {
	return t.status
}

func (t *Trip) GetNotes() string {
	return t.notes
}

func (t *Trip) GetUserID() UserID {
	return t.userID
}

func (t *Trip) GetImageID() ImageID {
	return t.imageID
}

func (t *Trip) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t *Trip) SetId(id TripID) {
	t.id = id
}

func (t *Trip) SetTitle(title string) {
	t.title = title
}

func (t *Trip) SetStartDate(startDate time.Time) {
	t.startDate = startDate
}

func (t *Trip) SetEndDate(endDate time.Time) {
	t.endDate = endDate
}

func (t *Trip) SetBudget(budget float64) {
	t.budget = budget
}

func (t *Trip) SetStatus(status string) {
	t.status = status
}

func (t *Trip) SetNotes(notes string) {
	t.notes = notes
}

func (t *Trip) SetUserID(userID UserID) {
	t.userID = userID
}

func (t *Trip) SetImageID(imageID ImageID) {
	t.imageID = imageID
}

func NewTrip(
	id TripID,
	title string,
	startDate, endDate time.Time,
	budget float64,
	status string,
	notes string,
	userID UserID,
	imageID ImageID,
) (*Trip, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTripTitle
	}
	if len(title) > GetConfig().MaxTripTitleLength {
		return nil, ErrTripTitleTooLong
	}

	if startDate.IsZero() {
		return nil, ErrInvalidStartDate
	}
	if endDate.IsZero() {
		return nil, ErrInvalidEndDate
	}
	if endDate.Before(startDate) {
		return nil, ErrEndDateBeforeStartDate
	}

	if budget < 0 {
		return nil, ErrNegativeBudget
	}

	if status == "" {
		return nil, ErrEmptyTripStatus
	}
	validStatuses := []string{TripStatusDraft, TripStatusPlanned, TripStatusOngoing, TripStatusFinished, TripStatusCanceled}
	valid := false
	for _, s := range validStatuses {
		if status == s {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidTripStatus
	}

	if len(notes) > GetConfig().MaxTripNotesLength {
		return nil, ErrTripNotesTooLong
	}

	if userID == (UserID{}) {
		return nil, ErrInvalidUserID
	}

	return &Trip{
		id:        id,
		title:     title,
		startDate: startDate,
		endDate:   endDate,
		budget:    budget,
		status:    status,
		notes:     notes,
		userID:    userID,
		imageID:   imageID,
		createdAt: time.Now(),
	}, nil
}

func NewTripFromDB(
	id TripID,
	title string,
	startDate, endDate time.Time,
	createdAt time.Time,
	budget float64,
	status string,
	notes string,
	userID UserID,
	imageID ImageID,
) *Trip {
	return &Trip{
		id:        id,
		title:     title,
		startDate: startDate,
		endDate:   endDate,
		budget:    budget,
		status:    status,
		notes:     notes,
		userID:    userID,
		imageID:   imageID,
		createdAt: createdAt,
	}
}
