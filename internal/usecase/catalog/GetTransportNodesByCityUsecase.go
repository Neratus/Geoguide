package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type GetTransportNodesByCityUseCase struct {
	cityRepo interfaces.CityRepository
	logger   *slog.Logger
}

func NewGetTransportNodesByCityUseCase(cityRepo interfaces.CityRepository, logger *slog.Logger) *GetTransportNodesByCityUseCase {
	return &GetTransportNodesByCityUseCase{cityRepo: cityRepo, logger: logger}
}

func (uc *GetTransportNodesByCityUseCase) Execute(ctx context.Context, req dto.GetTransportNodesByCityRequest) ([]dto.TransportNodeResponse, error) {
	uc.logger.Info("getting transport nodes by city", "city_id", req.CityID)
	nodes, err := uc.cityRepo.FindTransportNodesByCity(ctx, req.CityID)
	if err != nil {
		return nil, err
	}
	start := req.Offset
	if start > len(nodes) {
		start = len(nodes)
	}
	end := start + req.Limit
	if end > len(nodes) {
		end = len(nodes)
	}
	result := make([]dto.TransportNodeResponse, 0, end-start)
	for _, n := range nodes[start:end] {
		result = append(result, dto.TransportNodeResponse{
			ID:          n.GetId(),
			Name:        n.GetName(),
			NodeType:    string(n.GetNodeType()),
			Coordinates: n.GetCoordinates(),
			Address:     n.GetAddress(),
			CityID:      n.GetCityID(),
			ImageID:     n.GetImageID(),
		})
	}
	return result, nil
}
