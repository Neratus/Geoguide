package domain

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type DomainConfig struct {
	MaxReviewCommentLength int `mapstructure:"max_review_comment_length"`
	MinRating              int `mapstructure:"min_rating"`
	MaxRating              int `mapstructure:"max_rating"`

	MaxCityNameLength        int `mapstructure:"max_city_name_length"`
	MaxCityDescriptionLength int `mapstructure:"max_city_description_length"`
	MaxCityTimezoneLength    int `mapstructure:"max_city_timezone_length"`
	MaxCityTravelTipsLength  int `mapstructure:"max_city_travel_tips_length"`

	MaxCountryNameLength int `mapstructure:"max_country_name_length"`
	MaxCurrencyLength    int `mapstructure:"max_currency_length"`
	MaxLanguageLength    int `mapstructure:"max_language_length"`
	MaxPhoneCodeLength   int `mapstructure:"max_phone_code_length"`
	MaxBestSeasonLength  int `mapstructure:"max_best_season_length"`
	MaxReligionLength    int `mapstructure:"max_religion_length"`

	MaxPlaceNameLength        int `mapstructure:"max_place_name_length"`
	MaxPlaceCategoryLength    int `mapstructure:"max_place_category_length"`
	MaxPlaceAddressLength     int `mapstructure:"max_place_address_length"`
	MaxPlaceDescriptionLength int `mapstructure:"max_place_description_length"`
	MaxOpeningHoursLength     int `mapstructure:"max_opening_hours_length"`
	MaxPriceInfoLength        int `mapstructure:"max_price_info_length"`
	MaxContactPhoneLength     int `mapstructure:"max_contact_phone_length"`
	MaxWebsiteLength          int `mapstructure:"max_website_length"`

	MaxTripTitleLength int    `mapstructure:"max_trip_title_length"`
	MaxTripNotesLength int    `mapstructure:"max_trip_notes_length"`
	DefaultTripStatus  string `mapstructure:"default_trip_status"`
	MaxPlacesPerTrip   int    `mapstructure:"max_places_per_trip"`
	SessionTTLHours    int    `mapstructure:"session_ttl_hours"`

	MaxUsernameLength   int `mapstructure:"max_username_length"`
	MaxEmailLength      int `mapstructure:"max_email_length"`
	MinPasswordLength   int `mapstructure:"min_password_length"`
	MaxPasswordLength   int `mapstructure:"max_password_length"`
	MaxPhoneLength      int `mapstructure:"max_phone_length"`
	MaxCountryResidence int `mapstructure:"max_country_residence"`

	MaxStaticPageSlugLength            int `mapstructure:"max_static_page_slug_length"`
	MaxStaticPageTitleLength           int `mapstructure:"max_static_page_title_length"`
	MaxStaticPageContentLength         int `mapstructure:"max_static_page_content_length"`
	MaxStaticPageMetaDescriptionLength int `mapstructure:"max_static_page_meta_description_length"`
	MaxStaticPageImageURLLength        int `mapstructure:"max_static_page_image_url_length"`
	MaxStaticPageImageAltLength        int `mapstructure:"max_static_page_image_alt_length"`

	MaxTripPlaceNotesLength int `mapstructure:"max_trip_place_notes_length"`

	MaxHolidayNameLength int `mapstructure:"max_holiday_name_length"`

	MaxDistrictNameLength        int `mapstructure:"max_district_name_length"`
	MaxDistrictDescriptionLength int `mapstructure:"max_district_description_length"`

	MaxTransportNodeNameLength    int `mapstructure:"max_transport_node_name_length"`
	MaxTransportNodeTypeLength    int `mapstructure:"max_transport_node_type_length"`
	MaxTransportNodeAddressLength int `mapstructure:"max_transport_node_address_length"`
}

var (
	globalConfig *DomainConfig
	configMu     sync.RWMutex
)

func SetConfig(cfg *DomainConfig) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = cfg
}

func GetConfig() *DomainConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	if globalConfig == nil {
		return defaultConfig()
	}
	return globalConfig
}

