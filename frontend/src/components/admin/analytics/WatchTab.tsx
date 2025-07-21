import React, { useState, useEffect } from 'react';
import { analyticsApi } from '../../../lib/analytics-api';
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent } from '../../ui/chart';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';
import { getErrorMessage } from '../../../utils/errorHandling';

export const WatchTab: React.FC = () => {
  const [watchAnalytics, setWatchAnalytics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadWatchAnalytics();
  }, []);

  const loadWatchAnalytics = async () => {
    try {
      setLoading(true);
      const analytics = await analyticsApi.getWatchAnalytics(15);
      setWatchAnalytics(analytics);
      setError(null);
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      console.error('Failed to load watch analytics:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatWatchTime = (minutes: number) => {
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    if (hours < 24) return `${hours}h ${remainingMinutes}m`;
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}d ${remainingHours}h`;
  };

  const formatPlayDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    return formatWatchTime(minutes);
  };

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="bg-card p-6 rounded-lg shadow border">
              <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
              <div className="h-8 bg-muted rounded w-1/2"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/15 border border-destructive/20 text-destructive px-4 py-3 rounded">
        {error}
        <button
          onClick={loadWatchAnalytics}
          className="ml-4 bg-destructive text-destructive-foreground px-3 py-1 rounded text-sm hover:bg-destructive/90"
        >
          Retry
        </button>
      </div>
    );
  }

  if (!watchAnalytics?.jellystat_enabled) {
    return (
      <div className="text-center py-12">
        <div className="mb-4">
          <svg className="w-16 h-16 mx-auto text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium mb-2">Jellystat Not Connected</h3>
        <p className="text-muted-foreground mb-4">
          Connect Jellystat to view detailed watch analytics and user behavior insights.
        </p>
        <div className="text-sm text-muted-foreground bg-muted p-4 rounded-lg max-w-md mx-auto">
          <p className="font-medium mb-2">What you'll get with Jellystat:</p>
          <ul className="text-left space-y-1">
            <li>• User watch time and activity</li>
            <li>• Popular content analytics</li>
            <li>• Library overview statistics</li>
            <li>• Recently watched content</li>
            <li>• Individual user watch history</li>
          </ul>
        </div>
      </div>
    );
  }

  // Prepare chart data
  const playbackMethodsChartData = watchAnalytics?.playback_methods?.map((method: any) => ({
    name: method.name,
    value: method.count,
    fill: method.name === 'DirectPlay' ? '#22c55e' : 
          method.name === 'DirectStream' ? '#3b82f6' : 
          method.name === 'Transcode' ? '#f59e0b' : '#6b7280'
  })) || [];

  const activeUsersChartData = watchAnalytics?.active_users?.slice(0, 10).map((user: any) => ({
    name: user.user_name,
    plays: user.plays
  })) || [];

  const topContentData = watchAnalytics?.most_viewed_content?.slice(0, 8).map((item: any) => ({
    name: item.item_name.length > 20 ? item.item_name.substring(0, 20) + '...' : item.item_name,
    plays: item.total_plays,
    type: item.item_type
  })) || [];

  return (
    <>
      {/* Header Stats */}
      {watchAnalytics?.libraries && watchAnalytics.libraries.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {watchAnalytics.libraries.map((library: any, index: number) => (
            <div key={index} className="bg-gradient-to-br from-card to-card/50 p-6 rounded-xl shadow-lg border border-border/50 hover:shadow-xl transition-all duration-300">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-muted-foreground">{library.name}</h3>
                <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                  {library.collection_type === 'movies' ? '🎬' : library.collection_type === 'tvshows' ? '📺' : '📚'}
                </div>
              </div>
              <p className="text-3xl font-bold text-foreground mb-2">{library.library_count.toLocaleString()}</p>
              <div className="text-sm text-muted-foreground">
                <span className="capitalize">{library.collection_type || 'Mixed'} items</span>
                {library.season_count > 0 && (
                  <div className="mt-1 text-xs">
                    {library.season_count} seasons • {library.episode_count} episodes
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Charts Section */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-8 mb-8">
        {/* Playback Methods Chart */}
        {playbackMethodsChartData.length > 0 && (
          <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
                🎭 Playback Methods
              </h2>
              <div className="text-sm text-muted-foreground">Last 30 days</div>
            </div>
            <ChartContainer
              config={{
                DirectPlay: { label: "Direct Play", color: "#22c55e" },
                DirectStream: { label: "Direct Stream", color: "#3b82f6" },
                Transcode: { label: "Transcode", color: "#f59e0b" }
              }}
              className="h-[300px] w-full"
            >
              <PieChart>
                <Pie
                  data={playbackMethodsChartData}
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  innerRadius={40}
                  dataKey="value"
                  strokeWidth={2}
                >
                  {playbackMethodsChartData.map((entry: any, index: number) => (
                    <Cell key={`cell-${index}`} fill={entry.fill} />
                  ))}
                </Pie>
                <ChartTooltip content={<ChartTooltipContent />} />
                <ChartLegend content={<ChartLegendContent />} />
              </PieChart>
            </ChartContainer>
          </div>
        )}

        {/* Top Active Users Chart */}
        {activeUsersChartData.length > 0 && (
          <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
                👑 Most Active Users
              </h2>
              <div className="text-sm text-muted-foreground">Last 30 days</div>
            </div>
            <ChartContainer
              config={{
                plays: { label: "Plays", color: "#8b5cf6" }
              }}
              className="h-[300px] w-full"
            >
              <BarChart data={activeUsersChartData} margin={{ top: 20, right: 30, left: 20, bottom: 60 }}>
                <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                <XAxis 
                  dataKey="name" 
                  angle={-45}
                  textAnchor="end"
                  height={80}
                  className="text-xs"
                />
                <YAxis className="text-xs" />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="plays" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ChartContainer>
          </div>
        )}
      </div>

      {/* Top Content Chart */}
      {topContentData.length > 0 && (
        <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50 p-6 mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
              🔥 Most Viewed Content
            </h2>
            <div className="text-sm text-muted-foreground">Last 30 days</div>
          </div>
          <ChartContainer
            config={{
              plays: { label: "Views", color: "#ef4444" }
            }}
            className="h-[400px] w-full"
          >
            <BarChart data={topContentData} margin={{ top: 20, right: 30, left: 20, bottom: 100 }}>
              <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
              <XAxis 
                dataKey="name" 
                angle={-45}
                textAnchor="end"
                height={120}
                className="text-xs"
              />
              <YAxis className="text-xs" />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Bar dataKey="plays" fill="#ef4444" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </div>
      )}

      {/* Data Cards Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-2 gap-8 mb-8">
        {/* Popular Content List */}
        {watchAnalytics?.popular_content && watchAnalytics.popular_content.length > 0 && (
          <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50">
            <div className="px-6 py-4 border-b border-border/50">
              <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
                ⭐ Most Popular Content
              </h2>
              <p className="text-sm text-muted-foreground mt-1">Trending by unique viewers</p>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {watchAnalytics.popular_content.slice(0, 6).map((item: any, index: number) => (
                  <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors">
                    <div className="flex items-center gap-3 flex-1">
                      <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-sm font-bold text-primary">
                        #{index + 1}
                      </div>
                      <div className="flex-1">
                        <p className="font-medium text-foreground">{item.item_name}</p>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <span className="px-2 py-1 rounded-full bg-secondary text-xs">
                            {item.item_type === 'Movie' ? '🎬' : '📺'} {item.item_type}
                          </span>
                          <span>•</span>
                          <span>{item.library_name}</span>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-lg font-bold text-foreground">
                        {item.total_plays}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        plays
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Quick Stats */}
        {(watchAnalytics?.active_users || watchAnalytics?.playback_methods) && (
          <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50">
            <div className="px-6 py-4 border-b border-border/50">
              <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
                📊 Quick Stats
              </h2>
              <p className="text-sm text-muted-foreground mt-1">Key metrics at a glance</p>
            </div>
            <div className="p-6 space-y-6">
              {/* Total Active Users */}
              {watchAnalytics?.active_users && (
                <div className="flex items-center justify-between p-4 rounded-lg bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-blue-500/20 flex items-center justify-center">
                      👥
                    </div>
                    <div>
                      <p className="font-medium text-foreground">Active Users</p>
                      <p className="text-sm text-muted-foreground">Last 30 days</p>
                    </div>
                  </div>
                  <div className="text-2xl font-bold text-blue-600">
                    {watchAnalytics.active_users.length}
                  </div>
                </div>
              )}

              {/* Total Playback Sessions */}
              {watchAnalytics?.playback_methods && (
                <div className="flex items-center justify-between p-4 rounded-lg bg-gradient-to-r from-green-500/10 to-emerald-500/10 border border-green-500/20">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                      ▶️
                    </div>
                    <div>
                      <p className="font-medium text-foreground">Total Sessions</p>
                      <p className="text-sm text-muted-foreground">All playback methods</p>
                    </div>
                  </div>
                  <div className="text-2xl font-bold text-green-600">
                    {watchAnalytics.playback_methods.reduce((sum: number, method: any) => sum + method.count, 0).toLocaleString()}
                  </div>
                </div>
              )}

              {/* Most Active User */}
              {watchAnalytics?.active_users && watchAnalytics.active_users.length > 0 && (
                <div className="flex items-center justify-between p-4 rounded-lg bg-gradient-to-r from-orange-500/10 to-red-500/10 border border-orange-500/20">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-orange-500/20 flex items-center justify-center">
                      🏆
                    </div>
                    <div>
                      <p className="font-medium text-foreground">Top User</p>
                      <p className="text-sm text-muted-foreground">{watchAnalytics.active_users[0].user_name}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-2xl font-bold text-orange-600">
                      {watchAnalytics.active_users[0].plays}
                    </div>
                    <div className="text-xs text-muted-foreground">plays</div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

      </div>


      {/* Recently Watched */}
      {watchAnalytics?.recently_watched && watchAnalytics.recently_watched.length > 0 && (
        <div className="bg-gradient-to-br from-card to-card/50 rounded-xl shadow-lg border border-border/50">
          <div className="px-6 py-4 border-b border-border/50">
            <h2 className="text-xl font-semibold text-foreground flex items-center gap-2">
              🕒 Recently Watched
            </h2>
            <p className="text-sm text-muted-foreground mt-1">Latest viewing activity</p>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {watchAnalytics.recently_watched.map((item: any, index: number) => (
                <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors">
                  <div className="flex items-center gap-3 flex-1">
                    <div className="w-10 h-10 rounded-full bg-secondary flex items-center justify-center">
                      {item.item_type === 'Movie' ? '🎬' : '📺'}
                    </div>
                    <div className="flex-1">
                      <p className="font-medium text-foreground">{item.item_name}</p>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span className="capitalize">{item.item_type}</span>
                        <span>•</span>
                        <span>by {item.user_name}</span>
                        <span>•</span>
                        <span>{item.library_name}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-foreground">
                      {formatPlayDuration(item.play_duration)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(item.watched_at).toLocaleDateString()}
                    </div>
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