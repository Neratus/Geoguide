-- StaticPage
CREATE TABLE IF NOT EXISTS static_page_by_id (
    id UUID PRIMARY KEY,
    slug TEXT,
    title TEXT,
    content TEXT,
    meta_description TEXT,
    image_url TEXT,
    image_alt TEXT,
    updated_at TIMESTAMP,
    published_at TIMESTAMP,
    is_published BOOLEAN
);

CREATE TABLE IF NOT EXISTS static_page_by_slug (
    slug TEXT PRIMARY KEY,
    id UUID
);

-- Country
CREATE TABLE IF NOT EXISTS country_by_id (
    id UUID PRIMARY KEY,
    name TEXT,
    area DOUBLE,
    population BIGINT,
    gdp DOUBLE,
    currency TEXT,
    visa_requirements TEXT,
    description TEXT,
    safety_tips TEXT,
    best_season TEXT,
    language TEXT,
    phone_code TEXT,
    religion TEXT,
    capital_id UUID,
    image_id UUID
);

CREATE TABLE IF NOT EXISTS country_by_name (
    name TEXT PRIMARY KEY,
    id UUID
);

-- Holiday
CREATE TABLE IF NOT EXISTS holiday_by_country (
    country_id UUID,
    date DATE,
    id UUID,
    name TEXT,
    description TEXT,
    traditions TEXT,
    history TEXT,
    is_national BOOLEAN,
    image_id UUID,
    PRIMARY KEY (country_id, date, id)
) WITH CLUSTERING ORDER BY (date ASC);

CREATE TABLE IF NOT EXISTS holiday_by_date (
    date DATE,
    country_id UUID,
    id UUID,
    name TEXT,
    description TEXT,
    is_national BOOLEAN,
    PRIMARY KEY (date, country_id, id)
);

-- City
CREATE TABLE IF NOT EXISTS city_by_id (
    id UUID PRIMARY KEY,
    name TEXT,
    population BIGINT,
    is_capital BOOLEAN,
    coordinates TEXT,
    description TEXT,
    timezone TEXT,
    travel_tips TEXT,
    country_id UUID,
    image_id UUID
);

CREATE TABLE IF NOT EXISTS city_by_country (
    country_id UUID,
    name TEXT,
    id UUID,
    population BIGINT,
    is_capital BOOLEAN,
    coordinates TEXT,
    description TEXT,
    timezone TEXT,
    travel_tips TEXT,
    image_id UUID,
    PRIMARY KEY (country_id, name, id)
) WITH CLUSTERING ORDER BY (name ASC);

CREATE TABLE IF NOT EXISTS city_by_name (
    name TEXT,
    id UUID,
    country_id UUID,
    PRIMARY KEY (name, id)
);

-- CityDistrict
CREATE TABLE IF NOT EXISTS district_by_city (
    city_id UUID,
    id UUID,
    name TEXT,
    description TEXT,
    coordinates TEXT,
    image_id UUID,
    PRIMARY KEY (city_id, id)
);

-- TransportNode
CREATE TABLE IF NOT EXISTS transport_node_by_city (
    city_id UUID,
    type TEXT,
    id UUID,
    name TEXT,
    coordinates TEXT,
    address TEXT,
    image_id UUID,
    PRIMARY KEY (city_id, type, id)
) WITH CLUSTERING ORDER BY (type ASC);

