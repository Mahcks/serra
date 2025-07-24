import { useState, useEffect } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { AlertTriangle, Loader2 } from "lucide-react";
import { toast } from "sonner";
import DynamicDefaultPermissions from "@/components/admin/DynamicDefaultPermissions";

export default function AdminUsersSettings() {
  const [enableLocalAuth, setEnableLocalAuth] = useState(false);
  const [enableEmbyAuth, setEnableEmbyAuth] = useState(true);
  const [globalMovieRequestLimit, setGlobalMovieRequestLimit] = useState(0);
  const [globalSeriesRequestLimit, setGlobalSeriesRequestLimit] = useState(0);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Validation: at least one auth method must be enabled
  const isValid = enableLocalAuth || enableEmbyAuth;

  // Load settings from API
  useEffect(() => {
    loadSettings();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const loadSettings = async (showLoading = true) => {
    try {
      if (showLoading) {
        setLoading(true);
      }
      const response = await fetch('/v1/settings/system', {
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error('Failed to load system settings');
      }
      
      const data = await response.json();
      setEnableLocalAuth(data.enable_local_auth || false);
      setEnableEmbyAuth(data.enable_media_server_auth !== false); // Default true if not set
      setGlobalMovieRequestLimit(data.global_movie_request_limit || 0);
      setGlobalSeriesRequestLimit(data.global_series_request_limit || 0);
    } catch (error) {
      console.error('Failed to load system settings:', error);
      if (showLoading) {
        toast.error("Failed to load system settings. Using defaults.");
      }
      throw error; // Re-throw so caller can handle it
    } finally {
      if (showLoading) {
        setLoading(false);
      }
    }
  };

  const handleResetDefaults = () => {
    setEnableLocalAuth(false);
    setEnableEmbyAuth(true);
    setGlobalMovieRequestLimit(0);
    setGlobalSeriesRequestLimit(0);
  };


  const handleSaveChanges = async () => {
    if (!isValid) return;
    
    try {
      setSaving(true);
      const settings = {
        'enable_local_auth': enableLocalAuth,
        'enable_media_server_auth': enableEmbyAuth,
        'global_movie_request_limit': globalMovieRequestLimit,
        'global_series_request_limit': globalSeriesRequestLimit,
      };

      const response = await fetch('/v1/settings/system', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          settings: settings,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Failed to save system settings' }));
        throw new Error(errorData.message || 'Failed to save system settings');
      }

      toast.success("Authentication settings saved successfully.");
      
      // Reload settings from server to ensure UI is in sync
      try {
        await loadSettings(false); // Don't show loading spinner
      } catch (reloadError) {
        console.warn('Failed to reload settings after save:', reloadError);
        // Don't show error to user since the save was successful
      }
    } catch (error) {
      console.error('Failed to save system settings:', error);
      const errorMessage = error instanceof Error ? error.message : "Failed to save system settings. Please try again.";
      toast.error(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin mr-3" />
        <span>Loading settings...</span>
      </div>
    );
  }
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground mb-2">User Settings</h1>
        <p className="text-muted-foreground">
          Manage user preferences and account settings
        </p>
      </div>

      <Separator />

      {/* Authentication Methods Section */}
      <Card>
        <CardHeader>
          <CardTitle>Login Methods</CardTitle>
          <CardDescription>
            Configure login methods for users.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Validation Warning */}
          {!isValid && (
            <div className="flex items-center gap-2 p-3 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800/50 rounded-lg">
              <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <span className="text-sm font-medium text-amber-600 dark:text-amber-400">
                At least one authentication method must be selected.
              </span>
            </div>
          )}

          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox 
                id="local-sign-in" 
                checked={enableLocalAuth}
                onCheckedChange={(checked) => setEnableLocalAuth(checked === true)}
              />
              <Label htmlFor="local-sign-in" className="text-sm font-medium">
                Enable Local Sign-In
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Allow users to sign in using their email address and password
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox 
                id="emby-jellyfin-login" 
                checked={enableEmbyAuth}
                onCheckedChange={(checked) => setEnableEmbyAuth(checked === true)}
              />
              <Label htmlFor="emby-jellyfin-login" className="text-sm font-medium">
                Enable Emby Sign-In
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Allow users to sign in using their Emby account
            </p>
          </div>
          

          {/* Reset Button */}
          <div className="flex justify-start pt-4 border-t">
            <Button variant="outline" onClick={handleResetDefaults}>
              Reset to Defaults
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Request Limits Section */}
      <Card>
        <CardHeader>
          <CardTitle>Global Request Limits</CardTitle>
          <CardDescription>
            Set maximum number of requests per user (0 = unlimited)
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="movie-limit" className="text-sm font-medium">
              Movie Request Limit
            </Label>
            <Input
              id="movie-limit"
              type="number"
              min="0"
              value={globalMovieRequestLimit}
              onChange={(e) => setGlobalMovieRequestLimit(parseInt(e.target.value) || 0)}
              className="w-32"
            />
            <p className="text-xs text-muted-foreground">
              Maximum number of movie requests per user (0 for unlimited)
            </p>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="series-limit" className="text-sm font-medium">
              Series Request Limit
            </Label>
            <Input
              id="series-limit"
              type="number"
              min="0"
              value={globalSeriesRequestLimit}
              onChange={(e) => setGlobalSeriesRequestLimit(parseInt(e.target.value) || 0)}
              className="w-32"
            />
            <p className="text-xs text-muted-foreground">
              Maximum number of series requests per user (0 for unlimited)
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Default Permissions Section */}
      <DynamicDefaultPermissions />

      {/* User Registration Section */}
      <Card>
        <CardHeader>
          <CardTitle>User Registration</CardTitle>
          <CardDescription>
            Configure how new users can register and access the system
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="allow-registration" />
              <Label htmlFor="allow-registration" className="text-sm font-medium">
                Allow user registration
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              When enabled, users can create new accounts without admin intervention
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="email-verification" />
              <Label htmlFor="email-verification" className="text-sm font-medium">
                Require email verification
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Users must verify their email address before they can access the system
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="admin-approval" />
              <Label htmlFor="admin-approval" className="text-sm font-medium">
                Require admin approval for new accounts
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              All new user registrations must be manually approved by an administrator
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Notifications Section */}
      <Card>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>
            Control notification settings and user communication preferences
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center space-x-3">
            <Checkbox id="user-notifications" />
            <Label htmlFor="user-notifications" className="text-sm font-medium">
              Enable user notifications
            </Label>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="email-notifications" />
              <Label htmlFor="email-notifications" className="text-sm font-medium">
                Send email notifications for requests
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Users will receive email updates when their requests are approved, denied, or fulfilled
            </p>
          </div>
          
          <div className="flex items-center space-x-3">
            <Checkbox id="admin-notifications" />
            <Label htmlFor="admin-notifications" className="text-sm font-medium">
              Notify admins of new user registrations
            </Label>
          </div>
        </CardContent>
      </Card>

      {/* Account Security Section */}
      <Card>
        <CardHeader>
          <CardTitle>Account Security</CardTitle>
          <CardDescription>
            Security and authentication settings for user accounts
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="force-password-reset" />
              <Label htmlFor="force-password-reset" className="text-sm font-medium">
                Force password reset on first login
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              New users will be required to change their password when they first log in
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="session-timeout" />
              <Label htmlFor="session-timeout" className="text-sm font-medium">
                Enable automatic session timeout
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Users will be automatically logged out after a period of inactivity (default: 24 hours)
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="two-factor" />
              <Label htmlFor="two-factor" className="text-sm font-medium">
                Enable two-factor authentication (when available)
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Requires users to provide a second form of authentication for enhanced security
            </p>
          </div>
        </CardContent>
      </Card>

      {/* User Permissions Section */}
      <Card>
        <CardHeader>
          <CardTitle>Default Permissions</CardTitle>
          <CardDescription>
            Set default permissions for newly registered users
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center space-x-3">
            <Checkbox id="default-request" />
            <Label htmlFor="default-request" className="text-sm font-medium">
              Allow requests by default
            </Label>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="default-4k" />
              <Label htmlFor="default-4k" className="text-sm font-medium">
                Allow 4K requests by default
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              New users can request 4K/UHD quality content without special permissions
            </p>
          </div>
          
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Checkbox id="default-auto-approve" />
              <Label htmlFor="default-auto-approve" className="text-sm font-medium">
                Auto-approve requests from new users
              </Label>
            </div>
            <p className="text-xs text-muted-foreground ml-6">
              Automatically approve all requests from newly registered users (not recommended for large instances)
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Save Changes - Fixed at Bottom */}
      <div className="flex items-center justify-end gap-3 py-6 border-t border-border/60 bg-background/95 backdrop-blur-sm sticky bottom-0">
        {!isValid && (
          <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400 mr-auto">
            <AlertTriangle className="h-4 w-4" />
            <span className="text-sm font-medium">At least one authentication method must be enabled</span>
          </div>
        )}
        <Button 
          disabled={!isValid || saving}
          className="min-w-[140px]"
          onClick={handleSaveChanges}
        >
          {saving ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
              Saving...
            </>
          ) : (
            'Save Changes'
          )}
        </Button>
      </div>
    </div>
  );
}
