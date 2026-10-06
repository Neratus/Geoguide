package country_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type CassandraCountryRepo struct {
	session *gocql.Session
}

func NewCassandraCountryRepo(session *gocql.Session) *CassandraCountryRepo {
	return &CassandraCountryRepo{session: session}
}

func (r *CassandraCountryRepo) Close() error {
	return nil
}

func (r *CassandraCountryRepo) Save(ctx context.Context, country *domain.Country) error {
	id := country.GetId()
	if id == uuid.Nil {
		id = domain.CountryID(uuid.New())
		country.SetId(id)
	}
	cassID := gocql.UUID(id)
	name := country.GetName()
	capitalID := gocql.UUID(country.GetCapitalID())
	imageID := gocql.UUID(country.GetImageID())

	if err := r.session.Query(`
		INSERT INTO country_by_id (
			id, name, area, population, gdp, currency, visa_requirements,
			description, safety_tips, best_season, language, phone_code, religion,
			capital_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, name, country.GetArea(), country.GetPopulation(), country.GetGdp(),
		country.GetCurrency(), country.GetVisaRequirements(), country.GetDescription(),
		country.GetSafetyTips(), country.GetBestSeason(), country.GetLanguage(),
		country.GetPhoneCode(), country.GetReligion(), capitalID, imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert country_by_id: %w", err)
	}

	if err := r.session.Query(`INSERT INTO country_by_name (name, id) VALUES (?, ?)`, name, cassID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM country_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert country_by_name: %w", err)
	}
	return nil
}

func (r *CassandraCountryRepo) FindByID(ctx context.Context, id domain.CountryID) (*domain.Country, error) {
	var (
		cassID      gocql.UUID
		name        string
		area        float64
		population  int64
		gdp         float64
		currency    string
		visaReq     string
		description string
		safetyTips  string
		bestSeason  string
		language    string
		phoneCode   string
		religion    string
		capitalID   gocql.UUID
		imageID     gocql.UUID
	)
	query := r.session.Query(`
		SELECT id, name, area, population, gdp, currency, visa_requirements,
		       description, safety_tips, best_season, language, phone_code, religion,
		       capital_id, image_id
		FROM country_by_id WHERE id = ? LIMIT 1
	`, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &name, &area, &population, &gdp, &currency, &visaReq,
		&description, &safetyTips, &bestSeason, &language, &phoneCode, &religion,
		&capitalID, &imageID) {
		_ = iter.Close()
		return nil, fmt.Errorf("")
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	country := domain.NewCountryFromDB(
		domain.CountryID(cassID), name, area, gdp, population, currency,
		visaReq, description, safetyTips, bestSeason, language, phoneCode,
		religion, domain.CountryID(capitalID), domain.ImageID(imageID),
	)
	return country, nil
}

func (r *CassandraCountryRepo) FindAll(ctx context.Context) ([]*domain.Country, error) {
	var countries []*domain.Country
	iter := r.session.Query(`SELECT id FROM country_by_id`).WithContext(ctx).Iter()
	var id gocql.UUID
	for iter.Scan(&id) {
		country, err := r.FindByID(ctx, domain.CountryID(id))
		if err != nil {
			continue
		}
		if country != nil {
			countries = append(countries, country)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return countries, nil
}

func (r *CassandraCountryRepo) FindByName(ctx context.Context, name string) (*domain.Country, error) {
	var id gocql.UUID
	if err := r.session.Query(`SELECT id FROM country_by_name WHERE name = ? LIMIT 1`, name).WithContext(ctx).Scan(&id); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("")
		}
		return nil, err
	}
	return r.FindByID(ctx, domain.CountryID(id))
}

func (r *CassandraCountryRepo) Update(ctx context.Context, country *domain.Country) error {
	id := gocql.UUID(country.GetId())
	capitalID := gocql.UUID(country.GetCapitalID())
	imageID := gocql.UUID(country.GetImageID())
	if err := r.session.Query(`
		UPDATE country_by_id SET
			name = ?, area = ?, population = ?, gdp = ?, currency = ?,
			visa_requirements = ?, description = ?, safety_tips = ?,
			best_season = ?, language = ?, phone_code = ?, religion = ?,
			capital_id = ?, image_id = ?
		WHERE id = ?
	`, country.GetName(), country.GetArea(), country.GetPopulation(), country.GetGdp(),
		country.GetCurrency(), country.GetVisaRequirements(), country.GetDescription(),
		country.GetSafetyTips(), country.GetBestSeason(), country.GetLanguage(),
		country.GetPhoneCode(), country.GetReligion(), capitalID, imageID, id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update country_by_id: %w", err)
	}
	return nil
}

func (r *CassandraCountryRepo) Delete(ctx context.Context, id domain.CountryID) error {
	country, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if country == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	if err := r.session.Query(`DELETE FROM country_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM country_by_name WHERE name = ?`, country.GetName()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraCountryRepo) SaveHoliday(ctx context.Context, holiday *domain.Holiday) error {
	id := holiday.GetId()
	if id == uuid.Nil {
		id = domain.HolidayID(uuid.New())
		holiday.SetId(id)
	}
	cassID := gocql.UUID(id)
	countryID := gocql.UUID(holiday.GetCountryID())
	imageID := gocql.UUID(holiday.GetImageID())
	date := holiday.GetDate()

	if err := r.session.Query(`
		INSERT INTO holiday_by_country (country_id, date, id, name, description,
			traditions, history, is_national, image_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, countryID, date, cassID, holiday.GetName(), holiday.GetDescription(),
		holiday.GetTraditions(), holiday.GetHistory(), holiday.IsNational(), imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert holiday_by_country: %w", err)
	}
	if err := r.session.Query(`
		INSERT INTO holiday_by_date (date, country_id, id, name, description, is_national)
		VALUES (?, ?, ?, ?, ?, ?)
	`, date, countryID, cassID, holiday.GetName(), holiday.GetDescription(), holiday.IsNational()).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM holiday_by_country WHERE country_id = ? AND date = ? AND id = ?`, countryID, date, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert holiday_by_date: %w", err)
	}
	if err := r.ensureHolidayByIDTable(ctx); err != nil {
		return err
	}
	if err := r.session.Query(`
		INSERT INTO holiday_by_id (id, country_id, date, name, description, traditions, history, is_national, image_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, countryID, date, holiday.GetName(), holiday.GetDescription(),
		holiday.GetTraditions(), holiday.GetHistory(), holiday.IsNational(), imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM holiday_by_country WHERE country_id = ? AND date = ? AND id = ?`, countryID, date, cassID).WithContext(ctx).Exec()
		_ = r.session.Query(`DELETE FROM holiday_by_date WHERE date = ? AND country_id = ? AND id = ?`, date, countryID, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert holiday_by_id: %w", err)
	}
	return nil
}

func (r *CassandraCountryRepo) ensureHolidayByIDTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS holiday_by_id (
		id UUID PRIMARY KEY,
		country_id UUID,
		date DATE,
		name TEXT,
		description TEXT,
		traditions TEXT,
		history TEXT,
		is_national BOOLEAN,
		image_id UUID
	)`
	return r.session.Query(query).WithContext(ctx).Exec()
}

func (r *CassandraCountryRepo) FindHolidayByID(ctx context.Context, id domain.HolidayID) (*domain.Holiday, error) {
	if err := r.ensureHolidayByIDTable(ctx); err != nil {
		return nil, err
	}
	var (
		cassID      gocql.UUID
		countryID   gocql.UUID
		date        time.Time
		name        string
		description string
		traditions  string
		history     string
		isNational  bool
		imageID     gocql.UUID
	)
	query := r.session.Query(`SELECT id, country_id, date, name, description, traditions, history, is_national, image_id FROM holiday_by_id WHERE id = ? LIMIT 1`, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &countryID, &date, &name, &description, &traditions, &history, &isNational, &imageID) {
		_ = iter.Close()
		return nil, nil
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	holiday := domain.NewHolidayFromDB(
		domain.HolidayID(cassID), name, date, description, traditions, history,
		isNational, domain.CountryID(countryID), domain.ImageID(imageID),
	)
	return holiday, nil
}

func (r *CassandraCountryRepo) FindHolidaysByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.Holiday, error) {
	var holidays []*domain.Holiday
	iter := r.session.Query(`
		SELECT date, id, name, description, traditions, history, is_national, image_id
		FROM holiday_by_country WHERE country_id = ?
	`, gocql.UUID(countryID)).WithContext(ctx).Iter()
	var (
		date        time.Time
		id          gocql.UUID
		name        string
		description string
		traditions  string
		history     string
		isNational  bool
		imageID     gocql.UUID
	)
	for iter.Scan(&date, &id, &name, &description, &traditions, &history, &isNational, &imageID) {
		holiday := domain.NewHolidayFromDB(
			domain.HolidayID(id), name, date, description, traditions, history,
			isNational, countryID, domain.ImageID(imageID),
		)
		holidays = append(holidays, holiday)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return holidays, nil
}

func (r *CassandraCountryRepo) FindHolidaysByDate(ctx context.Context, dateStr string) ([]*domain.Holiday, error) {
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	var holidays []*domain.Holiday
	iter := r.session.Query(`
		SELECT country_id, id, name, description, is_national
		FROM holiday_by_date WHERE date = ?
	`, parsedDate).WithContext(ctx).Iter()
	var (
		countryID   gocql.UUID
		id          gocql.UUID
		name        string
		description string
		isNational  bool
	)
	for iter.Scan(&countryID, &id, &name, &description, &isNational) {
		holiday := domain.NewHolidayFromDB(
			domain.HolidayID(id), name, parsedDate, description, "", "", isNational,
			domain.CountryID(countryID), domain.ImageID(uuid.Nil),
		)
		holidays = append(holidays, holiday)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return holidays, nil
}

func (r *CassandraCountryRepo) UpdateHoliday(ctx context.Context, holiday *domain.Holiday) error {
	if err := r.DeleteHoliday(ctx, holiday.GetId()); err != nil {
		return err
	}
	return r.SaveHoliday(ctx, holiday)
}

func (r *CassandraCountryRepo) DeleteHoliday(ctx context.Context, id domain.HolidayID) error {
	holiday, err := r.FindHolidayByID(ctx, id)
	if err != nil {
		return err
	}
	if holiday == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	countryID := gocql.UUID(holiday.GetCountryID())
	date := holiday.GetDate()

	if err := r.session.Query(`DELETE FROM holiday_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM holiday_by_country WHERE country_id = ? AND date = ? AND id = ?`, countryID, date, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM holiday_by_date WHERE date = ? AND country_id = ? AND id = ?`, date, countryID, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}
