package domain

import "time"

type ReportFormat string

const (
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
)

type ReportOptions struct {
	IncludeHeaders bool
	PrettyPrint    bool
}

type ReportConfig struct {
	Title    string
	DateFrom time.Time
	DateTo   time.Time
	Format   ReportFormat
	Options  ReportOptions
}

type PlaceReportData struct {
	PlaceID      PlaceID `json:"placeId"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	TripsCount   int     `json:"tripsCount"`
	ReviewsCount int     `json:"reviewsCount"`
	AvgRating    float64 `json:"avgRating"`
}

type UserActivityData struct {
	UserID         UserID    `json:"userId"`
	Username       string    `json:"username"`
	TripsCreated   int       `json:"tripsCreated"`
	ReviewsWritten int       `json:"reviewsWritten"`
	LastActive     time.Time `json:"lastActive"`
}

type TripStatData struct {
	TripID      TripID    `json:"tripId"`
	Title       string    `json:"title"`
	UserID      UserID    `json:"userId"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	PlacesCount int       `json:"placesCount"`
	TotalCost   float64   `json:"totalCost"`
}
