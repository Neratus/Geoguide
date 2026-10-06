import apiClient from "./client";
import type { Review, CreateReviewRequest, ReviewBriefResponse } from "../types/api";

export const createReview = (placeId: string, data: CreateReviewRequest) =>
    apiClient.post<Review>(`/places/${placeId}/reviews`, data);

export const getReviewsByPlace = (placeId: string, params?: { onlyApproved?: boolean; limit?: number; offset?: number }) =>
    apiClient.get<Review[]>(`/places/${placeId}/reviews`, { params });

export const getUserReviews = (params?: { limit?: number; offset?: number }) =>
    apiClient.get<Review[]>("/users/me/reviews", { params });

export const deleteReview = (reviewId: string) =>
    apiClient.delete(`/reviews/${reviewId}`);

export const getPendingReviews = (params?: { isApproved?: boolean; limit?: number; offset?: number }) =>
    apiClient.get<ReviewBriefResponse[]>('/moderation/reviews', { params }).then(res => res.data);

export const moderateReview = (reviewId: string, data: { approved: boolean; comment?: string }) =>
    apiClient.post('/moderation/reviews', { reviewId, ...data });