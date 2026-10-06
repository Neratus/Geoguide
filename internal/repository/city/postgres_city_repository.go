package city_repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCityRepo struct {
	Pool    *pgxpool.Pool
	Queries *postgreSQL.Queries
}

func NewPostgresCityRepo(cfg *config.Config) (*PostgresCityRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresCityRepo{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresCityRepo) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresCityRepo) Save(ctx context.Context, city *domain.City) error {
	cityParams := toSaveCityParams(city)
	id, err := r.Queries.SaveCity(ctx, cityParams)
	if err != nil {
		return fmt.Errorf("failed to save city: %w", err)
	}
	city.SetId(domain.FromPgUUID(id))
	return err
}

func (r *PostgresCityRepo) FindByID(ctx context.Context, id domain.CityID) (*domain.City, error) {
	dbcity, err := r.Queries.FindCityByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find city by ID: %w", err)
	}
	return toDomainCity(dbcity)
}

func (r *PostgresCityRepo) FindByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.City, error) {
	dbcity, err := r.Queries.FindCitiesByCountry(ctx, domain.ToPgUUID(countryID))
	if err != nil {
		return nil, fmt.Errorf("failed to find city by ID: %w", err)
	}
	var result []*domain.City
	for _, c := range dbcity {
		city, err := toDomainCity(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert city: %w", err)
		}
		result = append(result, city)
	}
	return result, nil
}

func (r *PostgresCityRepo) FindByName(ctx context.Context, name string) (*domain.City, error) {
	dbcity, err := r.Queries.FindCityByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find city by name: %w", err)
	}
	return toDomainCity(dbcity)
}

func (r *PostgresCityRepo) Update(ctx context.Context, city *domain.City) error {
	params := postgreSQL.UpdateCityParams{
		ID:         domain.ToPgUUID(city.GetId()),
		Name:       city.GetName(),
		Population: pgtype.Int8{Int64: city.GetPopulation(), Valid: true},
		IsCapital:  pgtype.Bool{Bool: city.IsCapital(), Valid: true},
		Coordinates: pgtype.Point{
			P:     pgtype.Vec2{X: city.GetCoordinates().Lng(), Y: city.GetCoordinates().Lat()},
			Valid: true,
		},
		Description: pgtype.Text{String: city.GetDescription(), Valid: true},
		Timezone:    pgtype.Text{String: city.GetTimezone(), Valid: true},
		TravelTips:  pgtype.Text{String: city.GetTravelTips(), Valid: true},
		CountryID:   domain.ToPgUUID(city.GetCountryId()),
		ImageID:     domain.ToPgUUID(city.GetImageId()),
	}
	err := r.Queries.UpdateCity(ctx, params)
	return err
}

func (r *PostgresCityRepo) Delete(ctx context.Context, id domain.CityID) error {
	err := r.Queries.DeleteCity(ctx, domain.ToPgUUID(id))
	return err
}

func (r *PostgresCityRepo) SaveDistrict(ctx context.Context, district *domain.CityDistrict) error {
	districtParams := toSaveDistrictParams(district)
	id, err := r.Queries.SaveDistrict(ctx, districtParams)
	if err != nil {
		return fmt.Errorf("failed to save district: %w", err)
	}
	district.SetId(domain.FromPgUUID(id))

	return err
}

func (r *PostgresCityRepo) FindDistrictsByCity(ctx context.Context, cityID domain.CityID) ([]*domain.CityDistrict, error) {
	dbdistrict, err := r.Queries.FindDistrictsByCity(ctx, domain.ToPgUUID(cityID))
	if err != nil {
		return nil, fmt.Errorf("failed to find district by city ID: %w", err)
	}
	var result []*domain.CityDistrict
	for _, c := range dbdistrict {
		city, err := toDomainDistrict(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert city: %w", err)
		}
		result = append(result, city)
	}
	return result, nil
}

