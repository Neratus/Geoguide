import apiClient from "./client";
import type { ReportRequest, PopularPlace, UserActivity, TripStat } from "../types/api";

export const generateReport = (data: { reportType: string; format: string; limit?: number; dateFrom?: string; dateTo?: string }) =>
    apiClient.post('/analytics/reports', data).then(res => res.data);