func defaultConfig() *DomainConfig {
	return &DomainConfig{
		MaxReviewCommentLength:             2000,
		MinRating:                          1,
		MaxRating:                          5,
		MaxCityNameLength:                  255,
		MaxCityDescriptionLength:           2000,
		MaxCityTimezoneLength:              50,
		MaxCityTravelTipsLength:            2000,
		MaxCountryNameLength:               255,
		MaxCurrencyLength:                  3,
		MaxLanguageLength:                  100,
		MaxPhoneCodeLength:                 10,
		MaxBestSeasonLength:                50,
		MaxReligionLength:                  100,
		MaxPlaceNameLength:                 255,
		MaxPlaceCategoryLength:             100,
		MaxPlaceAddressLength:              500,
		MaxPlaceDescriptionLength:          2000,
		MaxOpeningHoursLength:              500,
		MaxPriceInfoLength:                 500,
		MaxContactPhoneLength:              50,
		MaxWebsiteLength:                   255,
		MaxTripTitleLength:                 255,
		MaxTripNotesLength:                 2000,
		DefaultTripStatus:                  "DRAFT",
		MaxPlacesPerTrip:                   50,
		SessionTTLHours:                    24,
		MaxUsernameLength:                  100,
		MaxEmailLength:                     255,
		MinPasswordLength:                  6,
		MaxPasswordLength:                  72,
		MaxPhoneLength:                     50,
		MaxCountryResidence:                100,
		MaxStaticPageSlugLength:            100,
		MaxStaticPageTitleLength:           255,
		MaxStaticPageContentLength:         10000,
		MaxStaticPageMetaDescriptionLength: 500,
		MaxStaticPageImageURLLength:        500,
		MaxStaticPageImageAltLength:        255,
		MaxTripPlaceNotesLength:            2000,
		MaxHolidayNameLength:               255,
		MaxDistrictNameLength:              255,
		MaxDistrictDescriptionLength:       1024,
		MaxTransportNodeNameLength:         255,
		MaxTransportNodeTypeLength:         50,
		MaxTransportNodeAddressLength:      500,
	}
}

type DomainConfigLoader interface {
	LoadAndWatch() error
}

