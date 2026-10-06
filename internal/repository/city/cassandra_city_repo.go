package city_repo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type CassandraCityRepo struct {
	session *gocql.Session
}

func NewCassandraCityRepo(session *gocql.Session) *CassandraCityRepo {
	return &CassandraCityRepo{session: session}
}

func (r *CassandraCityRepo) Close() error {
	return nil
}

func (r *CassandraCityRepo) Save(ctx context.Context, city *domain.City) error {
	id := city.GetId()
	if id == uuid.Nil {
		id = domain.CityID(uuid.New())
		city.SetId(id)
	}
	cassID := gocql.UUID(id)

	coords := city.GetCoordinates()
	coordStr := fmt.Sprintf("%f,%f", coords.Lat(), coords.Lng())
	countryID := gocql.UUID(city.GetCountryId())
	imageID := gocql.UUID(city.GetImageId())

	if err := r.session.Query(`
		INSERT INTO city_by_id (
			id, name, population, is_capital, coordinates, description,
			timezone, travel_tips, country_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, city.GetName(), city.GetPopulation(), city.IsCapital(),
		coordStr, city.GetDescription(), city.GetTimezone(), city.GetTravelTips(),
		countryID, imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert city_by_id: %w", err)
	}

	if err := r.session.Query(`
		INSERT INTO city_by_country (
			country_id, name, id, population, is_capital, coordinates,
			description, timezone, travel_tips, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, countryID, city.GetName(), cassID, city.GetPopulation(),
		city.IsCapital(), coordStr, city.GetDescription(), city.GetTimezone(),
		city.GetTravelTips(), imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM city_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert city_by_country: %w", err)
	}

	if err := r.session.Query(`
		INSERT INTO city_by_name (name, id, country_id) VALUES (?, ?, ?)
	`, city.GetName(), cassID, countryID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM city_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		_ = r.session.Query(`DELETE FROM city_by_country WHERE country_id = ? AND name = ?`, countryID, city.GetName()).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert city_by_name: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) FindByID(ctx context.Context, id domain.CityID) (*domain.City, error) {
	var (
		cassID      gocql.UUID
		name        string
		population  int64
		isCapital   bool
		coordStr    string
		description string
		timezone    string
		travelTips  string
		countryID   gocql.UUID
		imageID     gocql.UUID
	)
	query := r.session.Query(`
        SELECT id, name, population, is_capital, coordinates, description,
               timezone, travel_tips, country_id, image_id
        FROM city_by_id WHERE id = ? LIMIT 1
    `, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &name, &population, &isCapital, &coordStr,
		&description, &timezone, &travelTips, &countryID, &imageID) {
		_ = iter.Close()
		return nil, fmt.Errorf("")
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	coords := parseCoordinates(coordStr)
	city := domain.NewCityFromDB(
		domain.CityID(cassID), name, population, isCapital, coords,
		description, timezone, travelTips, domain.CountryID(countryID),
		domain.ImageID(imageID),
	)
	return city, nil
}

func (r *CassandraCityRepo) FindByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.City, error) {
	var cities []*domain.City
	iter := r.session.Query(`
		SELECT name, id, population, is_capital, coordinates, description,
		       timezone, travel_tips, image_id
		FROM city_by_country WHERE country_id = ?
	`, gocql.UUID(countryID)).WithContext(ctx).Iter()
	var (
		name        string
		cassID      gocql.UUID
		population  int64
		isCapital   bool
		coordStr    string
		description string
		timezone    string
		travelTips  string
		imageID     gocql.UUID
	)
	for iter.Scan(&name, &cassID, &population, &isCapital, &coordStr,
		&description, &timezone, &travelTips, &imageID) {
		coords := parseCoordinates(coordStr)
		city := domain.NewCityFromDB(
			domain.CityID(cassID), name, population, isCapital, coords,
			description, timezone, travelTips, countryID,
			domain.ImageID(imageID),
		)
		cities = append(cities, city)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return cities, nil
}

func (r *CassandraCityRepo) FindByName(ctx context.Context, name string) (*domain.City, error) {
	var id gocql.UUID
	var countryID gocql.UUID
	if err := r.session.Query(`SELECT id, country_id FROM city_by_name WHERE name = ? LIMIT 1`, name).WithContext(ctx).Scan(&id, &countryID); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("")
		}
		return nil, err
	}
	return r.FindByID(ctx, domain.CityID(id))
}

func (r *CassandraCityRepo) Update(ctx context.Context, city *domain.City) error {
	id := gocql.UUID(city.GetId())
	countryID := gocql.UUID(city.GetCountryId())
	coordStr := fmt.Sprintf("%f,%f", city.GetCoordinates().Lat(), city.GetCoordinates().Lng())
	imageID := gocql.UUID(city.GetImageId())

	if err := r.session.Query(`
		UPDATE city_by_id SET
			name = ?, population = ?, is_capital = ?, coordinates = ?,
			description = ?, timezone = ?, travel_tips = ?, country_id = ?, image_id = ?
		WHERE id = ?
	`, city.GetName(), city.GetPopulation(), city.IsCapital(), coordStr,
		city.GetDescription(), city.GetTimezone(), city.GetTravelTips(),
		countryID, imageID, id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update city_by_id: %w", err)
	}

	if err := r.session.Query(`
		UPDATE city_by_country SET
			population = ?, is_capital = ?, coordinates = ?, description = ?,
			timezone = ?, travel_tips = ?, image_id = ?
		WHERE country_id = ? AND name = ? AND id = ?
	`, city.GetPopulation(), city.IsCapital(), coordStr,
		city.GetDescription(), city.GetTimezone(), city.GetTravelTips(),
		imageID, countryID, city.GetName(), id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update city_by_country: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) Delete(ctx context.Context, id domain.CityID) error {
	city, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if city == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	if err := r.session.Query(`DELETE FROM city_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM city_by_country WHERE country_id = ? AND name = ? AND id = ?`, gocql.UUID(city.GetCountryId()), city.GetName(), cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM city_by_name WHERE name = ?`, city.GetName()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraCityRepo) SearchCities(ctx context.Context, query string, limit, offset int) ([]*domain.City, error) {
	var cities []*domain.City
	iter := r.session.Query(`SELECT id FROM city_by_id`).WithContext(ctx).Iter()
	var id gocql.UUID
	for iter.Scan(&id) {
		if len(cities) >= limit && limit > 0 {
			break
		}
		city, err := r.FindByID(ctx, domain.CityID(id))
		if err != nil {
			continue
		}
		if city == nil {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(city.GetName()), strings.ToLower(query)) {
			cities = append(cities, city)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return cities, nil
}

func (r *CassandraCityRepo) SaveDistrict(ctx context.Context, district *domain.CityDistrict) error {
	id := district.GetId()
	if id == uuid.Nil {
		id = domain.CityDistrictID(uuid.New())
		district.SetId(id)
	}
	cassID := gocql.UUID(id)
	coordStr := fmt.Sprintf("%f,%f", district.GetCoordinates().Lat(), district.GetCoordinates().Lng())

	if err := r.session.Query(`
		INSERT INTO district_by_city (city_id, id, name, description, coordinates, image_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, gocql.UUID(district.GetCityId()), cassID, district.GetName(),
		district.GetDescription(), coordStr, gocql.UUID(district.GetImageId())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert district: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) FindDistrictsByCity(ctx context.Context, cityID domain.CityID) ([]*domain.CityDistrict, error) {
	var districts []*domain.CityDistrict
	iter := r.session.Query(`
		SELECT id, name, description, coordinates, image_id
		FROM district_by_city WHERE city_id = ?
	`, gocql.UUID(cityID)).WithContext(ctx).Iter()
	var (
		id       gocql.UUID
		name     string
		desc     string
		coordStr string
		imageID  gocql.UUID
	)
	for iter.Scan(&id, &name, &desc, &coordStr, &imageID) {
		coords := parseCoordinates(coordStr)
		district := domain.NewCityDistrictFromDB(
			domain.CityDistrictID(id), name, desc, coords,
			cityID, domain.ImageID(imageID),
		)
		districts = append(districts, district)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return districts, nil
}

func (r *CassandraCityRepo) UpdateDistrict(ctx context.Context, district *domain.CityDistrict) error {
	coordStr := fmt.Sprintf("%f,%f", district.GetCoordinates().Lat(), district.GetCoordinates().Lng())
	if err := r.session.Query(`
		UPDATE district_by_city SET
			name = ?, description = ?, coordinates = ?, image_id = ?
		WHERE city_id = ? AND id = ?
	`, district.GetName(), district.GetDescription(), coordStr,
		gocql.UUID(district.GetImageId()), gocql.UUID(district.GetCityId()),
		gocql.UUID(district.GetId())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update district: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) DeleteDistrict(ctx context.Context, id domain.CityDistrictID) error {
	var cityID gocql.UUID
	iter := r.session.Query(`SELECT city_id FROM district_by_city WHERE id = ? LIMIT 1`, gocql.UUID(id)).WithContext(ctx).Iter()
	if !iter.Scan(&cityID) {
		_ = iter.Close()
		return nil
	}
	if err := iter.Close(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM district_by_city WHERE city_id = ? AND id = ?`, cityID, gocql.UUID(id)).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraCityRepo) SaveTransportNode(ctx context.Context, node *domain.TransportNode) error {
	id := node.GetId()
	if id == uuid.Nil {
		id = domain.NodeID(uuid.New())
		node.SetId(id)
	}
	cassID := gocql.UUID(id)
	coordStr := fmt.Sprintf("%f,%f", node.GetCoordinates().Lat(), node.GetCoordinates().Lng())
	if err := r.session.Query(`
		INSERT INTO transport_node_by_city (city_id, type, id, name, coordinates, address, image_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, gocql.UUID(node.GetCityID()), node.GetNodeType(), cassID, node.GetName(),
		coordStr, node.GetAddress(), gocql.UUID(node.GetImageID())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert transport node: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) FindTransportNodeByID(ctx context.Context, id domain.NodeID) (*domain.TransportNode, error) {
	var cityID gocql.UUID
	var nodeType string
	iter := r.session.Query(`SELECT city_id, type FROM transport_node_by_city WHERE id = ? LIMIT 1`, gocql.UUID(id)).WithContext(ctx).Iter()
	if !iter.Scan(&cityID, &nodeType) {
		_ = iter.Close()
		return nil, fmt.Errorf("")
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	var (
		name     string
		coordStr string
		address  string
		imageID  gocql.UUID
	)
	query := r.session.Query(`
		SELECT name, coordinates, address, image_id
		FROM transport_node_by_city WHERE city_id = ? AND type = ? AND id = ?
	`, cityID, nodeType, gocql.UUID(id))
	if err := query.WithContext(ctx).Scan(&name, &coordStr, &address, &imageID); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("")
		}
		return nil, err
	}
	coords := parseCoordinates(coordStr)
	node := domain.NewTransportNodeFromDB(
		id, name, nodeType, coords, address,
		domain.CityID(cityID), domain.ImageID(imageID),
	)
	return node, nil
}

func (r *CassandraCityRepo) FindTransportNodesByCity(ctx context.Context, cityID domain.CityID) ([]*domain.TransportNode, error) {
	var nodes []*domain.TransportNode
	iter := r.session.Query(`
		SELECT id, name, type, coordinates, address, image_id
		FROM transport_node_by_city WHERE city_id = ?
	`, gocql.UUID(cityID)).WithContext(ctx).Iter()
	var (
		id       gocql.UUID
		name     string
		nodeType string
		coordStr string
		address  string
		imageID  gocql.UUID
	)
	for iter.Scan(&id, &name, &nodeType, &coordStr, &address, &imageID) {
		coords := parseCoordinates(coordStr)
		node := domain.NewTransportNodeFromDB(
			domain.NodeID(id), name, nodeType, coords, address,
			cityID, domain.ImageID(imageID),
		)
		nodes = append(nodes, node)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *CassandraCityRepo) UpdateTransportNode(ctx context.Context, node *domain.TransportNode) error {
	coordStr := fmt.Sprintf("%f,%f", node.GetCoordinates().Lat(), node.GetCoordinates().Lng())
	if err := r.session.Query(`
		UPDATE transport_node_by_city SET name = ?, coordinates = ?, address = ?, image_id = ?
		WHERE city_id = ? AND type = ? AND id = ?
	`, node.GetName(), coordStr, node.GetAddress(), gocql.UUID(node.GetImageID()),
		gocql.UUID(node.GetCityID()), node.GetNodeType(), gocql.UUID(node.GetId())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update transport node: %w", err)
	}
	return nil
}

func (r *CassandraCityRepo) DeleteTransportNode(ctx context.Context, id domain.NodeID) error {
	var cityID gocql.UUID
	var nodeType string
	iter := r.session.Query(`SELECT city_id, type FROM transport_node_by_city WHERE id = ? LIMIT 1`, gocql.UUID(id)).WithContext(ctx).Iter()
	if !iter.Scan(&cityID, &nodeType) {
		_ = iter.Close()
		return nil
	}
	if err := iter.Close(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM transport_node_by_city WHERE city_id = ? AND type = ? AND id = ?`, cityID, nodeType, gocql.UUID(id)).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func parseCoordinates(coordStr string) domain.Coordinates {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return domain.Coordinates{}
	}
	lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return domain.NewCoordinates(lat, lng)
}
