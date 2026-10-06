package country_repo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCountryRepo struct {
	Pool    *pgxpool.Pool
	Queries *postgreSQL.Queries
}

func NewPostgresCountryRepo(cfg *config.Config) (*PostgresCountryRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresCountryRepo{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresCountryRepo) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresCountryRepo) Save(ctx context.Context, country *domain.Country) error {
	countryParams := toSaveCountryParams(country)

	id, err := r.Queries.SaveCountry(ctx, countryParams)
	if err != nil {
		return fmt.Errorf("failed to save country: %w", err)
	}
	country.SetId(domain.FromPgUUID(id))
	return err
}

func (r *PostgresCountryRepo) FindByID(ctx context.Context, id domain.CountryID) (*domain.Country, error) {
	dbcountry, err := r.Queries.FindCountryByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find country by ID: %w", err)
	}
	return toDomainCountry(dbcountry)
}

func (r *PostgresCountryRepo) FindHolidayByID(ctx context.Context, id domain.HolidayID) (*domain.Holiday, error) {
	dbholiday, err := r.Queries.FindHolidayByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find holiday by ID: %w", err)
	}
	return ToDomainHoliday(dbholiday)
}

func (r *PostgresCountryRepo) FindAll(ctx context.Context) ([]*domain.Country, error) {
	dbcountries, err := r.Queries.GetCountries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get countries by ID: %w", err)
	}
	var result []*domain.Country
	for _, c := range dbcountries {
		country, err := toDomainCountry(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert country: %w", err)
		}
		result = append(result, country)
	}
	return result, nil
}

func (r *PostgresCountryRepo) FindByName(ctx context.Context, name string) (*domain.Country, error) {
	dbcountry, err := r.Queries.FindCountryByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find country by name: %w", err)
	}
	return toDomainCountry(dbcountry)
}

func (r *PostgresCountryRepo) Update(ctx context.Context, country *domain.Country) error {
	params := postgreSQL.UpdateCountryParams{
		ID:               domain.ToPgUUID(country.GetId()),
		Name:             country.GetName(),
		Area:             float64ToNumeric(country.GetArea()),
		Population:       pgtype.Int8{Int64: country.GetPopulation(), Valid: true},
		Gdp:              float64ToNumeric(country.GetGdp()),
		Currency:         pgtype.Text{String: country.GetCurrency(), Valid: true},
		VisaRequirements: pgtype.Text{String: country.GetVisaRequirements(), Valid: true},
		Description:      pgtype.Text{String: country.GetDescription(), Valid: true},
		SafetyTips:       pgtype.Text{String: country.GetSafetyTips(), Valid: true},
		BestSeason:       pgtype.Text{String: country.GetBestSeason(), Valid: true},
		Language:         pgtype.Text{String: country.GetLanguage(), Valid: true},
		PhoneCode:        pgtype.Text{String: country.GetPhoneCode(), Valid: true},
		Religion:         pgtype.Text{String: country.GetReligion(), Valid: true},
		CapitalID:        domain.ToPgUUID(country.GetCapitalID()),
		ImageID:          domain.ToPgUUID(country.GetImageID()),
	}

	err := r.Queries.UpdateCountry(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update country: %w", err)
	}
	return nil
}

func (r *PostgresCountryRepo) Delete(ctx context.Context, id domain.CountryID) error {
	err := r.Queries.DeleteCountry(ctx, domain.ToPgUUID(id))
	return err
}

func (r *PostgresCountryRepo) SaveHoliday(ctx context.Context, holiday *domain.Holiday) error {
	holidayParams := ToSaveHolidayParams(holiday)
	id, err := r.Queries.SaveHoliday(ctx, holidayParams)
	if err != nil {
		return fmt.Errorf("failed to save holiday: %w", err)
	}
	holiday.SetId(domain.FromPgUUID(id))
	return err
}

func (r *PostgresCountryRepo) FindHolidaysByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.Holiday, error) {
	dbholidays, err := r.Queries.FindHolidaysByCountry(ctx, domain.ToPgUUID(countryID))
	if err != nil {
		return nil, fmt.Errorf("failed to find holidays by country by ID: %w", err)
	}
	var result []*domain.Holiday
	for _, c := range dbholidays {
		holiday, err := ToDomainHoliday(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert holiday: %w", err)
		}
		result = append(result, holiday)
	}
	return result, nil
}

func (r *PostgresCountryRepo) FindHolidaysByDate(ctx context.Context, date string) ([]*domain.Holiday, error) {
	parsedDate, _ := time.Parse("2006-01-02", date)
	pgDate := pgtype.Date{Time: parsedDate, Valid: true}
	dbholidays, err := r.Queries.FindHolidaysByDate(ctx, pgDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find holidays by date by ID: %w", err)
	}
	var result []*domain.Holiday
	for _, c := range dbholidays {
		holiday, err := ToDomainHoliday(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert holiday: %w", err)
		}
		result = append(result, holiday)
	}
	return result, nil
}

func (r *PostgresCountryRepo) UpdateHoliday(ctx context.Context, holiday *domain.Holiday) error {
	params := postgreSQL.UpdateHolidayParams{
		ID:          domain.ToPgUUID(holiday.GetId()),
		Name:        holiday.GetName(),
		Date:        pgtype.Date{Time: holiday.GetDate(), Valid: true},
		Description: pgtype.Text{String: holiday.GetDescription(), Valid: true},
		Traditions:  pgtype.Text{String: holiday.GetTraditions(), Valid: true},
		History:     pgtype.Text{String: holiday.GetHistory(), Valid: true},
		IsNational:  pgtype.Bool{Bool: holiday.IsNational(), Valid: true},
		CountryID:   domain.ToPgUUID(holiday.GetCountryID()),
		ImageID:     domain.ToPgUUID(holiday.GetImageID()),
	}
	err := r.Queries.UpdateHoliday(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update holiday: %w", err)
	}
	return nil
}

