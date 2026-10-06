export interface RegisterRequest {
    username: string;
    email: string;
    password: string;
    phone?: string;
    birthDate?: string;
    countryOfResidence?: string;
}
export interface RegisterResponse {
    id: string;
    username: string;
    email: string;
    message: string;
}

export interface LoginRequest {
    username: string;
    password: string;
}
export interface LoginResponse {
    id: string;
    username: string;
    email: string;
    token?: string;
    requiresTwoFactor?: boolean;
    challengeId?: string;
}

export interface TwoFactorVerifyRequest {
    challengeId: string;
    code: string;
}
export interface TwoFactorResponse {
    token: string;
}



export interface ProfileResponse {
    id: string;
    username: string;
    email: string;
    phone: string;
    registeredAt: string;
    birthDate?: string;
    countryOfResidence: string;
    avatarUrl: string;
    favoriteCategories: string[];
    isBlocked: boolean;
    role: string;
}

export interface Country {
    id: string;
    name: string;
    capital?: string;
    area?: number;
    population?: number;
    currency?: string;
    language?: string;
    phoneCode?: string;
    imageUrl?: string;
}

export interface CountryDetail extends Country {
    gdp?: number;
    visaRequirements?: string;
    description?: string;
    safetyTips?: string;
    bestSeason?: string;
    religion?: string;
    capitalId?: string;
    capitalName?: string;
    holidays?: Holiday[];
}

export interface Holiday {
    id: string;
    name: string;
    date: string;
    description?: string;
    isNational?: boolean;
}

export interface City {
    id: string;
    name: string;
    population?: number;
    isCapital?: boolean;
    coordinates?: Coordinates;
    description?: string;
    timezone?: string;
    imageUrl?: string;
}

export interface CityDetail extends City {
    travelTips?: string;
    countryId: string;
    imageId?: string;
}

export interface District {
    id: string;
    name: string;
    description?: string;
    coordinates?: Coordinates;
    cityId: string;
    imageId?: string;
}

export interface TransportNode {
    id: string;
    name: string;
    nodeType: "airport" | "train_station" | "bus_station" | "port";
    coordinates?: Coordinates;
    address?: string;
    cityId: string;
    imageId?: string;
}

export interface Place {
    id: string;
    name: string;
    category: string;
    description?: string;
    coordinates?: string;
    address?: string;
    openingHours?: string;
    priceInfo?: string;
    avgVisitDurationMin?: number;
    avgRating?: number;
    reviewsCount?: number;
    contactPhone?: string;
    website?: string;
    cityId: string;
    cityName?: string;
    districtId?: string;
    districtName?: string;
    imageUrl?: string;
}

export interface PlaceDetail extends Place {
    reviews?: ReviewBrief[];
}

export interface ReviewBrief {
    id: string;
    rating: number;
    comment: string;
    visitDate?: string;
    createdAt?: string;
    username?: string;
    isApproved?: boolean;
}

export interface Trip {
    id: string;
    title: string;
    startDate: string;
    endDate: string;
    budget?: number;
    status: "DRAFT" | "PLANNED" | "ONGOING" | "FINISHED" | "CANCELED";
    notes?: string;
}

export interface TripDetail extends Trip {
    places?: TripPlace[];
}

export interface TripPlace {
    placeId: string;
    dayNumber: number;
    arrivalTime?: string;
    durationMin?: number;
    notes?: string;
    visitStatus: "PLANNED" | "VISITED" | "SKIPPED";
    actualCost?: number;
}

export interface Review {
    id: string;
    rating: number;
    comment: string;
    visitDate: string;
    createdAt?: string;
    isModerated?: boolean;
    isApproved?: boolean;
    moderationComment?: string;
    userId?: string;
    placeId?: string;
    username?: string;
    userAvatar?: string;
    placeName?: string;
}

export interface CreateReviewRequest {
    rating: number;
    comment: string;
    visitDate: string;
    imageUrl?: string;
}

export interface ReportRequest {
    reportType: "popular_places" | "user_activity" | "trip_statistics";
    format: "json" | "csv";
    limit?: number;
    dateFrom?: string;
    dateTo?: string;
}

export interface PopularPlace {
    placeId: string;
    name: string;
    category: string;
    tripsCount: number;
    reviewsCount: number;
    avgRating: number;
}

export interface UserActivity {
    userId: string;
    username: string;
    tripsCreated: number;
    reviewsWritten: number;
    lastActive: string;
}

export interface TripStat {
    tripId: string;
    title: string;
    userId: string;
    startDate: string;
    endDate: string;
    placesCount: number;
    totalCost: number;
}

export interface UserResponse {
    id: string;
    username: string;
    email: string;
    isBlocked: boolean;
    registeredAt?: string;
}

export interface ReviewBriefResponse {
    id: string;
    rating: number;
    comment: string;
    visitDate?: string;
    createdAt?: string;
    username: string;
    isApproved: boolean;
}

export interface Coordinates {
    lat: number;
    lng: number;
}