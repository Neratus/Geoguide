package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
)

type ReportGeneratorService struct {
	logger *slog.Logger
}

func NewReportGeneratorService(logger *slog.Logger) *ReportGeneratorService {
	return &ReportGeneratorService{logger: logger}
}

func (r *ReportGeneratorService) GeneratePopularPlacesReport(places []domain.PlaceReportData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error) {
	switch format {
	case domain.FormatJSON:
		return r.generatePopularPlacesJSON(places, opts)
	case domain.FormatCSV:
		return r.generatePopularPlacesCSV(places, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *ReportGeneratorService) GenerateUserActivityReport(users []domain.UserActivityData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error) {
	switch format {
	case domain.FormatJSON:
		return r.generateUserActivityJSON(users, opts)
	case domain.FormatCSV:
		return r.generateUserActivityCSV(users, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *ReportGeneratorService) GenerateTripStatisticsReport(trips []domain.TripStatData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error) {
	switch format {
	case domain.FormatJSON:
		return r.generateTripStatsJSON(trips, opts)
	case domain.FormatCSV:
		return r.generateTripStatsCSV(trips, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *ReportGeneratorService) GenerateReport(data interface{}, config domain.ReportConfig) ([]byte, error) {
	switch d := data.(type) {
	case []domain.PlaceReportData:
		return r.GeneratePopularPlacesReport(d, config.Format, config.Options)
	case []domain.UserActivityData:
		return r.GenerateUserActivityReport(d, config.Format, config.Options)
	case []domain.TripStatData:
		return r.GenerateTripStatisticsReport(d, config.Format, config.Options)
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}
}

func (r *ReportGeneratorService) generatePopularPlacesJSON(places []domain.PlaceReportData, opts domain.ReportOptions) ([]byte, error) {
	if opts.PrettyPrint {
		return json.MarshalIndent(places, "", "  ")
	}
	return json.Marshal(places)
}

func (r *ReportGeneratorService) generatePopularPlacesCSV(places []domain.PlaceReportData, opts domain.ReportOptions) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if opts.IncludeHeaders {
		if err := w.Write([]string{"PlaceID", "Name", "Category", "TripsCount", "ReviewsCount", "AvgRating"}); err != nil {
			return nil, err
		}
	}
	for _, p := range places {
		row := []string{
			p.PlaceID.String(), p.Name, p.Category,
			fmt.Sprintf("%d", p.TripsCount),
			fmt.Sprintf("%d", p.ReviewsCount),
			fmt.Sprintf("%.2f", p.AvgRating),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), nil
}

func (r *ReportGeneratorService) generateUserActivityJSON(users []domain.UserActivityData, opts domain.ReportOptions) ([]byte, error) {
	if opts.PrettyPrint {
		return json.MarshalIndent(users, "", "  ")
	}
	return json.Marshal(users)
}

func (r *ReportGeneratorService) generateUserActivityCSV(users []domain.UserActivityData, opts domain.ReportOptions) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if opts.IncludeHeaders {
		if err := w.Write([]string{"UserID", "Username", "TripsCreated", "ReviewsWritten", "LastActive"}); err != nil {
			return nil, err
		}
	}
	for _, u := range users {
		row := []string{
			u.UserID.String(), u.Username,
			fmt.Sprintf("%d", u.TripsCreated),
			fmt.Sprintf("%d", u.ReviewsWritten),
			u.LastActive.Format("2006-01-02 15:04:05"),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), nil
}

func (r *ReportGeneratorService) generateTripStatsJSON(trips []domain.TripStatData, opts domain.ReportOptions) ([]byte, error) {
	if opts.PrettyPrint {
		return json.MarshalIndent(trips, "", "  ")
	}
	return json.Marshal(trips)
}

func (r *ReportGeneratorService) generateTripStatsCSV(trips []domain.TripStatData, opts domain.ReportOptions) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if opts.IncludeHeaders {
		if err := w.Write([]string{"TripID", "Title", "UserID", "StartDate", "EndDate", "PlacesCount", "TotalCost"}); err != nil {
			return nil, err
		}
	}
	for _, t := range trips {
		row := []string{
			t.TripID.String(), t.Title, t.UserID.String(),
			t.StartDate.Format("2006-01-02"),
			t.EndDate.Format("2006-01-02"),
			fmt.Sprintf("%d", t.PlacesCount),
			fmt.Sprintf("%.2f", t.TotalCost),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), nil
}
