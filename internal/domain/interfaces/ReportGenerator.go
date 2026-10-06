package interfaces

import "github.com/Neratus/geoguide/internal/domain"

type ReportGenerator interface {
	GeneratePopularPlacesReport(places []domain.PlaceReportData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error)
	GenerateUserActivityReport(users []domain.UserActivityData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error)
	GenerateTripStatisticsReport(trips []domain.TripStatData, format domain.ReportFormat, opts domain.ReportOptions) ([]byte, error)
	GenerateReport(data interface{}, config domain.ReportConfig) ([]byte, error)
}