func (r *PostgresCityRepo) UpdateDistrict(ctx context.Context, district *domain.CityDistrict) error {
	params := postgreSQL.UpdateDistrictParams{
		ID:   domain.ToPgUUID(district.GetId()),
		Name: district.GetName(),
		Coordinates: pgtype.Point{
			P:     pgtype.Vec2{X: district.GetCoordinates().Lng(), Y: district.GetCoordinates().Lat()},
			Valid: true,
		},
		Description: pgtype.Text{String: district.GetDescription(), Valid: true},
		CityID:      domain.ToPgUUID(district.GetCityId()),
		ImageID:     domain.ToPgUUID(district.GetImageId()),
	}
	err := r.Queries.UpdateDistrict(ctx, params)
	return err
}

func (r *PostgresCityRepo) DeleteDistrict(ctx context.Context, id domain.CityDistrictID) error {
	err := r.Queries.DeleteDistrict(ctx, domain.ToPgUUID(id))
	return err
}

func (r *PostgresCityRepo) SaveTransportNode(ctx context.Context, node *domain.TransportNode) error {
	nodeParams := toSaveTransportNodeParams(node)
	id, err := r.Queries.SaveTransportNode(ctx, nodeParams)
	if err != nil {
		return fmt.Errorf("failed to save node: %w", err)
	}
	node.SetId(domain.FromPgUUID(id))

	return err
}

func (r *PostgresCityRepo) FindTransportNodeByID(ctx context.Context, id domain.NodeID) (*domain.TransportNode, error) {
	dbnode, err := r.Queries.FindTransportNodeByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find transport node by ID: %w", err)
	}
	return toDomainTransportNode(dbnode)
}

func (r *PostgresCityRepo) FindTransportNodesByCity(ctx context.Context, cityID domain.CityID) ([]*domain.TransportNode, error) {
	dbnodes, err := r.Queries.FindTransportNodesByCity(ctx, domain.ToPgUUID(cityID))
	if err != nil {
		return nil, fmt.Errorf("failed to find transport node by city ID: %w", err)
	}
	var result []*domain.TransportNode
	for _, c := range dbnodes {
		node, err := toDomainTransportNode(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node: %w", err)
		}
		result = append(result, node)
	}
	return result, nil
}

func (r *PostgresCityRepo) UpdateTransportNode(ctx context.Context, node *domain.TransportNode) error {
	params := postgreSQL.UpdateTransportNodeParams{
		ID:   domain.ToPgUUID(node.GetId()),
		Name: node.GetName(),
		Type: node.GetNodeType(),
		Coordinates: pgtype.Point{
			P:     pgtype.Vec2{X: node.GetCoordinates().Lng(), Y: node.GetCoordinates().Lat()},
			Valid: true,
		},
		Address: pgtype.Text{String: node.GetAddress(), Valid: true},
		CityID:  domain.ToPgUUID(node.GetCityID()),
		ImageID: domain.ToPgUUID(node.GetImageID()),
	}
	err := r.Queries.UpdateTransportNode(ctx, params)
	return err
}

