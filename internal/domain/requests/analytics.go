package requests

import "github.com/Neratus/geoguide/internal/domain"

type GenerateReportRequest struct {
	ReportType string
	Format     domain.ReportFormat
	Options    domain.ReportOptions
	DateFrom   string
	DateTo     string
	Limit      int
}
