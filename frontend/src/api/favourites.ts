import apiClient from "./client";
import type { Place } from "../types/api";

export const getFavourites = (params?: { limit?: number; offset?: number }) =>
    apiClient.get<Place[]>("/favourites", { params }).then(res => res.data);

export const addFavourite = (placeId: string) =>
    apiClient.post("/favourites", { placeId });

export const removeFavourite = (placeId: string) =>
    apiClient.delete("/favourites", { data: { placeId } });