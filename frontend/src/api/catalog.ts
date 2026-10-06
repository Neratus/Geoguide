import apiClient from "./client";
import type {
    Country,
    CountryDetail,
    City,
    CityDetail,
    District,
    TransportNode,
    Place,
    PlaceDetail,
} from "../types/api";

export const getCountries = (params?: { limit?: number; offset?: number; search?: string }) =>
    apiClient.get<Country[]>("/countries", { params });

export const getCountry = (id: string) =>
    apiClient.get<CountryDetail>(`/countries/${id}`);

export const getCitiesByCountry = (countryId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<City[]>(`/countries/${countryId}/cities`, { params });

export const getCity = (id: string) =>
    apiClient.get<CityDetail>(`/cities/${id}`);

export const searchCities = (query: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<City[]>("/cities", { params: { query, ...params } });

export const getDistrictsByCity = (cityId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<District[]>(`/cities/${cityId}/districts`, { params });

export const getTransportNodes = (cityId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<TransportNode[]>(`/cities/${cityId}/transports`, { params });

export const getPlace = (id: string) => apiClient.get<PlaceDetail>(`/places/${id}`);

export const getPlacesByCity = (
    cityId: string,
    params?: { category?: string; limit?: number; offset?: number; sortBy?: "rating" | "name" | "visits" }
) =>
    apiClient.get<Place[]>("/places", { params: { cityId, ...params } });

export const getPlacesByCategory = (category: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<Place[]>("/places", { params: { category, ...params } });

export const getPlaces = (params?: {
    cityId?: string;
    category?: string;
    sortBy?: "rating" | "name" | "visits";
    limit?: number;
    offset?: number;
}) => apiClient.get<Place[]>("/places", { params });