func (r *PostgresCountryRepo) DeleteHoliday(ctx context.Context, id domain.HolidayID) error {
	err := r.Queries.DeleteHoliday(ctx, domain.ToPgUUID(id))
	return err
}

func toSaveCountryParams(country *domain.Country) postgreSQL.SaveCountryParams {
	var capitalIDParam pgtype.UUID
	if country.GetCapitalID() != uuid.Nil {
		capitalIDParam = pgtype.UUID{Bytes: country.GetCapitalID(), Valid: true}
	} else {
		capitalIDParam = pgtype.UUID{Valid: false}
	}

	return postgreSQL.SaveCountryParams{
		Name:             country.GetName(),
		Area:             float64ToNumeric(country.GetArea()),
		Population:       pgtype.Int8{Int64: country.GetPopulation(), Valid: true},
		Gdp:              float64ToNumeric(country.GetGdp()),
		Currency:         pgtype.Text{String: country.GetCurrency(), Valid: true},
		VisaRequirements: pgtype.Text{String: country.GetVisaRequirements(), Valid: true},
		Description:      pgtype.Text{String: country.GetDescription(), Valid: true},
		SafetyTips:       pgtype.Text{String: country.GetSafetyTips(), Valid: true},
		BestSeason:       pgtype.Text{String: country.GetBestSeason(), Valid: true},
		Language:         pgtype.Text{String: country.GetLanguage(), Valid: true},
		PhoneCode:        pgtype.Text{String: country.GetPhoneCode(), Valid: true},
		Religion:         pgtype.Text{String: country.GetReligion(), Valid: true},
		CapitalID:        capitalIDParam,
		ImageID:          domain.ToPgUUID(country.GetImageID()),
	}
}

func toDomainCountry(dbCountry postgreSQL.Country) (*domain.Country, error) {
	id := domain.FromPgUUID(dbCountry.ID)
	if id == uuid.Nil {
		return nil, errors.New("country id is null")
	}
	capitalID := domain.FromPgUUID(dbCountry.CapitalID)
	imageID := domain.FromPgUUID(dbCountry.ImageID)

	area := numericToFloat64(dbCountry.Area)
	gdp := numericToFloat64(dbCountry.Gdp)

	population := dbCountry.Population.Int64

	return domain.NewCountryFromDB(
		id,
		dbCountry.Name,
		area,
		gdp,
		population,
		dbCountry.Currency.String,
		dbCountry.VisaRequirements.String,
		dbCountry.Description.String,
		dbCountry.SafetyTips.String,
		dbCountry.BestSeason.String,
		dbCountry.Language.String,
		dbCountry.PhoneCode.String,
		dbCountry.Religion.String,
		capitalID,
		imageID,
	), nil
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

func ToSaveHolidayParams(holiday *domain.Holiday) postgreSQL.SaveHolidayParams {
	var imageIDParam pgtype.UUID
	if holiday.GetImageID() != uuid.Nil {
		imageIDParam = pgtype.UUID{Bytes: holiday.GetImageID(), Valid: true}
	} else {
		imageIDParam = pgtype.UUID{Valid: false}
	}

	return postgreSQL.SaveHolidayParams{
		Name:        holiday.GetName(),
		Date:        pgtype.Date{Time: holiday.GetDate(), Valid: true},
		Description: pgtype.Text{String: holiday.GetDescription(), Valid: true},
		Traditions:  pgtype.Text{String: holiday.GetTraditions(), Valid: true},
		History:     pgtype.Text{String: holiday.GetHistory(), Valid: true},
		IsNational:  pgtype.Bool{Bool: holiday.IsNational(), Valid: true},
		CountryID:   domain.ToPgUUID(holiday.GetCountryID()),
		ImageID:     imageIDParam,
	}
}

func ToDomainHoliday(dbHoliday postgreSQL.Holiday) (*domain.Holiday, error) {
	id := domain.FromPgUUID(dbHoliday.ID)
	if id == uuid.Nil {
		return nil, errors.New("holiday id is null")
	}
	countryID := domain.FromPgUUID(dbHoliday.CountryID)
	imageID := domain.FromPgUUID(dbHoliday.ImageID)

	return domain.NewHolidayFromDB(
		id,
		dbHoliday.Name,
		dbHoliday.Date.Time,
		dbHoliday.Description.String,
		dbHoliday.Traditions.String,
		dbHoliday.History.String,
		dbHoliday.IsNational.Bool,
		countryID,
		imageID,
	), nil
}

func float64ToNumeric(f float64) pgtype.Numeric {
	if math.IsNaN(f) {
		return pgtype.Numeric{NaN: true, Valid: true}
	}
	if math.IsInf(f, 1) {
		return pgtype.Numeric{InfinityModifier: pgtype.Infinity, Valid: true}
	}
	if math.IsInf(f, -1) {
		return pgtype.Numeric{InfinityModifier: pgtype.NegativeInfinity, Valid: true}
	}

	s := strconv.FormatFloat(f, 'f', -1, 64)

	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{Valid: true}
	}
	return n
}
