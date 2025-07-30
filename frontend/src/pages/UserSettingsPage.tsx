import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import { Avatar } from '@/components/ui/avatar';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { AlertCircle, Save, User, Bell, Shield, Palette, Check } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { backendApi } from '@/lib/api';

interface UserSettings {
  profile: {
    id: string;
    username: string;
    email: string;
    avatar_url: string;
    user_type: string;
    created_at: string;
  };
  permissions: {
    id: string;
    name: string;
    description: string;
    category: string;
    dangerous: boolean;
  }[];
  notification_preferences: {
    requests_approved: boolean;
    requests_denied: boolean;
    download_completed: boolean;
    media_available: boolean;
    system_alerts: boolean;
    web_notifications: boolean;
    email_notifications: boolean;
    push_notifications: boolean;
  };
  account_settings: {
    language: string;
    theme: string;
    timezone: string;
    date_format: string;
    time_format: string;
  };
  privacy_settings: {
    show_online_status: boolean;
    show_watch_history: boolean;
    show_request_history: boolean;
  };
}

export default function UserSettingsPage() {
  const { user } = useAuth();
  const [settings, setSettings] = useState<UserSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [savedSuccessfully, setSavedSuccessfully] = useState(false);

  // Load user settings
  useEffect(() => {
    const loadSettings = async () => {
      try {
        const data = await backendApi.getUserSettings();
        setSettings(data);
      } catch (error) {
        console.error('Failed to load settings:', error);
        toast.error('Failed to load user settings');
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, []);

  const updateSettings = (section: keyof UserSettings, updates: any) => {
    if (!settings) return;

    setSettings({
      ...settings,
      [section]: {
        ...settings[section],
        ...updates,
      },
    });
    setHasChanges(true);
  };

  const saveSettings = async () => {
    if (!settings || !hasChanges) return;

    setSaving(true);
    try {
      await backendApi.updateUserSettings({
        profile: {
          email: settings.profile.email,
          avatar_url: settings.profile.avatar_url,
        },
        notification_preferences: settings.notification_preferences,
        account_settings: settings.account_settings,
        privacy_settings: settings.privacy_settings,
      });

      setHasChanges(false);
      setSavedSuccessfully(true);
      toast.success('Settings saved successfully');
      
      // Hide success indicator after 2 seconds
      setTimeout(() => setSavedSuccessfully(false), 2000);
    } catch (error) {
      console.error('Failed to save settings:', error);
      toast.error('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="space-y-6">
          <div className="space-y-2">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-4 w-96" />
          </div>
          
          <Tabs defaultValue="profile" className="space-y-4">
            <div className="flex space-x-1">
              {['Profile', 'Notifications', 'Privacy', 'Appearance'].map((tab) => (
                <Skeleton key={tab} className="h-10 w-24" />
              ))}
            </div>
            
            <Card>
              <CardHeader>
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-64" />
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <Skeleton className="h-16 w-16 rounded-full" />
                    <div className="space-y-2 flex-1">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-16" />
                    <Skeleton className="h-10 w-full" />
                  </div>
                  
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-20" />
                    <Skeleton className="h-10 w-full" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </Tabs>
        </div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="container mx-auto py-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5" />
              Error Loading Settings
            </CardTitle>
            <CardDescription>
              We couldn't load your user settings. Please try refreshing the page.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold">User Settings</h1>
          <p className="text-muted-foreground mt-1">
            Manage your account preferences and notifications
          </p>
        </div>
        {(hasChanges || savedSuccessfully) && (
          <Button 
            onClick={saveSettings} 
            disabled={saving || (!hasChanges && !savedSuccessfully)}
            variant={savedSuccessfully ? "default" : "default"}
            className={savedSuccessfully ? "bg-green-600 hover:bg-green-700" : ""}
          >
            {saving ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 mr-2 border-2 border-background border-t-current" />
                Saving...
              </>
            ) : savedSuccessfully ? (
              <>
                <Check className="h-4 w-4 mr-2" />
                Saved!
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </>
            )}
          </Button>
        )}
      </div>

      <Tabs defaultValue="profile" className="space-y-4">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="profile" className="flex items-center gap-2">
            <User className="h-4 w-4" />
            Profile
          </TabsTrigger>
          <TabsTrigger value="notifications" className="flex items-center gap-2">
            <Bell className="h-4 w-4" />
            Notifications
          </TabsTrigger>
          <TabsTrigger value="privacy" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Privacy
          </TabsTrigger>
          <TabsTrigger value="appearance" className="flex items-center gap-2">
            <Palette className="h-4 w-4" />
            Appearance
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription>
                View and update your basic profile information
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={settings.profile.username}
                    disabled
                    className="bg-muted"
                  />
                  <p className="text-sm text-muted-foreground">
                    Username cannot be changed
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="user-type">Account Type</Label>
                  <div className="flex items-center gap-2">
                    <Badge variant={settings.profile.user_type === 'local' ? 'secondary' : 'default'}>
                      {settings.profile.user_type === 'local' ? 'Local Account' : 'Media Server'}
                    </Badge>
                  </div>
                </div>
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="email">Email Address</Label>
                <Input
                  id="email"
                  type="email"
                  value={settings.profile.email}
                  onChange={(e) => updateSettings('profile', { email: e.target.value })}
                  placeholder="Enter your email address"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="avatar">Profile Picture</Label>
                <div className="flex items-start gap-4">
                  <div className="flex flex-col items-center gap-2">
                    <Avatar 
                      src={settings.profile.avatar_url ? `v1${settings.profile.avatar_url}` : undefined}
                      alt={settings.profile.username || "User"}
                      fallback={settings.profile.username?.charAt(0) || "U"}
                      size="xl"
                      className="border-2 border-border"
                    />
                    <p className="text-xs text-muted-foreground text-center">Preview</p>
                  </div>
                  <div className="flex-1 space-y-2">
                    <Label htmlFor="avatar-input">Avatar URL</Label>
                    <Input
                      id="avatar-input"
                      type="url"
                      value={settings.profile.avatar_url || ''}
                      onChange={(e) => updateSettings('profile', { avatar_url: e.target.value })}
                      placeholder={settings.profile.user_type === 'local' ? 'https://example.com/avatar.jpg' : 'Managed by media server'}
                      disabled={settings.profile.user_type !== 'local'}
                    />
                    <p className="text-sm text-muted-foreground">
                      {settings.profile.user_type === 'local' 
                        ? 'Enter a URL to your profile picture (optional)'
                        : 'Avatar is managed by your media server account'
                      }
                    </p>
                  </div>
                </div>
              </div>

              <Separator />

              <div className="space-y-3">
                <Label>Permissions</Label>
                {settings.permissions.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No permissions assigned</p>
                ) : (
                  <div className="space-y-4">
                    {/* Group permissions by category */}
                    {Object.entries(
                      settings.permissions.reduce((acc, permission) => {
                        const category = permission.category || 'General';
                        if (!acc[category]) acc[category] = [];
                        acc[category].push(permission);
                        return acc;
                      }, {} as Record<string, typeof settings.permissions>)
                    ).map(([category, categoryPermissions]) => (
                      <div key={category} className="space-y-2">
                        <div className="text-sm font-medium text-muted-foreground">{category}</div>
                        <div className="space-y-1">
                          {categoryPermissions.map((permission) => (
                            <div key={permission.id} className="flex items-start space-x-3 p-3 rounded-lg border bg-muted/20">
                              <Shield className="h-4 w-4 mt-0.5 text-muted-foreground" />
                              <div className="flex-1 min-w-0">
                                <div className="font-medium text-sm">{permission.name}</div>
                                <div className="text-sm text-muted-foreground">{permission.description}</div>
                              </div>
                              {permission.dangerous && (
                                <Badge variant="destructive" className="text-xs">
                                  Admin
                                </Badge>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="text-sm text-muted-foreground">
                <p>Account created: {new Date(settings.profile.created_at).toLocaleDateString()}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Notification Preferences</CardTitle>
              <CardDescription>
                Choose what notifications you want to receive and when
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <h4 className="font-medium">Notification Types</h4>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Request Approved</Label>
                      <p className="text-sm text-muted-foreground">
                        When your media requests are approved
                      </p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.requests_approved}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { requests_approved: checked })
                      }
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Request Denied</Label>
                      <p className="text-sm text-muted-foreground">
                        When your media requests are denied
                      </p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.requests_denied}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { requests_denied: checked })
                      }
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Download Completed</Label>
                      <p className="text-sm text-muted-foreground">
                        When downloads finish processing
                      </p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.download_completed}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { download_completed: checked })
                      }
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Media Available</Label>
                      <p className="text-sm text-muted-foreground">
                        When requested media becomes available
                      </p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.media_available}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { media_available: checked })
                      }
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>System Alerts</Label>
                      <p className="text-sm text-muted-foreground">
                        Important system notifications and alerts
                      </p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.system_alerts}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { system_alerts: checked })
                      }
                    />
                  </div>
                </div>
              </div>

              <Separator />

              <div className="space-y-3">
                <h4 className="font-medium">Delivery Methods</h4>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Label>Web Notifications</Label>
                    <Switch
                      checked={settings.notification_preferences.web_notifications}
                      onCheckedChange={(checked) =>
                        updateSettings('notification_preferences', { web_notifications: checked })
                      }
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Email Notifications</Label>
                      <p className="text-sm text-muted-foreground">Coming soon</p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.email_notifications}
                      disabled
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Push Notifications</Label>
                      <p className="text-sm text-muted-foreground">Coming soon</p>
                    </div>
                    <Switch
                      checked={settings.notification_preferences.push_notifications}
                      disabled
                    />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="privacy" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Privacy Settings</CardTitle>
              <CardDescription>
                Control what information is visible to other users
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Show Online Status</Label>
                  <p className="text-sm text-muted-foreground">
                    Let others see when you're online
                  </p>
                </div>
                <Switch
                  checked={settings.privacy_settings.show_online_status}
                  onCheckedChange={(checked) =>
                    updateSettings('privacy_settings', { show_online_status: checked })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Show Watch History</Label>
                  <p className="text-sm text-muted-foreground">
                    Allow others to see your recently watched content
                  </p>
                </div>
                <Switch
                  checked={settings.privacy_settings.show_watch_history}
                  onCheckedChange={(checked) =>
                    updateSettings('privacy_settings', { show_watch_history: checked })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Show Request History</Label>
                  <p className="text-sm text-muted-foreground">
                    Let others see your media requests
                  </p>
                </div>
                <Switch
                  checked={settings.privacy_settings.show_request_history}
                  onCheckedChange={(checked) =>
                    updateSettings('privacy_settings', { show_request_history: checked })
                  }
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="appearance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Appearance & Localization</CardTitle>
              <CardDescription>
                Customize how the application looks and feels
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Theme</Label>
                  <Select
                    value={settings.account_settings.theme}
                    onValueChange={(value) =>
                      updateSettings('account_settings', { theme: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="light">Light</SelectItem>
                      <SelectItem value="dark">Dark</SelectItem>
                      <SelectItem value="system">System</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label>Language</Label>
                  <Select
                    value={settings.account_settings.language}
                    onValueChange={(value) =>
                      updateSettings('account_settings', { language: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="en">English</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Date Format</Label>
                  <Select
                    value={settings.account_settings.date_format}
                    onValueChange={(value) =>
                      updateSettings('account_settings', { date_format: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="YYYY-MM-DD">YYYY-MM-DD</SelectItem>
                      <SelectItem value="MM/DD/YYYY">MM/DD/YYYY</SelectItem>
                      <SelectItem value="DD/MM/YYYY">DD/MM/YYYY</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label>Time Format</Label>
                  <Select
                    value={settings.account_settings.time_format}
                    onValueChange={(value) =>
                      updateSettings('account_settings', { time_format: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="12h">12 Hour</SelectItem>
                      <SelectItem value="24h">24 Hour</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label>Timezone</Label>
                <Select
                  value={settings.account_settings.timezone}
                  onValueChange={(value) =>
                    updateSettings('account_settings', { timezone: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="UTC">UTC</SelectItem>
                    <SelectItem value="America/New_York">Eastern Time</SelectItem>
                    <SelectItem value="America/Chicago">Central Time</SelectItem>
                    <SelectItem value="America/Denver">Mountain Time</SelectItem>
                    <SelectItem value="America/Los_Angeles">Pacific Time</SelectItem>
                    <SelectItem value="Europe/London">London</SelectItem>
                    <SelectItem value="Europe/Paris">Paris</SelectItem>
                    <SelectItem value="Asia/Tokyo">Tokyo</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {(hasChanges || savedSuccessfully) && (
        <div className="fixed bottom-4 right-4 z-50">
          <Button 
            onClick={saveSettings} 
            disabled={saving || (!hasChanges && !savedSuccessfully)} 
            size="lg" 
            className={`shadow-lg transition-colors ${savedSuccessfully ? "bg-green-600 hover:bg-green-700" : ""}`}
          >
            {saving ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 mr-2 border-2 border-background border-t-current" />
                Saving...
              </>
            ) : savedSuccessfully ? (
              <>
                <Check className="h-4 w-4 mr-2" />
                Saved!
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      )}
    </div>
  );
}