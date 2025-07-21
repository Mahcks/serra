import React from 'react';
import type { AnalyticsOverview } from '../../types/analytics';

interface OverviewTabProps {
  overview: AnalyticsOverview;
  watchAnalytics?: any;
  formatBytes: (bytes: number | undefined | null) => string;
  formatPercentage: (percentage: number | undefined | null) => string;
  handleAcknowledgeAlert: (alertId: number) => void;
  getAlertTypeColor: (alertType: string) => string;
  getAlertPriority: (alert: any) => string;
}

export const OverviewTab: React.FC<OverviewTabProps> = ({
  overview,
  watchAnalytics,
  formatBytes,
  formatPercentage,
  handleAcknowledgeAlert,
  getAlertTypeColor,
  getAlertPriority
}) => {
  return (
    <>
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-card p-6 rounded-lg shadow border">
          <h3 className="text-sm font-medium text-muted-foreground mb-2">Total Storage</h3>
          <p className="text-2xl font-bold text-foreground">
            {formatBytes((overview.summary.total_storage_gb ?? 0) * 1024 * 1024 * 1024)}
          </p>
          <p className="text-sm text-muted-foreground">
            {formatPercentage(overview.summary.overall_usage_percent)} used
          </p>
        </div>

        <div className="bg-card p-6 rounded-lg shadow border">
          <h3 className="text-sm font-medium text-muted-foreground mb-2">Drive Status</h3>
          <p className="text-2xl font-bold text-foreground">{overview.summary.total_drives}</p>
          <p className="text-sm text-muted-foreground">
            {overview.summary.critical_drives} critical, {overview.summary.warning_drives} warning
          </p>
        </div>

        <div className="bg-card p-6 rounded-lg shadow border">
          <h3 className="text-sm font-medium text-muted-foreground mb-2">Active Alerts</h3>
          <p className="text-2xl font-bold text-destructive">{overview.summary.drives_with_alerts}</p>
          <p className="text-sm text-muted-foreground">Require attention</p>
        </div>

        <div className="bg-card p-6 rounded-lg shadow border">
          <h3 className="text-sm font-medium text-muted-foreground mb-2">Available Space</h3>
          <p className="text-2xl font-bold text-green-600">
            {formatBytes((overview.summary.available_storage_gb ?? 0) * 1024 * 1024 * 1024)}
          </p>
          <p className="text-sm text-muted-foreground">Free space remaining</p>
        </div>
      </div>

      {/* Watch Analytics Summary */}
      {watchAnalytics?.jellystat_enabled && (
        <div className="mb-8">
          <h2 className="text-xl font-semibold text-foreground mb-4 flex items-center gap-2">
            👥 Watch Analytics Overview
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Active Users */}
            {watchAnalytics?.active_users && (
              <div className="bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 p-6 rounded-xl">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-sm font-medium text-muted-foreground">Active Users</h3>
                  <div className="w-10 h-10 rounded-full bg-blue-500/20 flex items-center justify-center">
                    👥
                  </div>
                </div>
                <p className="text-3xl font-bold text-blue-600 mb-2">
                  {watchAnalytics.active_users.length}
                </p>
                <p className="text-sm text-muted-foreground">Last 30 days</p>
                {watchAnalytics.active_users.length > 0 && (
                  <p className="text-xs text-muted-foreground mt-2">
                    Top: {watchAnalytics.active_users[0].user_name} ({watchAnalytics.active_users[0].plays} plays)
                  </p>
                )}
              </div>
            )}

            {/* Total Sessions */}
            {watchAnalytics?.playback_methods && (
              <div className="bg-gradient-to-br from-green-500/10 to-emerald-500/10 border border-green-500/20 p-6 rounded-xl">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-sm font-medium text-muted-foreground">Total Sessions</h3>
                  <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                    ▶️
                  </div>
                </div>
                <p className="text-3xl font-bold text-green-600 mb-2">
                  {watchAnalytics.playback_methods.reduce((sum: number, method: any) => sum + method.count, 0).toLocaleString()}
                </p>
                <p className="text-sm text-muted-foreground">All playback methods</p>
                <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
                  {watchAnalytics.playback_methods.map((method: any, index: number) => (
                    <span key={index} className="px-2 py-1 bg-muted rounded">
                      {method.name}: {method.count}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Most Popular Content */}
            {watchAnalytics?.most_viewed_content && watchAnalytics.most_viewed_content.length > 0 && (
              <div className="bg-gradient-to-br from-orange-500/10 to-red-500/10 border border-orange-500/20 p-6 rounded-xl">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-sm font-medium text-muted-foreground">Top Content</h3>
                  <div className="w-10 h-10 rounded-full bg-orange-500/20 flex items-center justify-center">
                    🔥
                  </div>
                </div>
                <p className="text-3xl font-bold text-orange-600 mb-2">
                  {watchAnalytics.most_viewed_content[0].total_plays}
                </p>
                <p className="text-sm text-muted-foreground">Most viewed plays</p>
                <p className="text-xs text-muted-foreground mt-2 truncate">
                  "{watchAnalytics.most_viewed_content[0].item_name}"
                </p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Active Alerts */}
      {overview.active_alerts && overview.active_alerts.length > 0 && (
        <div className="bg-card rounded-lg shadow border mb-8">
          <div className="px-6 py-4 border-b border-border">
            <h2 className="text-xl font-semibold text-foreground">Active Drive Alerts</h2>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {overview.active_alerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`p-4 rounded-lg border ${getAlertTypeColor(alert.alert_type)}`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-sm font-medium">
                          {getAlertPriority(alert)}
                        </span>
                        <span className="text-sm text-muted-foreground">
                          {alert.drive_name} ({alert.mount_path})
                        </span>
                      </div>
                      <p className="text-sm mb-2">{alert.alert_message}</p>
                      <div className="text-xs text-muted-foreground">
                        Triggered: {new Date(alert.last_triggered).toLocaleString()}
                        {alert.acknowledge_count > 0 && (
                          <span className="ml-2">
                            • Acknowledged {alert.acknowledge_count} time(s)
                          </span>
                        )}
                      </div>
                    </div>
                    <button
                      onClick={() => handleAcknowledgeAlert(alert.id)}
                      className="ml-4 bg-secondary text-secondary-foreground px-3 py-1 rounded text-sm hover:bg-secondary/80"
                    >
                      Acknowledge
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  );
};