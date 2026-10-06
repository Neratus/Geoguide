package domain

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CountryID = uuid.UUID
type CityID = uuid.UUID
type PlaceID = uuid.UUID
type UserID = uuid.UUID
type TripID = uuid.UUID
type ReviewID = uuid.UUID
type HolidayID = uuid.UUID
type NodeID = uuid.UUID
type CityDistrictID = uuid.UUID
type ImageID = uuid.UUID
type TripPlaceID = uuid.UUID
type StaticPageID = uuid.UUID

func Int64ToUUID(id int64) uuid.UUID {
	var u uuid.UUID
	binary.BigEndian.PutUint64(u[8:], uint64(id))
	return u
}

func UUIDToInt64(u uuid.UUID) int64 {
	return int64(binary.BigEndian.Uint64(u[8:]))
}

func Uint64ToUUID(id uint64) uuid.UUID {
	var u uuid.UUID
	binary.BigEndian.PutUint64(u[8:], id)
	return u
}

func UUIDToUint64(u uuid.UUID) uint64 {
	return binary.BigEndian.Uint64(u[8:])
}

func ToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: id != uuid.Nil,
	}
}

func FromPgUUID(pg pgtype.UUID) uuid.UUID {
	if !pg.Valid {
		return uuid.Nil
	}
	return uuid.UUID(pg.Bytes)
}

type Coordinates struct {
	lat float64
	lng float64
}

func (c *Coordinates) String() string {
	return fmt.Sprintf("%f,%f", c.lat, c.lng)
}

func (c Coordinates) Lat() float64 { return c.lat }
func (c Coordinates) Lng() float64 { return c.lng }

func (c *Coordinates) SetLat(lat float64) {
	c.lat = lat
}

func (c *Coordinates) SetLng(lng float64) {
	c.lng = lng
}

func NewCoordinates(lat, lng float64) Coordinates {
	return Coordinates{lat: lat, lng: lng}
}

func validateCoordinates(c Coordinates) error {
	lat := c.Lat()
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	lng := c.Lng()
	if lng < -180 || lng > 180 {
		return ErrInvalidLongitude
	}
	return nil
}
