package domain

import (
	"time"
)

const (
	TripPlaceStatusPlanned = "PLANNED"
	TripPlaceStatusVisited = "VISITED"
	TripPlaceStatusSkipped = "SKIPPED"
)

type TripPlace struct {
	id          TripPlaceID
	dayNumber   int
	arrivalTime *time.Time
	durationMin int
	notes       string
	visitStatus string
	actualCost  float64
	tripID      TripID
	placeID     PlaceID
}

func (tp *TripPlace) GetId() TripPlaceID         { return tp.id }
func (tp *TripPlace) GetDayNumber() int          { return tp.dayNumber }
func (tp *TripPlace) GetArrivalTime() *time.Time { return tp.arrivalTime }
func (tp *TripPlace) GetDurationMin() int        { return tp.durationMin }
func (tp *TripPlace) GetNotes() string           { return tp.notes }
func (tp *TripPlace) GetVisitStatus() string     { return tp.visitStatus }
func (tp *TripPlace) GetActualCost() float64     { return tp.actualCost }
func (tp *TripPlace) GetTripID() TripID          { return tp.tripID }
func (tp *TripPlace) GetPlaceID() PlaceID        { return tp.placeID }

func (tp *TripPlace) SetId(id TripPlaceID)         { tp.id = id }
func (tp *TripPlace) SetDayNumber(day int)         { tp.dayNumber = day }
func (tp *TripPlace) SetArrivalTime(t *time.Time)  { tp.arrivalTime = t }
func (tp *TripPlace) SetDurationMin(min int)       { tp.durationMin = min }
func (tp *TripPlace) SetNotes(notes string)        { tp.notes = notes }
func (tp *TripPlace) SetVisitStatus(status string) { tp.visitStatus = status }
func (tp *TripPlace) SetActualCost(cost float64)   { tp.actualCost = cost }
func (tp *TripPlace) SetTripID(tid TripID)         { tp.tripID = tid }
func (tp *TripPlace) SetPlaceID(pid PlaceID)       { tp.placeID = pid }

func NewTripPlace(
	id TripPlaceID,
	dayNumber int,
	arrivalTime *time.Time,
	durationMin int,
	notes string,
	visitStatus string,
	actualCost float64,
	tripID TripID,
	placeID PlaceID,
) (*TripPlace, error) {
	if dayNumber < 1 {
		return nil, ErrInvalidDayNumber
	}
	if durationMin < 0 {
		return nil, ErrNegativeDuration
	}
	if len(notes) > GetConfig().MaxTripPlaceNotesLength {
		return nil, ErrTripPlaceNotesTooLong
	}
	if visitStatus == "" {
		return nil, ErrEmptyVisitStatus
	}
	validStatuses := []string{TripPlaceStatusPlanned, TripPlaceStatusVisited, TripPlaceStatusSkipped}
	valid := false
	for _, s := range validStatuses {
		if visitStatus == s {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidVisitStatus
	}
	if actualCost < 0 {
		return nil, ErrNegativeActualCost
	}
	if tripID == (TripID{}) {
		return nil, ErrInvalidTripID
	}
	if placeID == (PlaceID{}) {
		return nil, ErrInvalidPlaceID
	}
	return &TripPlace{
		id:          id,
		dayNumber:   dayNumber,
		arrivalTime: arrivalTime,
		durationMin: durationMin,
		notes:       notes,
		visitStatus: visitStatus,
		actualCost:  actualCost,
		tripID:      tripID,
		placeID:     placeID,
	}, nil
}

func NewTripPlaceFromDB(
	id TripPlaceID,
	dayNumber int,
	arrivalTime *time.Time,
	durationMin int,
	notes string,
	visitStatus string,
	actualCost float64,
	tripID TripID,
	placeID PlaceID,
) *TripPlace {
	return &TripPlace{
		id:          id,
		dayNumber:   dayNumber,
		arrivalTime: arrivalTime,
		durationMin: durationMin,
		notes:       notes,
		visitStatus: visitStatus,
		actualCost:  actualCost,
		tripID:      tripID,
		placeID:     placeID,
	}
}
