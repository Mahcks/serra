import React, { useState, useEffect } from 'react';
import { analyticsApi } from '../../lib/analytics-api';
import { getErrorMessage } from '../../utils/errorHandling';
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent } from '../ui/chart';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, LineChart, Line } from 'recharts';

export const RequestsTab: React.FC = () => {
  const [requestAnalytics, setRequestAnalytics] = useState<any>(null);
  const [requestTrends, setRequestTrends] = useState<any>(null);
  const [failureAnalysis, setFailureAnalysis] = useState<any>(null);
  const [contentAvailability, setContentAvailability] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [days, setDays] = useState(30);

  useEffect(() => {
    loadRequestAnalytics();
  }, [days]);

  const loadRequestAnalytics = async () => {
    try {
      setLoading(true);
      const [analytics, trends, failures, availability] = await Promise.all([
        analyticsApi.getRequestAnalytics(days, 10),
        analyticsApi.getRequestTrends(days),
        analyticsApi.getFailureAnalysis(days),
        analyticsApi.getContentAvailability(days, 10)
      ]);
      
      setRequestAnalytics(analytics);
      setRequestTrends(trends);
      setFailureAnalysis(failures);
      setContentAvailability(availability);
      setError(null);
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      console.error('Failed to load request analytics:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatPercentage = (value: number | undefined | null) => {
    if (value === undefined || value === null || isNaN(value)) {
      return '0.0%';
    }
    return `${value.toFixed(1)}%`;
  };
  const formatTime = (seconds: number) => {
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds.toFixed(0)}s`;
  };

  // Chart color configurations
  const statusColors = {
    approved: '#3b82f6',     // blue
    completed: '#10b981',    // green  
    fulfilled: '#10b981',    // green
    failed: '#ef4444',       // red
    pending: '#f59e0b',      // amber
    denied: '#ef4444'        // red
  };

  const chartConfig = {
    approved: { label: 'Approved', color: statusColors.approved },
    completed: { label: 'Completed', color: statusColors.completed },
    fulfilled: { label: 'Fulfilled', color: statusColors.fulfilled },
    failed: { label: 'Failed', color: statusColors.failed },
    pending: { label: 'Pending', color: statusColors.pending },
    denied: { label: 'Denied', color: statusColors.denied },
    movie: { label: 'Movies', color: '#8b5cf6' },
    tv: { label: 'TV Shows', color: '#06b6d4' }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto p-6 space-y-8">
          <div className="text-center py-12">
            <div className="animate-spin w-8 h-8 border-4 border-primary border-t-transparent rounded-full mx-auto mb-4"></div>
            <h2 className="text-xl font-semibold text-foreground mb-2">Loading Request Analytics</h2>
            <p className="text-muted-foreground">Fetching performance data, trends, and insights...</p>
          </div>
          <div className="animate-pulse">
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="bg-card p-6 rounded-xl border border-border/40">
                  <div className="h-4 bg-muted rounded w-3/4 mb-3"></div>
                  <div className="h-8 bg-muted rounded w-1/2 mb-2"></div>
                  <div className="h-3 bg-muted rounded w-2/3"></div>
                </div>
              ))}
            </div>
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="bg-card p-6 rounded-xl border border-border/40">
                  <div className="h-6 bg-muted rounded w-1/2 mb-4"></div>
                  <div className="h-64 bg-muted rounded"></div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/15 border border-destructive/20 text-destructive px-4 py-3 rounded">
        {error}
        <button
          onClick={loadRequestAnalytics}
          className="ml-4 bg-destructive text-destructive-foreground px-3 py-1 rounded text-sm hover:bg-destructive/90"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto p-6 space-y-8">
        {/* Header Section */}
        <div className="bg-gradient-to-r from-background to-accent/10 rounded-xl p-6 border border-border/50">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-foreground mb-2">Request Analytics</h1>
              <p className="text-muted-foreground">Monitor request performance, trends, and user activity</p>
            </div>
            <div className="flex items-center gap-3">
              <select
                value={days}
                onChange={(e) => setDays(Number(e.target.value))}
                className="px-4 py-2 border border-input rounded-lg bg-background text-foreground text-sm shadow-sm hover:border-border transition-colors"
              >
                <option value={7}>Last 7 days</option>
                <option value={30}>Last 30 days</option>
                <option value={90}>Last 90 days</option>
                <option value={365}>Last year</option>
              </select>
            </div>
          </div>
        </div>

        {/* Key Metrics Cards */}
        {requestAnalytics?.success_rates && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
            {requestAnalytics.success_rates.map((rate: any, index: number) => (
              <div key={index} className="bg-card rounded-xl p-6 border border-border/40 hover:border-border hover:shadow-lg transition-all duration-200 group">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-3">
                      <div className={`w-3 h-3 rounded-full ${
                        rate.status === 'approved' ? 'bg-blue-500 shadow-blue-500/30' :
                        rate.status === 'fulfilled' || rate.status === 'completed' ? 'bg-green-500 shadow-green-500/30' :
                        rate.status === 'failed' ? 'bg-red-500 shadow-red-500/30' :
                        rate.status === 'pending' ? 'bg-yellow-500 shadow-yellow-500/30' :
                        'bg-gray-500 shadow-gray-500/30'
                      } shadow-lg`} />
                      <p className="text-sm font-medium text-muted-foreground capitalize">{rate.status}</p>
                    </div>
                    <p className="text-3xl font-bold text-foreground mb-1 group-hover:text-primary transition-colors">
                      {rate.total_requests}
                    </p>
                    <p className="text-sm text-muted-foreground">{formatPercentage(rate.percentage)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Processing Insights */}
        {requestAnalytics && requestTrends && (
          <div className="bg-gradient-to-r from-card to-accent/5 rounded-xl p-6 border border-border/40">
            <h3 className="text-xl font-semibold text-foreground mb-4">Processing Insights</h3>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
              <div className="text-center">
                <div className="text-2xl font-bold text-blue-600">
                  {requestAnalytics.success_rates?.reduce((sum: number, rate: any) => sum + rate.total_requests, 0) || 0}
                </div>
                <div className="text-sm text-muted-foreground">Total Requests</div>
                <div className="text-xs text-muted-foreground mt-1">Last {days} days</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600">
                  {(() => {
                    const successfulStatuses = ['approved', 'completed', 'fulfilled'];
                    const successful = requestAnalytics.success_rates?.filter((rate: any) => 
                      successfulStatuses.includes(rate.status)
                    ).reduce((sum: number, rate: any) => sum + rate.total_requests, 0) || 0;
                    
                    const total = requestAnalytics.success_rates?.reduce((sum: number, rate: any) => 
                      sum + rate.total_requests, 0) || 1;
                    
                    return Math.round((successful / total) * 100);
                  })()}%
                </div>
                <div className="text-sm text-muted-foreground">Success Rate</div>
                <div className="text-xs text-muted-foreground mt-1">Approved/Completed</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-purple-600">
                  {requestTrends?.daily_trends?.length || 0}
                </div>
                <div className="text-sm text-muted-foreground">Active Days</div>
                <div className="text-xs text-muted-foreground mt-1">With requests</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-orange-600">
                  {(() => {
                    const total = requestTrends?.daily_trends?.reduce((sum: number, day: any) => 
                      sum + day.total_requests, 0) || 0;
                    const days = requestTrends?.daily_trends?.length || 1;
                    return Math.round(total / days);
                  })()}
                </div>
                <div className="text-sm text-muted-foreground">Avg Daily</div>
                <div className="text-xs text-muted-foreground mt-1">Requests per day</div>
              </div>
            </div>
          </div>
        )}

        {/* Charts Section */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
          {/* Status Distribution */}
          {requestAnalytics?.success_rates && requestAnalytics.success_rates.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Status Distribution</h3>
                <p className="text-sm text-muted-foreground mt-1">Request breakdown by status</p>
              </div>
              <div className="p-6">
                <ChartContainer config={chartConfig} className="h-80">
                  <PieChart>
                    <Pie
                      data={requestAnalytics.success_rates.map((item: any) => ({
                        name: item.status,
                        value: item.total_requests,
                        fill: statusColors[item.status as keyof typeof statusColors] || '#6b7280'
                      }))}
                      cx="50%"
                      cy="50%"
                      outerRadius={120}
                      innerRadius={60}
                      dataKey="value"
                      strokeWidth={2}
                      stroke="hsl(var(--background))"
                      label={({ name, value }) => `${name}: ${value}`}
                    />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <ChartLegend content={<ChartLegendContent />} />
                  </PieChart>
                </ChartContainer>
              </div>
            </div>
          )}

          {/* Performance by Type */}
          {requestAnalytics?.performance && requestAnalytics.performance.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Performance by Type & Status</h3>
                <p className="text-sm text-muted-foreground mt-1">Request volume by media type and status</p>
              </div>
              <div className="p-6">
                <ChartContainer 
                  config={chartConfig} 
                  className="h-80 [&_.recharts-tooltip-cursor]:!fill-transparent [&_.recharts-cartesian-grid]:opacity-30 [&_.recharts-rectangle.recharts-tooltip-cursor]:!fill-transparent [&_.recharts-rectangle.recharts-tooltip-cursor]:!opacity-0"
                >
                  <BarChart data={(() => {
                    // Transform the performance data into a format suitable for stacked bars
                    const groupedData: Record<string, any> = {};
                    
                    requestAnalytics.performance.forEach((item: any) => {
                      if (!groupedData[item.media_type]) {
                        groupedData[item.media_type] = { name: item.media_type };
                      }
                      groupedData[item.media_type][item.status] = item.count;
                    });
                    
                    return Object.values(groupedData);
                  })()}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.3} />
                    <XAxis 
                      dataKey="name" 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Media Type', position: 'insideBottom', offset: -10, style: { textAnchor: 'middle' } }}
                    />
                    <YAxis 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Count', angle: -90, position: 'insideLeft', style: { textAnchor: 'middle' } }}
                    />
                    {/* Create a Bar for each possible status */}
                    <Bar 
                      dataKey="approved" 
                      stackId="a" 
                      fill={statusColors.approved} 
                      name="Approved" 
                      radius={[0, 0, 0, 0]}
                      stroke="none"
                    />
                    <Bar 
                      dataKey="fulfilled" 
                      stackId="a" 
                      fill={statusColors.fulfilled} 
                      name="Fulfilled" 
                      radius={[0, 0, 0, 0]}
                      stroke="none"
                    />
                    <Bar 
                      dataKey="completed" 
                      stackId="a" 
                      fill={statusColors.completed} 
                      name="Completed" 
                      radius={[0, 0, 0, 0]}
                      stroke="none"
                    />
                    <Bar 
                      dataKey="failed" 
                      stackId="a" 
                      fill={statusColors.failed} 
                      name="Failed" 
                      radius={[0, 0, 0, 0]}
                      stroke="none"
                    />
                    <Bar 
                      dataKey="pending" 
                      stackId="a" 
                      fill={statusColors.pending} 
                      name="Pending" 
                      radius={[6, 6, 0, 0]}
                      stroke="none"
                    />
                    <Bar 
                      dataKey="denied" 
                      stackId="a" 
                      fill={statusColors.denied} 
                      name="Denied" 
                      radius={[0, 0, 0, 0]}
                      stroke="none"
                    />
                    <ChartTooltip 
                      content={<ChartTooltipContent />} 
                      cursor={false}
                    />
                    <ChartLegend content={<ChartLegendContent />} />
                  </BarChart>
                </ChartContainer>
              </div>
            </div>
          )}
        </div>

        {/* Trends Section */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
          {/* Daily Trends */}
          {requestTrends?.daily_trends && requestTrends.daily_trends.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Daily Trends</h3>
                <p className="text-sm text-muted-foreground mt-1">Request volume over time</p>
              </div>
              <div className="p-6">
                <ChartContainer config={chartConfig} className="h-80">
                  <LineChart data={requestTrends.daily_trends}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.3} />
                    <XAxis 
                      dataKey="request_date" 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Date', position: 'insideBottom', offset: -10, style: { textAnchor: 'middle' } }}
                    />
                    <YAxis 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Requests', angle: -90, position: 'insideLeft', style: { textAnchor: 'middle' } }}
                    />
                    <Line 
                      type="monotone" 
                      dataKey="total_requests" 
                      stroke="#3b82f6" 
                      strokeWidth={3}
                      dot={{ fill: "#3b82f6", strokeWidth: 2, r: 4 }}
                      activeDot={{ r: 6, stroke: "#3b82f6", strokeWidth: 2 }}
                      name="Total Requests"
                    />
                    <Line 
                      type="monotone" 
                      dataKey="completed_requests" 
                      stroke="#10b981" 
                      strokeWidth={3}
                      dot={{ fill: "#10b981", strokeWidth: 2, r: 4 }}
                      activeDot={{ r: 6, stroke: "#10b981", strokeWidth: 2 }}
                      name="Completed Requests"
                    />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <ChartLegend content={<ChartLegendContent />} />
                  </LineChart>
                </ChartContainer>
              </div>
            </div>
          )}

          {/* Hourly Patterns */}
          {requestTrends?.hourly_volume && requestTrends.hourly_volume.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Hourly Patterns</h3>
                <p className="text-sm text-muted-foreground mt-1">Request activity by hour</p>
              </div>
              <div className="p-6">
                <ChartContainer config={chartConfig} className="h-80">
                  <BarChart data={requestTrends.hourly_volume}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.3} />
                    <XAxis 
                      dataKey="hour_of_day" 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Hour of Day', position: 'insideBottom', offset: -10, style: { textAnchor: 'middle' } }}
                    />
                    <YAxis 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Requests', angle: -90, position: 'insideLeft', style: { textAnchor: 'middle' } }}
                    />
                    <Bar 
                      dataKey="total_requests" 
                      fill="#3b82f6" 
                      radius={[6, 6, 0, 0]}
                      name="Total Requests"
                    />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <ChartLegend content={<ChartLegendContent />} />
                  </BarChart>
                </ChartContainer>
              </div>
            </div>
          )}
        </div>

        {/* Additional Analytics Section */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
          {/* Failure Analysis */}
          {failureAnalysis && failureAnalysis.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Failure Analysis</h3>
                <p className="text-sm text-muted-foreground mt-1">Breakdown of failed request causes</p>
              </div>
              <div className="p-6">
                <ChartContainer config={chartConfig} className="h-80">
                  <BarChart data={failureAnalysis.map((item: any) => ({
                    name: item.media_type,
                    not_found: item.not_found_failures,
                    connection: item.connection_failures,
                    quality: item.quality_failures,
                    storage: item.storage_failures,
                    total: item.total_failures
                  }))}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.3} />
                    <XAxis 
                      dataKey="name" 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Media Type', position: 'insideBottom', offset: -10, style: { textAnchor: 'middle' } }}
                    />
                    <YAxis 
                      stroke="hsl(var(--muted-foreground))" 
                      fontSize={12}
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                      label={{ value: 'Failures', angle: -90, position: 'insideLeft', style: { textAnchor: 'middle' } }}
                    />
                    <Bar dataKey="not_found" stackId="failure" fill="#ef4444" name="Not Found" />
                    <Bar dataKey="connection" stackId="failure" fill="#f97316" name="Connection" />
                    <Bar dataKey="quality" stackId="failure" fill="#eab308" name="Quality" />
                    <Bar dataKey="storage" stackId="failure" fill="#6b7280" name="Storage" />
                    <ChartTooltip content={<ChartTooltipContent />} cursor={false} />
                    <ChartLegend content={<ChartLegendContent />} />
                  </BarChart>
                </ChartContainer>
              </div>
            </div>
          )}

          {/* Content Availability vs Requests */}
          {contentAvailability && contentAvailability.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Content Availability Gap</h3>
                <p className="text-sm text-muted-foreground mt-1">Popular requested content not yet available</p>
              </div>
              <div className="p-6">
                <div className="space-y-4 max-h-80 overflow-y-auto">
                  {contentAvailability
                    .filter((item: any) => !item.is_available)
                    .slice(0, 10)
                    .map((item: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-4 rounded-lg bg-destructive/5 border border-destructive/20 hover:bg-destructive/10 transition-colors">
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-foreground truncate text-base">
                          {item.title?.String || item.title || 'Unknown Title'}
                        </p>
                        <div className="flex items-center gap-3 text-sm text-muted-foreground mt-2">
                          <span className="capitalize bg-primary/10 text-primary px-2 py-1 rounded-md font-medium">
                            {item.media_type}
                          </span>
                          <span>{item.request_count} requests</span>
                          <span className="text-orange-600 font-medium">Not Available</span>
                        </div>
                      </div>
                      <div className="text-right ml-4">
                        <div className="text-lg font-bold text-destructive">
                          {item.fulfilled_requests}/{item.request_count}
                        </div>
                        <div className="text-xs text-muted-foreground">fulfilled</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Data Tables Section */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
          {/* Popular Requested Content */}
          {requestAnalytics?.popular_content && requestAnalytics.popular_content.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">Most Requested Content</h3>
                <p className="text-sm text-muted-foreground mt-1">Popular content and fulfillment rates</p>
              </div>
              <div className="p-6">
                <div className="space-y-4 max-h-96 overflow-y-auto">
                  {requestAnalytics.popular_content.map((item: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-4 rounded-lg bg-accent/20 hover:bg-accent/30 transition-colors border border-border/20">
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-foreground truncate text-base">
                          {item.title?.String || item.title || 'Unknown Title'}
                        </p>
                        <div className="flex items-center gap-3 text-sm text-muted-foreground mt-2">
                          <span className="capitalize bg-primary/10 text-primary px-2 py-1 rounded-md font-medium">
                            {item.media_type}
                          </span>
                          <span>{item.request_count} requests</span>
                          <span className={`font-semibold ${
                            item.fulfillment_rate >= 80 ? 'text-green-600' :
                            item.fulfillment_rate >= 50 ? 'text-yellow-600' :
                            'text-red-600'
                          }`}>
                            {formatPercentage(item.fulfillment_rate)}
                          </span>
                        </div>
                      </div>
                      <div className="text-right ml-4">
                        <div className="text-lg font-bold text-foreground">
                          {item.fulfilled_count}/{item.request_count}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* User Fulfillment Stats */}
          {requestAnalytics?.user_fulfillment && requestAnalytics.user_fulfillment.length > 0 && (
            <div className="bg-card rounded-xl border border-border/40 overflow-hidden">
              <div className="p-6 border-b border-border/20">
                <h3 className="text-xl font-semibold text-foreground">User Request Stats</h3>
                <p className="text-sm text-muted-foreground mt-1">Request activity and success rates by user</p>
              </div>
              <div className="p-6">
                <div className="space-y-4 max-h-96 overflow-y-auto">
                  {requestAnalytics.user_fulfillment.map((user: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-4 rounded-lg bg-accent/20 hover:bg-accent/30 transition-colors border border-border/20">
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-foreground truncate text-base">{user.username}</p>
                        <div className="flex items-center gap-3 text-sm text-muted-foreground mt-2">
                          <span>{user.total_requests} total</span>
                          <span className="text-green-600 font-medium">{user.completed_requests} completed</span>
                          <span className="text-red-600 font-medium">{user.failed_requests} failed</span>
                        </div>
                      </div>
                      <div className="text-right ml-4">
                        <div className={`text-lg font-bold ${
                          user.success_rate >= 80 ? 'text-green-600' :
                          user.success_rate >= 50 ? 'text-yellow-600' :
                          'text-red-600'
                        }`}>
                          {formatPercentage(user.success_rate)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};