-- Place
CREATE TABLE IF NOT EXISTS place_by_id (
    id UUID PRIMARY KEY,
    name TEXT,
    category TEXT,
    description TEXT,
    coordinates TEXT,
    address TEXT,
    opening_hours TEXT,
    price_info TEXT,
    avg_visit_duration_min INT,
    avg_rating DOUBLE,
    reviews_count INT,
    contact_phone TEXT,
    website TEXT,
    city_id UUID,
    city_name TEXT,
    district_id UUID,
    district_name TEXT,
    image_id UUID,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS place_by_city (
    city_id UUID,
    avg_rating DOUBLE,
    id UUID,
    name TEXT,
    category TEXT,
    description TEXT,
    coordinates TEXT,
    address TEXT,
    opening_hours TEXT,
    price_info TEXT,
    avg_visit_duration_min INT,
    reviews_count INT,
    contact_phone TEXT,
    website TEXT,
    district_id UUID,
    district_name TEXT,
    image_id UUID,
    created_at TIMESTAMP,
    PRIMARY KEY (city_id, avg_rating, id)
) WITH CLUSTERING ORDER BY (avg_rating DESC);

CREATE TABLE IF NOT EXISTS place_by_category (
    category TEXT,
    avg_rating DOUBLE,
    id UUID,
    name TEXT,
    description TEXT,
    city_id UUID,
    city_name TEXT,
    image_id UUID,
    PRIMARY KEY (category, avg_rating, id)
) WITH CLUSTERING ORDER BY (avg_rating DESC);

-- User
CREATE TABLE IF NOT EXISTS user_by_id (
    id UUID PRIMARY KEY,
    role TEXT,
    username TEXT,
    email TEXT,
    password_hash TEXT,
    phone TEXT,
    registered_at TIMESTAMP,
    birth_date DATE,
    country_of_residence TEXT,
    avatar_url TEXT,
    favorite_categories SET<TEXT>,
    is_blocked BOOLEAN,
    blocked_at TIMESTAMP,
    block_reason TEXT,
    email_verified BOOLEAN,
    phone_verified BOOLEAN,
    two_factor_enabled BOOLEAN,
    two_factor_secret TEXT,
    backup_codes SET<TEXT>
);

CREATE TABLE IF NOT EXISTS user_by_username (
    username TEXT PRIMARY KEY,
    id UUID
);

CREATE TABLE IF NOT EXISTS user_by_email (
    email TEXT PRIMARY KEY,
    id UUID
);

-- Review
CREATE TABLE IF NOT EXISTS review_by_place (
    place_id UUID,
    created_at TIMESTAMP,
    id UUID,
    user_id UUID,
    rating INT,
    comment TEXT,
    visit_date DATE,
    username TEXT,
    is_approved BOOLEAN,
    is_moderated BOOLEAN,
    moderation_comment TEXT,
    image_id UUID,
    PRIMARY KEY (place_id, created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC);

CREATE TABLE IF NOT EXISTS review_by_user (
    user_id UUID,
    created_at TIMESTAMP,
    id UUID,
    place_id UUID,
    rating INT,
    comment TEXT,
    visit_date DATE,
    is_approved BOOLEAN,
    is_moderated BOOLEAN,
    PRIMARY KEY (user_id, created_at, id)
);

CREATE TABLE IF NOT EXISTS review_pending (
    created_at TIMESTAMP,
    id UUID,
    rating INT,
    comment TEXT,
    visit_date DATE,
    user_id UUID,
    place_id UUID,
    username TEXT,
    PRIMARY KEY ((created_at), id)
) WITH CLUSTERING ORDER BY (id ASC);

-- Trip
CREATE TABLE IF NOT EXISTS trip_by_id (
    id UUID PRIMARY KEY,
    title TEXT,
    start_date DATE,
    end_date DATE,
    budget DOUBLE,
    status TEXT,
    notes TEXT,
    created_at TIMESTAMP,
    user_id UUID,
    image_id UUID
);

CREATE TABLE IF NOT EXISTS trip_by_user (
    user_id UUID,
    start_date DATE,
    id UUID,
    title TEXT,
    end_date DATE,
    budget DOUBLE,
    status TEXT,
    notes TEXT,
    created_at TIMESTAMP,
    image_id UUID,
    PRIMARY KEY (user_id, start_date, id)
) WITH CLUSTERING ORDER BY (start_date DESC);

-- TripPlace
CREATE TABLE IF NOT EXISTS trip_place_by_trip (
    trip_id UUID,
    day_number INT,
    place_id UUID,
    id UUID,
    arrival_time BIGINT,
    duration_min INT,
    notes TEXT,
    visit_status TEXT,
    actual_cost DOUBLE,
    PRIMARY KEY (trip_id, day_number, place_id)
) WITH CLUSTERING ORDER BY (day_number ASC);

-- User Favourite Places
CREATE TABLE IF NOT EXISTS favourite_by_user (
    user_id UUID,
    favorited_at TIMESTAMP,
    place_id UUID,
    name TEXT,
    category TEXT,
    description TEXT,
    coordinates TEXT,
    address TEXT,
    opening_hours TEXT,
    price_info TEXT,
    avg_visit_duration_min INT,
    avg_rating DOUBLE,
    reviews_count INT,
    contact_phone TEXT,
    website TEXT,
    city_id UUID,
    city_name TEXT,
    district_id UUID,
    district_name TEXT,
    image_id UUID,
    PRIMARY KEY (user_id, favorited_at, place_id)
) WITH CLUSTERING ORDER BY (favorited_at DESC);


CREATE TABLE IF NOT EXISTS review_by_id (
    id UUID PRIMARY KEY,
    place_id UUID,
    user_id UUID,
    rating INT,
    comment TEXT,
    visit_date DATE,
    created_at TIMESTAMP,
    is_moderated BOOLEAN,
    is_approved BOOLEAN,
    moderation_comment TEXT,
    image_id UUID,
    username TEXT
);