func UpdateConfigFromMap(updates map[string]interface{}) error {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		globalConfig = defaultConfig()
	}

	for key, value := range updates {
		switch key {
		case "MaxReviewCommentLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxReviewCommentLength expects int")
			}
			globalConfig.MaxReviewCommentLength = v
		case "MinRating":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MinRating expects int")
			}
			globalConfig.MinRating = v
		case "MaxRating":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxRating expects int")
			}
			globalConfig.MaxRating = v
		case "MaxCityNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCityNameLength expects int")
			}
			globalConfig.MaxCityNameLength = v
		case "MaxCityDescriptionLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCityDescriptionLength expects int")
			}
			globalConfig.MaxCityDescriptionLength = v
		case "MaxCityTimezoneLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCityTimezoneLength expects int")
			}
			globalConfig.MaxCityTimezoneLength = v
		case "MaxCityTravelTipsLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCityTravelTipsLength expects int")
			}
			globalConfig.MaxCityTravelTipsLength = v
		case "MaxCountryNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCountryNameLength expects int")
			}
			globalConfig.MaxCountryNameLength = v
		case "MaxCurrencyLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCurrencyLength expects int")
			}
			globalConfig.MaxCurrencyLength = v
		case "MaxLanguageLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxLanguageLength expects int")
			}
			globalConfig.MaxLanguageLength = v
		case "MaxPhoneCodeLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPhoneCodeLength expects int")
			}
			globalConfig.MaxPhoneCodeLength = v
		case "MaxBestSeasonLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxBestSeasonLength expects int")
			}
			globalConfig.MaxBestSeasonLength = v
		case "MaxReligionLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxReligionLength expects int")
			}
			globalConfig.MaxReligionLength = v
		case "MaxPlaceNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPlaceNameLength expects int")
			}
			globalConfig.MaxPlaceNameLength = v
		case "MaxPlaceCategoryLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPlaceCategoryLength expects int")
			}
			globalConfig.MaxPlaceCategoryLength = v
		case "MaxPlaceAddressLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPlaceAddressLength expects int")
			}
			globalConfig.MaxPlaceAddressLength = v
		case "MaxPlaceDescriptionLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPlaceDescriptionLength expects int")
			}
			globalConfig.MaxPlaceDescriptionLength = v
		case "MaxOpeningHoursLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxOpeningHoursLength expects int")
			}
			globalConfig.MaxOpeningHoursLength = v
		case "MaxPriceInfoLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPriceInfoLength expects int")
			}
			globalConfig.MaxPriceInfoLength = v
		case "MaxContactPhoneLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxContactPhoneLength expects int")
			}
			globalConfig.MaxContactPhoneLength = v
		case "MaxWebsiteLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxWebsiteLength expects int")
			}
			globalConfig.MaxWebsiteLength = v
		case "MaxTripTitleLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTripTitleLength expects int")
			}
			globalConfig.MaxTripTitleLength = v
		case "MaxTripNotesLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTripNotesLength expects int")
			}
			globalConfig.MaxTripNotesLength = v
		case "DefaultTripStatus":
			v, ok := value.(string)
			if !ok {
				return fmt.Errorf("field DefaultTripStatus expects string")
			}
			globalConfig.DefaultTripStatus = v
		case "MaxPlacesPerTrip":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPlacesPerTrip expects int")
			}
			globalConfig.MaxPlacesPerTrip = v
		case "SessionTTLHours":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field SessionTTLHours expects int")
			}
			globalConfig.SessionTTLHours = v
		case "MaxUsernameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxUsernameLength expects int")
			}
			globalConfig.MaxUsernameLength = v
		case "MaxEmailLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxEmailLength expects int")
			}
			globalConfig.MaxEmailLength = v
		case "MinPasswordLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MinPasswordLength expects int")
			}
			globalConfig.MinPasswordLength = v
		case "MaxPasswordLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPasswordLength expects int")
			}
			globalConfig.MaxPasswordLength = v
		case "MaxPhoneLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxPhoneLength expects int")
			}
			globalConfig.MaxPhoneLength = v
		case "MaxCountryResidence":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxCountryResidence expects int")
			}
			globalConfig.MaxCountryResidence = v
		case "MaxStaticPageSlugLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageSlugLength expects int")
			}
			globalConfig.MaxStaticPageSlugLength = v
		case "MaxStaticPageTitleLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageTitleLength expects int")
			}
			globalConfig.MaxStaticPageTitleLength = v
		case "MaxStaticPageContentLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageContentLength expects int")
			}
			globalConfig.MaxStaticPageContentLength = v
		case "MaxStaticPageMetaDescriptionLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageMetaDescriptionLength expects int")
			}
			globalConfig.MaxStaticPageMetaDescriptionLength = v
		case "MaxStaticPageImageURLLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageImageURLLength expects int")
			}
			globalConfig.MaxStaticPageImageURLLength = v
		case "MaxStaticPageImageAltLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxStaticPageImageAltLength expects int")
			}
			globalConfig.MaxStaticPageImageAltLength = v
		case "MaxTripPlaceNotesLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTripPlaceNotesLength expects int")
			}
			globalConfig.MaxTripPlaceNotesLength = v
		case "MaxHolidayNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxHolidayNameLength expects int")
			}
			globalConfig.MaxHolidayNameLength = v
		case "MaxDistrictNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxDistrictNameLength expects int")
			}
			globalConfig.MaxDistrictNameLength = v
		case "MaxDistrictDescriptionLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxDistrictDescriptionLength expects int")
			}
			globalConfig.MaxDistrictDescriptionLength = v
		case "MaxTransportNodeNameLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTransportNodeNameLength expects int")
			}
			globalConfig.MaxTransportNodeNameLength = v
		case "MaxTransportNodeTypeLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTransportNodeTypeLength expects int")
			}
			globalConfig.MaxTransportNodeTypeLength = v
		case "MaxTransportNodeAddressLength":
			v, ok := value.(int)
			if !ok {
				return fmt.Errorf("field MaxTransportNodeAddressLength expects int")
			}
			globalConfig.MaxTransportNodeAddressLength = v
		default:
			return fmt.Errorf("unknown config field: %s", key)
		}
	}
	return nil
}

type ViperDomainLoader struct {
	path string
	v    *viper.Viper
}

func NewViperDomainLoader(path string) *ViperDomainLoader {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	return &ViperDomainLoader{
		path: path,
		v:    v,
	}
}

func (l *ViperDomainLoader) Load() error {
	if err := l.v.ReadInConfig(); err != nil {
		slog.Warn("Domain config not found, using defaults", "error", err)
		return nil
	}
	var cfg DomainConfig
	if err := l.v.Unmarshal(&cfg); err != nil {
		slog.Error("Failed to unmarshal domain config", "error", err)
		return err
	}
	SetConfig(&cfg)
	slog.Info("Domain config loaded", "path", l.path)
	return nil
}

func (l *ViperDomainLoader) StartPolling(interval time.Duration) {
	_ = l.Load()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := l.v.ReadInConfig(); err != nil {
				slog.Warn("Failed to re-read domain config", "error", err)
				continue
			}
			var newCfg DomainConfig
			if err := l.v.Unmarshal(&newCfg); err != nil {
				slog.Error("Failed to unmarshal domain config after reload", "error", err)
				continue
			}
			SetConfig(&newCfg)
			slog.Info("Domain config reloaded via polling", "interval", interval)
		}
	}()
}
