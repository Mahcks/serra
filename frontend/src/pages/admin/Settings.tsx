import { useParams, useNavigate } from "react-router-dom";
import { Settings as SettingsIcon, Server, HardDrive, Info, Globe } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function Settings() {
  const { tab } = useParams();
  const navigate = useNavigate();
  const activeTab = tab || "general";

  const handleTabChange = (value: string) => {
    if (value === "general") {
      navigate("/admin/settings");
    } else {
      navigate(`/admin/settings/${value}`);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="p-2 bg-primary/20 rounded-lg border">
          <SettingsIcon className="w-6 h-6 text-primary" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-foreground">Settings</h1>
          <p className="text-muted-foreground">Configure your Serra installation</p>
        </div>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="general" className="flex items-center gap-2">
            <Globe className="w-4 h-4" />
            General
          </TabsTrigger>
          <TabsTrigger value="services" className="flex items-center gap-2">
            <Server className="w-4 h-4" />
            Services
          </TabsTrigger>
          <TabsTrigger value="media-server" className="flex items-center gap-2">
            <HardDrive className="w-4 h-4" />
            Media Server
          </TabsTrigger>
          <TabsTrigger value="about" className="flex items-center gap-2">
            <Info className="w-4 h-4" />
            About
          </TabsTrigger>
        </TabsList>

        {/* General Tab */}
        <TabsContent value="general" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>General Settings</CardTitle>
              <CardDescription>
                Configure basic application settings and preferences
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm text-muted-foreground">
                General settings content will go here (application title, language, theme, etc.)
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Services Tab */}
        <TabsContent value="services" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Download Services</CardTitle>
              <CardDescription>
                Configure Radarr and Sonarr connections and settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Radarr and Sonarr configuration will go here
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Media Server Tab */}
        <TabsContent value="media-server" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Media Server</CardTitle>
              <CardDescription>
                Configure Emby or Jellyfin connection and settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Emby/Jellyfin configuration will go here
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* About Tab */}
        <TabsContent value="about" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>About Serra</CardTitle>
              <CardDescription>
                System information and version details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Version information and system status will go here
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