func (r *PostgresCityRepo) SearchCities(ctx context.Context, query string, limit, offset int) ([]*domain.City, error) {
	dbCities, err := r.Queries.SearchCities(ctx, postgreSQL.SearchCitiesParams{
		Column1: query,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search cities: %w", err)
	}

	cities := make([]*domain.City, 0, len(dbCities))
	for _, dbc := range dbCities {
		city, err := toDomainCity(dbc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert city: %w", err)
		}
		cities = append(cities, city)
	}
	return cities, nil
}

func (r *PostgresCityRepo) DeleteTransportNode(ctx context.Context, id domain.NodeID) error {
	err := r.Queries.DeleteTransportNode(ctx, domain.ToPgUUID(id))
	return err
}

func toSaveCityParams(city *domain.City) postgreSQL.SaveCityParams {
	var countryIDParam pgtype.UUID
	if city.GetCountryId() == uuid.Nil {
		countryIDParam = pgtype.UUID{Valid: false}
	} else {
		countryIDParam = pgtype.UUID{Bytes: city.GetCountryId(), Valid: true}
	}

	var imageIDParam pgtype.UUID
	if city.GetImageId() == uuid.Nil {
		imageIDParam = pgtype.UUID{Valid: false}
	} else {
		imageIDParam = pgtype.UUID{Bytes: city.GetImageId(), Valid: true}
	}

	coords := city.GetCoordinates()
	return postgreSQL.SaveCityParams{
		Name:       city.GetName(),
		Population: pgtype.Int8{Int64: city.GetPopulation(), Valid: true},
		IsCapital:  pgtype.Bool{Bool: city.IsCapital(), Valid: true},
		Coordinates: pgtype.Point{
			P:     pgtype.Vec2{X: coords.Lng(), Y: coords.Lat()},
			Valid: true,
		},
		Description: pgtype.Text{String: city.GetDescription(), Valid: true},
		Timezone:    pgtype.Text{String: city.GetTimezone(), Valid: true},
		TravelTips:  pgtype.Text{String: city.GetTravelTips(), Valid: true},
		CountryID:   countryIDParam,
		ImageID:     imageIDParam,
	}
}

func toDomainCity(dbCity postgreSQL.City) (*domain.City, error) {
	if !dbCity.CountryID.Valid {
		return nil, errors.New("country_id is null")
	}
	countryID := uuid.UUID(dbCity.CountryID.Bytes)

	imageID := uuid.Nil
	if dbCity.ImageID.Valid {
		imageID = uuid.UUID(dbCity.ImageID.Bytes)
	}

	coords := domain.NewCoordinates(dbCity.Coordinates.P.Y, dbCity.Coordinates.P.X)

	return domain.NewCityFromDB(
		dbCity.ID.Bytes,
		dbCity.Name,
		dbCity.Population.Int64,
		dbCity.IsCapital.Bool,
		coords,
		dbCity.Description.String,
		dbCity.Timezone.String,
		dbCity.TravelTips.String,
		countryID,
		imageID,
	), nil
}

func toSaveTransportNodeParams(node *domain.TransportNode) postgreSQL.SaveTransportNodeParams {
	coords := node.GetCoordinates()
	return postgreSQL.SaveTransportNodeParams{
		Name:        node.GetName(),
		Type:        string(node.GetNodeType()),
		Coordinates: pgtype.Point{P: pgtype.Vec2{X: coords.Lng(), Y: coords.Lat()}, Valid: true},
		Address:     pgtype.Text{String: node.GetAddress(), Valid: true},
		CityID:      domain.ToPgUUID(node.GetCityID()),
		ImageID:     domain.ToPgUUID(node.GetImageID()),
	}
}

func toDomainTransportNode(dbNode postgreSQL.TransportNode) (*domain.TransportNode, error) {
	id := domain.FromPgUUID(dbNode.ID)
	if id == uuid.Nil {
		return nil, errors.New("transport node id is null")
	}
	cityID := domain.FromPgUUID(dbNode.CityID)
	imageID := domain.FromPgUUID(dbNode.ImageID)

	coords := domain.NewCoordinates(dbNode.Coordinates.P.Y, dbNode.Coordinates.P.X)

	return domain.NewTransportNodeFromDB(
		id,
		dbNode.Name,
		dbNode.Type,
		coords,
		dbNode.Address.String,
		cityID,
		imageID,
	), nil
}

func toSaveDistrictParams(district *domain.CityDistrict) postgreSQL.SaveDistrictParams {
	coords := district.GetCoordinates()
	return postgreSQL.SaveDistrictParams{
		Name:        district.GetName(),
		Coordinates: pgtype.Point{P: pgtype.Vec2{X: coords.Lng(), Y: coords.Lat()}, Valid: true},
		Description: pgtype.Text{String: district.GetDescription(), Valid: true},
		CityID:      domain.ToPgUUID(district.GetCityId()),
		ImageID:     domain.ToPgUUID(district.GetImageId()),
	}
}

func toDomainDistrict(dbDistrict postgreSQL.CityDistrict) (*domain.CityDistrict, error) {
	id := domain.FromPgUUID(dbDistrict.ID)
	if id == uuid.Nil {
		return nil, errors.New("district id is null")
	}
	cityID := domain.FromPgUUID(dbDistrict.CityID)
	imageID := domain.FromPgUUID(dbDistrict.ImageID)

	coords := domain.NewCoordinates(dbDistrict.Coordinates.P.Y, dbDistrict.Coordinates.P.X)

	return domain.NewCityDistrictFromDB(
		id,
		dbDistrict.Name,
		dbDistrict.Description.String,
		coords,
		cityID,
		imageID,
	), nil
}
