import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Save, TestTube, Activity } from "lucide-react";
import { backendApi } from "@/lib/api";
import { toast } from "sonner";
import { getErrorMessage } from "@/utils/errorHandling";

interface JellystatSettingsData {
  jellystat_enabled: boolean;
  jellystat_host: string;
  jellystat_port: string;
  jellystat_use_ssl: boolean;
  jellystat_url: string;
  jellystat_api_key: string;
}

export default function JellystatSettings() {
  const [settings, setSettings] = useState<JellystatSettingsData>({
    jellystat_enabled: false,
    jellystat_host: "",
    jellystat_port: "3000",
    jellystat_use_ssl: false,
    jellystat_url: "",
    jellystat_api_key: "",
  });
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  // Load current settings
  useEffect(() => {
    const loadSettings = async () => {
      try {
        const response = await backendApi.getSettings();
        setSettings({
          jellystat_enabled: response.jellystat_enabled || false,
          jellystat_host: response.jellystat_host || "",
          jellystat_port: response.jellystat_port || "3000",
          jellystat_use_ssl: response.jellystat_use_ssl || false,
          jellystat_url: response.jellystat_url || "",
          jellystat_api_key: response.jellystat_api_key || "",
        });
      } catch (error) {
        const errorMessage = getErrorMessage(error);
        console.error("Failed to load Jellystat settings:", error);
        toast.error(`Failed to load Jellystat settings: ${errorMessage}`);
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await backendApi.updateSettings({
        jellystat_enabled: settings.jellystat_enabled,
        jellystat_host: settings.jellystat_host,
        jellystat_port: settings.jellystat_port,
        jellystat_use_ssl: settings.jellystat_use_ssl,
        jellystat_url: settings.jellystat_url,
        jellystat_api_key: settings.jellystat_api_key,
      });
      toast.success("Jellystat settings saved successfully");
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      console.error("Failed to save Jellystat settings:", error);
      toast.error(`Failed to save Jellystat settings: ${errorMessage}`);
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    if (!settings.jellystat_host || !settings.jellystat_port) {
      toast.error("Please fill in host and port before testing");
      return;
    }

    setTesting(true);
    try {
      // Construct the URL for testing
      const protocol = settings.jellystat_use_ssl ? "https" : "http";
      const testUrl = `${protocol}://${settings.jellystat_host}:${settings.jellystat_port}`;
      
      // TODO: Add actual test connection API call when available
      toast.info("Connection test functionality will be implemented");
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      console.error("Connection test failed:", error);
      toast.error(`Connection test failed: ${errorMessage}`);
    } finally {
      setTesting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="w-6 h-6 animate-spin" />
        <span className="ml-2">Loading Jellystat settings...</span>
      </div>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-3">
          <div className="p-2 bg-purple-500/20 rounded-lg border border-purple-500/30">
            <Activity className="w-5 h-5 text-purple-500" />
          </div>
          <div>
            <CardTitle>Jellystat Integration</CardTitle>
            <CardDescription>
              Configure Jellystat connection for viewing detailed media server analytics
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Enable/Disable Toggle */}
        <div className="flex items-center justify-between p-4 border rounded-lg bg-muted/50">
          <div className="space-y-1">
            <Label className="text-sm font-medium">Enable Jellystat Integration</Label>
            <p className="text-xs text-muted-foreground">
              Toggle Jellystat analytics integration for your media server
            </p>
          </div>
          <Switch
            checked={settings.jellystat_enabled}
            onCheckedChange={(checked) =>
              setSettings({ ...settings, jellystat_enabled: checked })
            }
          />
        </div>

        {/* Connection Settings */}
        {settings.jellystat_enabled && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="jellystat_host">Host/IP Address</Label>
                <Input
                  id="jellystat_host"
                  placeholder="192.168.1.100 or jellystat.local"
                  value={settings.jellystat_host}
                  onChange={(e) =>
                    setSettings({ ...settings, jellystat_host: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="jellystat_port">Port</Label>
                <Input
                  id="jellystat_port"
                  placeholder="3000"
                  value={settings.jellystat_port}
                  onChange={(e) =>
                    setSettings({ ...settings, jellystat_port: e.target.value })
                  }
                />
              </div>
            </div>

            {/* SSL Toggle */}
            <div className="flex items-center justify-between p-4 border rounded-lg bg-muted/50">
              <div className="space-y-1">
                <Label className="text-sm font-medium">Use SSL (HTTPS)</Label>
                <p className="text-xs text-muted-foreground">
                  Enable if your Jellystat instance uses HTTPS
                </p>
              </div>
              <Switch
                checked={settings.jellystat_use_ssl}
                onCheckedChange={(checked) =>
                  setSettings({ ...settings, jellystat_use_ssl: checked })
                }
              />
            </div>

            {/* API Key */}
            <div className="space-y-2">
              <Label htmlFor="jellystat_api_key">API Key</Label>
              <Input
                id="jellystat_api_key"
                type="password"
                placeholder="Enter your Jellystat API key"
                value={settings.jellystat_api_key}
                onChange={(e) =>
                  setSettings({ ...settings, jellystat_api_key: e.target.value })
                }
              />
              <p className="text-xs text-muted-foreground">
                You can generate an API key in your Jellystat settings
              </p>
            </div>

            {/* Full URL Override (Optional) */}
            <div className="space-y-2">
              <Label htmlFor="jellystat_url">Full URL (Optional)</Label>
              <Input
                id="jellystat_url"
                placeholder="https://jellystat.example.com (overrides host/port/SSL)"
                value={settings.jellystat_url}
                onChange={(e) =>
                  setSettings({ ...settings, jellystat_url: e.target.value })
                }
              />
              <p className="text-xs text-muted-foreground">
                If specified, this will override the host, port, and SSL settings above
              </p>
            </div>

            {/* Connection Info */}
            <div className="p-4 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded-lg">
              <p className="text-sm text-blue-700 dark:text-blue-300">
                <strong>Connection URL:</strong>{" "}
                {settings.jellystat_url || 
                  `${settings.jellystat_use_ssl ? "https" : "http"}://${settings.jellystat_host || "hostname"}:${settings.jellystat_port || "3000"}`
                }
              </p>
            </div>

            {/* Action Buttons */}
            <div className="flex gap-3 pt-4">
              <Button
                onClick={handleTestConnection}
                variant="outline"
                disabled={testing}
              >
                {testing ? (
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                ) : (
                  <TestTube className="w-4 h-4 mr-2" />
                )}
                Test Connection
              </Button>
            </div>
          </>
        )}

        {/* Save Button */}
        <div className="flex justify-end pt-6 border-t">
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <Loader2 className="w-4 h-4 animate-spin mr-2" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            Save Settings
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}