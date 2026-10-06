package requests

import "github.com/Neratus/geoguide/internal/domain"

type GetTransportNodesByCityRequest struct {
	CityID domain.CityID
	Limit  int
	Offset int
}

type TransportNodeResponse struct {
	ID          domain.NodeID
	Name        string
	NodeType    string
	Coordinates domain.Coordinates
	Address     string
	CityID      domain.CityID
	ImageID     domain.ImageID
}
