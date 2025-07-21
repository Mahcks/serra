import type { StoragePool } from './index';

export interface DriveAlert {
  id: number;
  drive_id: string;
  drive_name: string;
  mount_path: string;
  alert_type: 'usage_threshold' | 'growth_rate' | 'projected_full';
  threshold_value: number;
  current_value: number;
  alert_message: string;
  is_active: boolean;
  last_triggered: string;
  acknowledge_count: number;
}

export interface DriveUsageData {
  drive_id?: string;
  name?: string;
  mount_path?: string;
  total_size: number;
  used_size: number;
  available_size: number;
  usage_percentage: number;
  growth_rate_gb_per_day: number;
  projected_full_date?: string;
  recorded_at: string;
}

export interface DriveUsageHistory {
  drive_id: string;
  days: number;
  history: DriveUsageData[];
  count: number;
}

export interface RequestAnalytic {
  id: number;
  tmdb_id: number;
  media_type: string;
  title: string;
  request_count: number;
  last_requested: string;
  first_requested: string;
  avg_processing_time_seconds: number;
  success_rate: number;
  popularity_score: number;
  created_at: string;
  updated_at: string;
}

export interface PopularityTrend {
  id: number;
  tmdb_id: number;
  media_type: string;
  title: string;
  trend_source: string;
  popularity_score: number;
  trend_direction: 'rising' | 'stable' | 'declining';
  forecast_confidence: number;
  metadata: string;
  valid_until: string;
  created_at: string;
}

export interface AnalyticsOverview {
  summary: {
    total_drives: number;
    drives_with_alerts: number;
    critical_drives: number;
    warning_drives: number;
    overall_usage_percent: number;
    total_storage_gb: number;
    used_storage_gb: number;
    available_storage_gb: number;
  };
  active_alerts: DriveAlert[];
  drive_usage: DriveUsageData[];
  storage_pools: StoragePool[];
  request_analytics: RequestAnalytic[];
  trending_content: PopularityTrend[];
}

export interface DriveAlertsResponse {
  alerts: DriveAlert[];
  count: number;
}