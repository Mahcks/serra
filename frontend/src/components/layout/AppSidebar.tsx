import {
  Home,
  Search,
  Bell,
  LogOut,
  User,
  Settings,
  Film,
  Tv,
  Calendar,
  Users,
  ChartArea,
} from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { Avatar } from "@/components/ui/avatar";
import { useAuth } from "@/lib/auth";
import { ThemeToggle } from "@/components/shared/ThemeToggle";

interface AppSidebarProps {
  onLogout: () => Promise<void>;
}

export function AppSidebar({ onLogout }: AppSidebarProps) {
  const location = useLocation();
  const { user } = useAuth();

  const handleLogout = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      console.log("🚪 Sidebar logout button clicked");
      await onLogout();
    } catch (error) {
      console.error("❌ Logout error in sidebar:", error);
    }
  };

  const mainNavItems = [
    {
      title: "Dashboard",
      path: "/dashboard",
      icon: Home,
      isActive: location.pathname === "/dashboard" || location.pathname === "/",
    },
  ];

  const discoverNavItems = [
    {
      title: "Discover",
      path: "/requests",
      icon: Search,
      isActive: location.pathname === "/requests" && !location.search,
    },
    {
      title: "Movies",
      path: "/requests?tab=movies",
      icon: Film,
      isActive: location.search.includes("tab=movies"),
    },
    {
      title: "Series",
      path: "/requests?tab=series",
      icon: Tv,
      isActive: location.search.includes("tab=series"),
    },
    {
      title: "My Requests",
      path: "/requests?tab=requests",
      icon: Calendar,
      isActive: location.search.includes("tab=requests"),
    },
  ];

  const adminNavItems = [
    {
      title: "Users",
      path: "/admin/users",
      icon: User,
      isActive: location.pathname === "/admin/users",
    },
    {
      title: "Requests",
      path: "/admin/requests",
      icon: Users,
      isActive: location.pathname === "/admin/requests",
    },
    {
      title: "Analytics",
      path: "/admin/analytics",
      icon: ChartArea,
      isActive: location.pathname.startsWith("/admin/analytics"),
    },
    {
      title: "Settings",
      path: "/admin/settings",
      icon: Settings,
      isActive: location.pathname.startsWith("/admin/settings"),
    },
  ];

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2">
          <Avatar 
            src={user?.avatar_url ? `v1${user.avatar_url}` : undefined}
            alt={user?.username || "User"}
            fallback={user?.username?.charAt(0) || "U"}
            size="sm"
            className="border border-border"
          />
          <div className="grid flex-1 text-left text-sm leading-tight">
            <span className="truncate font-semibold">
              Welcome back, {user?.username}
            </span>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent className="overflow-x-hidden">
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainNavItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton
                    asChild
                    isActive={item.isActive}
                    tooltip={item.title}
                  >
                    <Link to={item.path}>
                      <item.icon className="size-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        {/* Discover Section */}
        <SidebarGroup>
          <SidebarGroupLabel>Discover</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {discoverNavItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton
                    asChild
                    isActive={item.isActive}
                    tooltip={item.title}
                  >
                    <Link to={item.path}>
                      <item.icon className="size-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        {/* Admin tools/pages */}
        <SidebarGroup>
          <SidebarGroupLabel>Admin</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {adminNavItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton
                    asChild
                    isActive={item.isActive}
                    tooltip={item.title}
                  >
                    <Link to={item.path}>
                      <item.icon className="size-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarSeparator className="mx-0.6"/>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton tooltip="Notifications">
              <Bell className="size-4" />
              <span>Notifications</span>
              <div className="ml-auto size-2 rounded-full bg-destructive animate-pulse" />
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton tooltip="Theme">
              <ThemeToggle />
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton asChild tooltip="Sign Out">
              <button
                onClick={handleLogout}
                className="w-full flex items-center gap-2 cursor-pointer"
                type="button"
              >
                <LogOut className="size-4" />
                <span>Sign Out</span>
              </button>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
