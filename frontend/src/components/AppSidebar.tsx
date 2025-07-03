import { Home, Search, Bell, LogOut, User, Activity, Download, Settings } from "lucide-react";
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

interface AppSidebarProps {
  onLogout: () => Promise<void>;
}

export function AppSidebar({ onLogout }: AppSidebarProps) {
  const location = useLocation();

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
    {
      title: "Requests",
      path: "/requests", 
      icon: Search,
      isActive: location.pathname === "/requests",
    },
  ];

  const toolsNavItems = [
    {
      title: "Downloads",
      path: "/downloads",
      icon: Download,
      isActive: location.pathname === "/downloads",
    },
    {
      title: "Settings",
      path: "/settings",
      icon: Settings,
      isActive: location.pathname === "/settings",
    },
  ];

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2 px-1">
          <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Activity className="size-4" />
          </div>
          <div className="grid flex-1 text-left text-sm leading-tight">
            <span className="truncate font-semibold">Serra</span>
            <span className="truncate text-xs text-muted-foreground">Media Server</span>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
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

        <SidebarGroup>
          <SidebarGroupLabel>Tools</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {toolsNavItems.map((item) => (
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
        
        <SidebarSeparator />
        
        <SidebarMenu>
          <SidebarMenuItem>
            <div className="flex items-center gap-2 px-1 py-1.5">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-muted">
                <User className="size-4 text-muted-foreground" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">Welcome back</span>
                <span className="truncate text-xs text-muted-foreground">Ready to innovate</span>
              </div>
            </div>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}