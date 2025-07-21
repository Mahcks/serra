import { api } from './api';
import type { AnalyticsOverview, DriveAlertsResponse, DriveUsageHistory } from '../types/analytics';

export const analyticsApi = {
  // Get comprehensive analytics overview
  getOverview: async (): Promise<AnalyticsOverview> => {
    const response = await api.get('/analytics/overview');
    return response.data;
  },

  // Get active drive alerts
  getDriveAlerts: async (): Promise<DriveAlertsResponse> => {
    const response = await api.get('/analytics/drive/alerts');
    return response.data;
  },

  // Acknowledge a drive alert
  acknowledgeDriveAlert: async (alertId: number): Promise<{ message: string }> => {
    const response = await api.post(`/analytics/drive/alerts/${alertId}/acknowledge`);
    return response.data;
  },

  // Get drive usage history
  getDriveUsageHistory: async (driveId: string, days: number = 30): Promise<DriveUsageHistory> => {
    const response = await api.get(`/analytics/drive/${driveId}/history?days=${days}`);
    return response.data;
  },

  // Get request analytics
  getRequestAnalytics: async (days: number = 30, limit: number = 10): Promise<any> => {
    const response = await api.get(`/analytics/requests?days=${days}&limit=${limit}`);
    return response.data;
  },

  // Get request trends
  getRequestTrends: async (days: number = 30): Promise<any> => {
    const response = await api.get(`/analytics/requests/trends?days=${days}`);
    return response.data;
  },

  // Get watch analytics from Jellystat
  getWatchAnalytics: async (limit: number = 10): Promise<any> => {
    const response = await api.get(`/analytics/watch?limit=${limit}`);
    return response.data;
  },

  // Get failure analysis
  getFailureAnalysis: async (days: number = 30): Promise<any> => {
    const response = await api.get(`/analytics/requests/failures?days=${days}`);
    return response.data;
  },

  // Get content availability vs requests
  getContentAvailability: async (days: number = 30, limit: number = 10): Promise<any> => {
    const response = await api.get(`/analytics/requests/availability?days=${days}&limit=${limit}`);
    return response.data;
  },
};