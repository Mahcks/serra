import React from 'react';
import { TrendingUp, AlertTriangle, HardDrive, Activity, Users, Database, Clock, CheckCircle2 } from 'lucide-react';
import { Progress } from '../../ui/progress';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import type { AnalyticsOverview } from '../../../types/analytics';

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
  getAlertPriority
}) => {
  const usagePercentage = overview.summary.overall_usage_percent ?? 0;
  const criticalThreshold = 90;
  const warningThreshold = 80;
  
  const getUsageColor = (percentage: number) => {
    if (percentage >= criticalThreshold) return 'text-red-500';
    if (percentage >= warningThreshold) return 'text-yellow-500';
    return 'text-green-500';
  };
  

  return (
    <div className="space-y-6">
      {/* Hero Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Storage Card */}
        <Card className="border-l-4 border-l-blue-500 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Storage</CardTitle>
            <Database className="h-5 w-5 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {formatBytes((overview.summary.total_storage_gb ?? 0) * 1024 * 1024 * 1024)}
            </div>
            <div className="flex items-center space-x-2 mt-2">
              <Progress value={usagePercentage} className="flex-1 h-2" />
              <span className={`text-sm font-medium ${getUsageColor(usagePercentage)}`}>
                {formatPercentage(usagePercentage)}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {formatBytes((overview.summary.available_storage_gb ?? 0) * 1024 * 1024 * 1024)} available
            </p>
          </CardContent>
        </Card>

        {/* Drive Health Card */}
        <Card className="border-l-4 border-l-green-500 bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-950/20 dark:to-emerald-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Drive Health</CardTitle>
            <HardDrive className="h-5 w-5 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {overview.summary.total_drives - overview.summary.critical_drives - overview.summary.warning_drives}
              <span className="text-sm text-muted-foreground">/{overview.summary.total_drives}</span>
            </div>
            <div className="flex items-center gap-2 mt-2">
              <Badge variant="outline" className="text-xs text-green-600 border-green-200">
                <CheckCircle2 className="w-3 h-3 mr-1" />
                Healthy
              </Badge>
              {overview.summary.warning_drives > 0 && (
                <Badge variant="outline" className="text-xs text-yellow-600 border-yellow-200">
                  {overview.summary.warning_drives} Warning
                </Badge>
              )}
              {overview.summary.critical_drives > 0 && (
                <Badge variant="destructive" className="text-xs">
                  {overview.summary.critical_drives} Critical
                </Badge>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Alerts Card */}
        <Card className={`border-l-4 ${
          overview.summary.drives_with_alerts > 0 
            ? 'border-l-red-500 bg-gradient-to-br from-red-50 to-pink-50 dark:from-red-950/20 dark:to-pink-950/20' 
            : 'border-l-gray-300 bg-gradient-to-br from-gray-50 to-slate-50 dark:from-gray-950/20 dark:to-slate-950/20'
        }`}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">System Alerts</CardTitle>
            {overview.summary.drives_with_alerts > 0 ? (
              <AlertTriangle className="h-5 w-5 text-red-500" />
            ) : (
              <CheckCircle2 className="h-5 w-5 text-green-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${
              overview.summary.drives_with_alerts > 0 ? 'text-red-600' : 'text-green-600'
            }`}>
              {overview.summary.drives_with_alerts}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {overview.summary.drives_with_alerts === 0 
                ? 'All systems operational' 
                : `${overview.summary.drives_with_alerts} alert${overview.summary.drives_with_alerts > 1 ? 's' : ''} need attention`
              }
            </p>
          </CardContent>
        </Card>

        {/* System Status Card */}
        <Card className="border-l-4 border-l-purple-500 bg-gradient-to-br from-purple-50 to-violet-50 dark:from-purple-950/20 dark:to-violet-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">System Status</CardTitle>
            <Activity className="h-5 w-5 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 rounded-full bg-green-500 animate-pulse" />
              <span className="text-lg font-semibold text-green-600">Operational</span>
            </div>
            <div className="flex items-center gap-2 mt-2">
              <Clock className="w-3 h-3 text-muted-foreground" />
              <span className="text-xs text-muted-foreground">
                Last updated: {new Date().toLocaleTimeString()}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Watch Analytics Summary */}
      {watchAnalytics?.jellystat_enabled && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5 text-blue-500" />
              Media Server Analytics
            </CardTitle>
            <CardDescription>User engagement and content popularity metrics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Active Users */}
              {watchAnalytics?.active_users && (
                <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-blue-500/10 via-blue-500/5 to-transparent border border-blue-200/50 dark:border-blue-800/50 p-6">
                  <div className="absolute top-4 right-4">
                    <div className="w-12 h-12 rounded-full bg-blue-500/20 flex items-center justify-center">
                      <Users className="h-6 w-6 text-blue-600" />
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-muted-foreground mb-1">Active Users</h3>
                    <p className="text-3xl font-bold text-blue-600 mb-1">
                      {watchAnalytics.active_users.length}
                    </p>
                    <p className="text-sm text-blue-600/70 mb-3">Last 30 days</p>
                    {watchAnalytics.active_users.length > 0 && (
                      <div className="bg-blue-500/10 rounded-lg p-2">
                        <p className="text-xs text-blue-700 dark:text-blue-300 font-medium">
                          Most Active: {watchAnalytics.active_users[0].user_name}
                        </p>
                        <p className="text-xs text-blue-600/70">
                          {watchAnalytics.active_users[0].plays} plays • {Math.round(watchAnalytics.active_users[0].total_watch_time / 60)} hours
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Total Sessions */}
              {watchAnalytics?.playback_methods && (
                <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-green-500/10 via-green-500/5 to-transparent border border-green-200/50 dark:border-green-800/50 p-6">
                  <div className="absolute top-4 right-4">
                    <div className="w-12 h-12 rounded-full bg-green-500/20 flex items-center justify-center">
                      <Activity className="h-6 w-6 text-green-600" />
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-muted-foreground mb-1">Total Sessions</h3>
                    <p className="text-3xl font-bold text-green-600 mb-1">
                      {watchAnalytics.playback_methods.reduce((sum: number, method: any) => sum + method.count, 0).toLocaleString()}
                    </p>
                    <p className="text-sm text-green-600/70 mb-3">All time</p>
                    <div className="space-y-1">
                      {watchAnalytics.playback_methods.slice(0, 2).map((method: { name: string; count: number }, index: number) => (
                        <div key={index} className="flex items-center justify-between bg-green-500/10 rounded px-2 py-1">
                          <span className="text-xs font-medium text-green-700 dark:text-green-300">{method.name}</span>
                          <span className="text-xs text-green-600/70">{method.count}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Most Popular Content */}
              {watchAnalytics?.most_viewed_content && watchAnalytics.most_viewed_content.length > 0 && (
                <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-orange-500/10 via-orange-500/5 to-transparent border border-orange-200/50 dark:border-orange-800/50 p-6">
                  <div className="absolute top-4 right-4">
                    <div className="w-12 h-12 rounded-full bg-orange-500/20 flex items-center justify-center">
                      <TrendingUp className="h-6 w-6 text-orange-600" />
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-muted-foreground mb-1">Top Content</h3>
                    <p className="text-3xl font-bold text-orange-600 mb-1">
                      {watchAnalytics.most_viewed_content[0].total_plays}
                    </p>
                    <p className="text-sm text-orange-600/70 mb-3">Most plays</p>
                    <div className="bg-orange-500/10 rounded-lg p-2">
                      <p className="text-xs font-medium text-orange-700 dark:text-orange-300 truncate">
                        "{watchAnalytics.most_viewed_content[0].item_name}"
                      </p>
                      <p className="text-xs text-orange-600/70">
                        {watchAnalytics.most_viewed_content[0].library_name}
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Active Alerts */}
      {overview.active_alerts && overview.active_alerts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              System Alerts
              <Badge variant="destructive" className="ml-2">
                {overview.active_alerts.length}
              </Badge>
            </CardTitle>
            <CardDescription>Critical system notifications requiring attention</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {overview.active_alerts.map((alert) => {
                const priority = getAlertPriority(alert);
                const isHighPriority = priority === 'CRITICAL';
                return (
                  <div
                    key={alert.id}
                    className={`relative rounded-lg border p-4 transition-all hover:shadow-md ${
                      isHighPriority 
                        ? 'bg-red-50 border-red-200 dark:bg-red-950/20 dark:border-red-800/50' 
                        : 'bg-yellow-50 border-yellow-200 dark:bg-yellow-950/20 dark:border-yellow-800/50'
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1 space-y-2">
                        <div className="flex items-center gap-2">
                          <Badge 
                            variant={isHighPriority ? 'destructive' : 'secondary'}
                            className="text-xs"
                          >
                            <AlertTriangle className="w-3 h-3 mr-1" />
                            {priority}
                          </Badge>
                          <Badge variant="outline" className="text-xs">
                            <HardDrive className="w-3 h-3 mr-1" />
                            {alert.drive_name}
                          </Badge>
                        </div>
                        <p className="text-sm font-medium">{alert.alert_message}</p>
                        <div className="flex items-center gap-4 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {new Date(alert.last_triggered).toLocaleString()}
                          </span>
                          {alert.acknowledge_count > 0 && (
                            <span className="flex items-center gap-1">
                              <CheckCircle2 className="w-3 h-3" />
                              Acked {alert.acknowledge_count}x
                            </span>
                          )}
                          <span className="text-xs text-muted-foreground">
                            {alert.mount_path}
                          </span>
                        </div>
                      </div>
                      <button
                        onClick={() => handleAcknowledgeAlert(alert.id)}
                        className="ml-4 bg-white hover:bg-gray-50 dark:bg-gray-800 dark:hover:bg-gray-700 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 px-3 py-1.5 rounded-md text-sm font-medium transition-colors flex items-center gap-1"
                      >
                        <CheckCircle2 className="w-3 h-3" />
                        Acknowledge
                      </button>
                    </div>
                    {isHighPriority && (
                      <div className="absolute -top-px -right-px w-3 h-3 bg-red-500 rounded-full animate-pulse" />
                    )}
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};