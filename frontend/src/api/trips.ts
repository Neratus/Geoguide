import apiClient from "./client";
import type { Trip, TripDetail, TripPlace } from "../types/api";

export const createTrip = (data: { title: string; startDate: string; endDate: string; budget?: number; notes?: string }) =>
    apiClient.post<Trip>("/trips", data);

export const getUserTrips = () =>
    apiClient.get<Trip[]>("/trips");

export const getTrip = (tripId: string) =>
    apiClient.get<TripDetail>(`/trips/${tripId}`);

export const updateTrip = (tripId: string, data: Partial<Omit<Trip, "id">>) =>
    apiClient.put<Trip>(`/trips/${tripId}`, data);

export const deleteTrip = (tripId: string) =>
    apiClient.delete(`/trips/${tripId}`);

export const addPlaceToTrip = (tripId: string, data: { placeId: string; dayNumber: number; arrivalTime?: string; durationMin?: number; notes?: string }) =>
    apiClient.post(`/trips/${tripId}/places`, data);

export const removePlaceFromTrip = (tripId: string, placeId: string) =>
    apiClient.delete(`/trips/${tripId}/places`, { params: { placeId } });