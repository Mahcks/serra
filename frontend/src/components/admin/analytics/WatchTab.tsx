import React, { useState, useEffect } from 'react';
import { analyticsApi } from '../../../lib/analytics-api';
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent } from '../../ui/chart';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, PieChart, Pie, Cell, LineChart, Line, Area, AreaChart } from 'recharts';
import { getErrorMessage } from '../../../utils/errorHandling';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Badge } from '../../ui/badge';
import { Progress } from '../../ui/progress';
import { Button } from '../../ui/button';
import { Users, Play, Clock, TrendingUp, Eye, Award, Activity, BarChart3, Settings, RefreshCw } from 'lucide-react';

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
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="animate-pulse">
              <Card>
                <CardContent className="p-6">
                  <div className="h-4 bg-muted rounded w-3/4 mb-4"></div>
                  <div className="h-8 bg-muted rounded w-1/2 mb-2"></div>
                  <div className="h-3 bg-muted rounded w-2/3"></div>
                </CardContent>
              </Card>
            </div>
          ))}
        </div>
        {[...Array(3)].map((_, i) => (
          <div key={i} className="animate-pulse">
            <Card>
              <CardContent className="p-6">
                <div className="h-6 bg-muted rounded w-1/4 mb-4"></div>
                <div className="h-64 bg-muted rounded"></div>
              </CardContent>
            </Card>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="bg-destructive/15 border border-destructive/20 text-destructive px-4 py-3 rounded">
            {error}
            <button
              onClick={loadWatchAnalytics}
              className="ml-4 bg-destructive text-destructive-foreground px-3 py-1 rounded text-sm hover:bg-destructive/90"
            >
              Retry
            </button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!watchAnalytics?.jellystat_enabled) {
    return (
      <Card className="border-dashed border-2">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-gradient-to-br from-purple-100 to-blue-100 dark:from-purple-900/20 dark:to-blue-900/20 flex items-center justify-center">
            <BarChart3 className="w-10 h-10 text-purple-600" />
          </div>
          <h3 className="text-2xl font-semibold mb-3">Watch Analytics Unavailable</h3>
          <p className="text-muted-foreground text-center max-w-md mb-6">
            Connect Jellystat to unlock powerful viewing insights, user engagement metrics, and content performance analytics.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-4xl mx-auto mb-8">
            {[
              { icon: Users, title: 'User Analytics', desc: 'Track viewing habits and user engagement' },
              { icon: TrendingUp, title: 'Content Trends', desc: 'Identify popular shows and movies' },
              { icon: Activity, title: 'Real-time Stats', desc: 'Live viewing activity and session data' }
            ].map((feature, index) => (
              <div key={index} className="bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-800/50 dark:to-gray-900/50 p-4 rounded-xl border">
                <feature.icon className="w-8 h-8 text-purple-600 mb-3" />
                <h4 className="font-semibold text-sm mb-1">{feature.title}</h4>
                <p className="text-xs text-muted-foreground">{feature.desc}</p>
              </div>
            ))}
          </div>
          <Button className="flex items-center gap-2">
            <Settings className="w-4 h-4" />
            Configure Jellystat
          </Button>
        </CardContent>
      </Card>
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

  const totalActiveUsers = watchAnalytics?.active_users?.length || 0;
  const totalSessions = watchAnalytics?.playback_methods?.reduce((sum: number, method: any) => sum + method.count, 0) || 0;
  const totalWatchTime = watchAnalytics?.active_users?.reduce((sum: number, user: any) => sum + (user.total_watch_time || 0), 0) || 0;
  const avgSessionLength = totalSessions > 0 ? Math.round(totalWatchTime / totalSessions) : 0;

  return (
    <div className="space-y-6">
      {/* Hero Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-l-4 border-l-blue-500 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Active Users</CardTitle>
            <Users className="h-5 w-5 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">{totalActiveUsers}</div>
            <p className="text-xs text-blue-600/70 mt-1">Last 30 days</p>
            {watchAnalytics?.active_users && watchAnalytics.active_users.length > 0 && (
              <div className="mt-3 flex items-center gap-2">
                <Badge variant="outline" className="text-xs text-blue-700 border-blue-200">
                  <Award className="w-3 h-3 mr-1" />
                  Top: {watchAnalytics.active_users[0].user_name}
                </Badge>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-l-4 border-l-green-500 bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-950/20 dark:to-emerald-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Sessions</CardTitle>
            <Play className="h-5 w-5 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{totalSessions.toLocaleString()}</div>
            <p className="text-xs text-green-600/70 mt-1">All playback methods</p>
            <div className="mt-3">
              <div className="text-xs text-green-700">Avg: {Math.round(totalSessions / Math.max(totalActiveUsers, 1))} per user</div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-l-4 border-l-purple-500 bg-gradient-to-br from-purple-50 to-violet-50 dark:from-purple-950/20 dark:to-violet-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Watch Time</CardTitle>
            <Clock className="h-5 w-5 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-purple-600">{formatWatchTime(totalWatchTime)}</div>
            <p className="text-xs text-purple-600/70 mt-1">Total hours watched</p>
            <div className="mt-3">
              <div className="text-xs text-purple-700">Avg session: {formatWatchTime(avgSessionLength)}</div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-l-4 border-l-orange-500 bg-gradient-to-br from-orange-50 to-red-50 dark:from-orange-950/20 dark:to-red-950/20">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Library Size</CardTitle>
            <Eye className="h-5 w-5 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">
              {watchAnalytics?.libraries?.reduce((sum: number, lib: any) => sum + lib.library_count, 0) || 0}
            </div>
            <p className="text-xs text-orange-600/70 mt-1">Total media items</p>
            <div className="mt-3">
              <div className="text-xs text-orange-700">
                {watchAnalytics?.libraries?.length || 0} libraries
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Library Overview */}
      {watchAnalytics?.libraries && watchAnalytics.libraries.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-indigo-500" />
              Library Overview
            </CardTitle>
            <CardDescription>Content distribution across your media libraries</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {watchAnalytics.libraries.map((library: any, index: number) => {
                const getLibraryIcon = (type: string) => {
                  switch (type?.toLowerCase()) {
                    case 'movies': return '🎬';
                    case 'tvshows': return '📺';
                    case 'music': return '🎵';
                    default: return '📚';
                  }
                };
                
                const totalContent = watchAnalytics.libraries.reduce((sum: number, lib: any) => sum + lib.library_count, 0);
                const percentage = totalContent > 0 ? (library.library_count / totalContent) * 100 : 0;
                
                return (
                  <div key={index} className="relative overflow-hidden rounded-xl border-2 border-indigo-200 bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-950/20 dark:to-purple-950/20 dark:border-indigo-800/50 p-4">
                    <div className="flex items-center gap-3 mb-3">
                      <div className="w-12 h-12 rounded-lg bg-white/80 dark:bg-gray-800/80 flex items-center justify-center border text-xl">
                        {getLibraryIcon(library.collection_type)}
                      </div>
                      <div className="flex-1">
                        <h3 className="font-semibold text-gray-900 dark:text-gray-100">{library.name}</h3>
                        <Badge variant="outline" className="text-xs mt-1 capitalize">
                          {library.collection_type || 'Mixed'}
                        </Badge>
                      </div>
                    </div>
                    
                    <div className="space-y-2 mb-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600 dark:text-gray-400">Items</span>
                        <span className="font-semibold">{library.library_count.toLocaleString()}</span>
                      </div>
                      {library.season_count > 0 && (
                        <>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-gray-600 dark:text-gray-400">Seasons</span>
                            <span className="font-semibold">{library.season_count}</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-gray-600 dark:text-gray-400">Episodes</span>
                            <span className="font-semibold">{library.episode_count}</span>
                          </div>
                        </>
                      )}
                    </div>
                    
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-gray-600 dark:text-gray-400">Library Share</span>
                        <span className="text-xs font-semibold">{percentage.toFixed(1)}%</span>
                      </div>
                      <Progress value={percentage} className="h-2" />
                    </div>
                    
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Advanced Analytics Charts */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        {/* Playback Performance Analysis */}
        {playbackMethodsChartData.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-green-500" />
                Playback Performance
              </CardTitle>
              <CardDescription>Distribution of playback methods and streaming quality</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {playbackMethodsChartData.map((method: any, index: number) => {
                  const total = playbackMethodsChartData.reduce((sum: number, m: any) => sum + m.value, 0);
                  const percentage = (method.value / total) * 100;
                  
                  return (
                    <div key={index} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div 
                            className="w-3 h-3 rounded-full" 
                            style={{ backgroundColor: method.fill }}
                          />
                          <span className="font-medium">{method.name}</span>
                        </div>
                        <div className="text-right">
                          <span className="font-semibold">{method.value.toLocaleString()}</span>
                          <span className="text-sm text-muted-foreground ml-2">({percentage.toFixed(1)}%)</span>
                        </div>
                      </div>
                      <Progress value={percentage} className="h-2" />
                    </div>
                  );
                })}
              </div>
              
              <div className="mt-6">
                <ChartContainer
                  config={{
                    DirectPlay: { label: "Direct Play", color: "#22c55e" },
                    DirectStream: { label: "Direct Stream", color: "#3b82f6" },
                    Transcode: { label: "Transcode", color: "#f59e0b" }
                  }}
                  className="h-[200px] w-full"
                >
                  <PieChart>
                    <Pie
                      data={playbackMethodsChartData}
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                      innerRadius={50}
                      dataKey="value"
                      strokeWidth={3}
                      stroke="hsl(var(--background))"
                    >
                      {playbackMethodsChartData.map((entry: any, index: number) => (
                        <Cell key={`cell-${index}`} fill={entry.fill} />
                      ))}
                    </Pie>
                    <ChartTooltip content={<ChartTooltipContent />} />
                  </PieChart>
                </ChartContainer>
              </div>
            </CardContent>
          </Card>
        )}

        {/* User Engagement Leaderboard */}
        {activeUsersChartData.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5 text-purple-500" />
                User Engagement Leaderboard
              </CardTitle>
              <CardDescription>Most active users by playback sessions</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {activeUsersChartData.slice(0, 5).map((user: any, index: number) => {
                  const maxPlays = Math.max(...activeUsersChartData.map((u: any) => u.plays));
                  const percentage = (user.plays / maxPlays) * 100;
                  
                  return (
                    <div key={index} className="flex items-center gap-3">
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm ${
                        index === 0 ? 'bg-yellow-100 text-yellow-700 border-2 border-yellow-300' :
                        index === 1 ? 'bg-gray-100 text-gray-700 border-2 border-gray-300' :
                        index === 2 ? 'bg-orange-100 text-orange-700 border-2 border-orange-300' :
                        'bg-blue-100 text-blue-700 border-2 border-blue-300'
                      }`}>
                        {index + 1}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center justify-between mb-1">
                          <span className="font-medium">{user.name}</span>
                          <span className="font-semibold">{user.plays}</span>
                        </div>
                        <Progress value={percentage} className="h-2" />
                      </div>
                    </div>
                  );
                })}
              </div>
              
              {activeUsersChartData.length > 5 && (
                <div className="mt-4 pt-4 border-t">
                  <p className="text-sm text-muted-foreground text-center">
                    +{activeUsersChartData.length - 5} more users
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Content Performance Dashboard */}
      {topContentData.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-red-500" />
              Content Performance Insights
            </CardTitle>
            <CardDescription>Top performing content and viewing trends</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Top Content List */}
              <div className="lg:col-span-2">
                <h4 className="font-semibold mb-4 flex items-center gap-2">
                  <Award className="w-4 h-4 text-yellow-500" />
                  Most Popular Content
                </h4>
                <div className="space-y-3">
                  {topContentData.slice(0, 6).map((item: any, index: number) => {
                    const maxPlays = Math.max(...topContentData.map((i: any) => i.plays));
                    const percentage = (item.plays / maxPlays) * 100;
                    
                    return (
                      <div key={index} className="flex items-center gap-3 p-3 rounded-lg bg-gradient-to-r from-gray-50 to-gray-100 dark:from-gray-800/50 dark:to-gray-900/50 border">
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm ${
                          index < 3 ? 'bg-gradient-to-br from-yellow-400 to-orange-500 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                        }`}>
                          {index + 1}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="font-medium truncate">{item.name}</p>
                          <div className="flex items-center gap-2 mt-1">
                            <Badge variant="outline" className="text-xs">
                              {item.type === 'Movie' ? '🎬' : '📺'} {item.type}
                            </Badge>
                            <span className="text-xs text-muted-foreground">{item.plays} plays</span>
                          </div>
                          <Progress value={percentage} className="h-1 mt-2" />
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
              
              {/* Quick Stats */}
              <div className="space-y-4">
                <h4 className="font-semibold mb-4 flex items-center gap-2">
                  <BarChart3 className="w-4 h-4 text-blue-500" />
                  Quick Insights
                </h4>
                <div className="space-y-3">
                  <div className="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800/50 rounded-lg p-3">
                    <div className="text-2xl font-bold text-blue-600 mb-1">
                      {topContentData.reduce((sum: number, item: any) => sum + item.plays, 0)}
                    </div>
                    <div className="text-sm text-blue-700 dark:text-blue-300">Total Views</div>
                  </div>
                  <div className="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800/50 rounded-lg p-3">
                    <div className="text-2xl font-bold text-green-600 mb-1">
                      {topContentData.length}
                    </div>
                    <div className="text-sm text-green-700 dark:text-green-300">Popular Items</div>
                  </div>
                  <div className="bg-purple-50 dark:bg-purple-950/20 border border-purple-200 dark:border-purple-800/50 rounded-lg p-3">
                    <div className="text-2xl font-bold text-purple-600 mb-1">
                      {Math.round(topContentData.reduce((sum: number, item: any) => sum + item.plays, 0) / topContentData.length)}
                    </div>
                    <div className="text-sm text-purple-700 dark:text-purple-300">Avg Views</div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Recent Activity Feed */}
      {watchAnalytics?.recently_watched && watchAnalytics.recently_watched.length > 0 && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5 text-blue-500" />
                  Recent Activity
                </CardTitle>
                <CardDescription>Latest viewing sessions across your server</CardDescription>
              </div>
              <Button variant="outline" size="sm" onClick={loadWatchAnalytics}>
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {watchAnalytics.recently_watched.slice(0, 8).map((item: any, index: number) => {
                const timeAgo = new Date(item.watched_at);
                const now = new Date();
                const diffHours = Math.floor((now.getTime() - timeAgo.getTime()) / (1000 * 60 * 60));
                const timeString = diffHours < 1 ? 'Just now' : 
                                 diffHours < 24 ? `${diffHours}h ago` : 
                                 `${Math.floor(diffHours / 24)}d ago`;
                
                return (
                  <div key={index} className="flex items-center gap-4 p-4 rounded-lg border bg-gradient-to-r from-gray-50/50 to-transparent dark:from-gray-800/50 hover:from-gray-100/50 dark:hover:from-gray-700/50 transition-all">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-100 to-purple-100 dark:from-blue-900/20 dark:to-purple-900/20 flex items-center justify-center border text-lg">
                      {item.item_type === 'Movie' ? '🎬' : '📺'}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-semibold text-foreground truncate">{item.item_name}</p>
                      <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                        <Badge variant="outline" className="text-xs">
                          {item.user_name}
                        </Badge>
                        <span>•</span>
                        <span className="capitalize">{item.item_type}</span>
                        <span>•</span>
                        <span>{item.library_name}</span>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold text-foreground">
                        {formatPlayDuration(item.play_duration)}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {timeString}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            
            {watchAnalytics.recently_watched.length > 8 && (
              <div className="mt-4 pt-4 border-t text-center">
                <p className="text-sm text-muted-foreground">
                  +{watchAnalytics.recently_watched.length - 8} more recent activities
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
};