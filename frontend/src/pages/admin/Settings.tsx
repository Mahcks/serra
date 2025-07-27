import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useNavigate, useLocation } from "react-router-dom";
import React, { useMemo, useCallback, lazy, Suspense } from "react";

// Lazy load components
const AdminGeneralSettings = lazy(() => import("@/components/admin/settings/GeneralSettings"));
const AdminUsersSettings = lazy(() => import("@/components/admin/settings/UsersSettings"));
const AdminMediaServerSettings = lazy(() => import("@/components/admin/settings/MediaServerSettings"));
const AdminServicesSettings = lazy(() => import("@/components/admin/settings/ServicesSettings"));
const EmailSettings = lazy(() => import("@/components/admin/settings/EmailSettings"));
const AdminAboutSettings = lazy(() => import("@/components/admin/settings/AboutSettings"));

// Loading component
const LoadingState = () => (
  <div className="flex items-center justify-center p-8">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    <span className="ml-3 text-muted-foreground">Loading...</span>
  </div>
);

const Settings = React.memo(() => {
  const navigate = useNavigate();
  const location = useLocation();

  const currentTab = useMemo(() => {
    const path = location.pathname;
    if (path.endsWith("/users")) return "users";
    if (path.endsWith("/media-server")) return "media-server";
    if (path.endsWith("/services")) return "services";
    if (path.endsWith("/email")) return "email";
    if (path.endsWith("/about")) return "about";
    return "general";
  }, [location.pathname]);

  const handleTabChange = useCallback(
    (tab: string) => {
      const basePath = "/admin/settings";
      const newPath = tab === "general" ? basePath : `${basePath}/${tab}`;
      navigate(newPath, { replace: true });
    },
    [navigate]
  );

  return (
    <div className="container mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground mb-2">
          Serra Settings
        </h1>
        <p className="text-muted-foreground">Description here</p>
      </div>

      <Tabs
        value={currentTab}
        onValueChange={handleTabChange}
        className="w-full"
      >
        <TabsList className="grid w-full grid-cols-6 mb-8">
          <TabsTrigger value="general" className="flex items-center gap-2">
            General
          </TabsTrigger>
          <TabsTrigger value="users" className="flex items-center gap-2">
            Users
          </TabsTrigger>
          <TabsTrigger value="media-server" className="flex items-center gap-2">
            Media Server
          </TabsTrigger>
          <TabsTrigger value="services" className="flex items-center gap-2">
            Services
          </TabsTrigger>
          <TabsTrigger value="email" className="flex items-center gap-2">
            Email
          </TabsTrigger>
          <TabsTrigger value="about" className="flex items-center gap-2">
            About
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <AdminGeneralSettings />
          </Suspense>
        </TabsContent>

        <TabsContent value="users" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <AdminUsersSettings />
          </Suspense>
        </TabsContent>

        <TabsContent value="media-server" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <AdminMediaServerSettings />
          </Suspense>
        </TabsContent>
        
        <TabsContent value="services" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <AdminServicesSettings />
          </Suspense>
        </TabsContent>

        <TabsContent value="email" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <EmailSettings />
          </Suspense>
        </TabsContent>

        <TabsContent value="about" className="mt-0">
          <Suspense fallback={<LoadingState />}>
            <AdminAboutSettings />
          </Suspense>
        </TabsContent>
      </Tabs>
    </div>
  );
});

export default Settings;
