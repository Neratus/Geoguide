import apiClient from "./client";
import type {
  RegisterRequest,
  RegisterResponse,
  LoginRequest,
  LoginResponse,
  TwoFactorVerifyRequest,
  TwoFactorResponse,
  ProfileResponse,
  UserResponse
} from "../types/api";



export const register = (data: RegisterRequest) =>
  apiClient.post<RegisterResponse>("/auth/register", data);

export const login = (data: LoginRequest) =>
  apiClient.post<LoginResponse>("/auth/login", data);

export const verifyTwoFactor = (data: TwoFactorVerifyRequest) =>
  apiClient.post<TwoFactorResponse>("/auth/2fa/verify", data);

export const getProfile = () =>
  apiClient.get<ProfileResponse>("/auth/profile");

export const updateProfile = (data: Partial<ProfileResponse>) =>
  apiClient.patch<ProfileResponse>("/auth/profile", data);

export const logout = () => apiClient.post("/auth/logout");

export const verifyContact = (data: { userId: string; contact: string; code: string }) =>
  apiClient.post("/auth/verify", data);

export const getUsers = (params?: { limit?: number; offset?: number; search?: string }) =>
  apiClient.get<UserResponse[]>('/moderation/users', { params }).then(res => res.data);

export const moderateUser = (userId: string, data: { block: boolean; reason?: string }) =>
  apiClient.post('/moderation/users', { userId, ...data });