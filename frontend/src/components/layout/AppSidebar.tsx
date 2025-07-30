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
  Mail,
} from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { useState, useEffect } from "react";
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
import { NotificationsPanel } from "./NotificationsPanel";
import { backendApi } from "@/lib/api";
import { useWebSocket } from "@/lib/WebSocketContext";
import { OpcodeNotification } from "@/types";

interface AppSidebarProps {
  onLogout: () => Promise<void>;
}

export function AppSidebar({ onLogout }: AppSidebarProps) {
  const location = useLocation();
  const { user } = useAuth();
  const { subscribe } = useWebSocket();
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    const loadUnreadCount = async () => {
      try {
        const data = await backendApi.getUnreadCount();
        console.log('Unread count API response:', data);
        // Handle both direct number and object format
        const count = typeof data === 'number' ? data : (data?.unread_count || data?.UnreadCount || 0);
        setUnreadCount(count);
      } catch (error) {
        console.error('Failed to load unread count:', error);
      }
    };

    loadUnreadCount();
    // Poll for updates every 30 seconds
    const interval = setInterval(loadUnreadCount, 30000);

    // Listen for real-time notification updates via WebSocket
    const unsubscribeWebSocket = subscribe((message) => {
      if (message.op === OpcodeNotification) {
        console.log('📱 Received notification WebSocket message:', message.d);
        // Refresh unread count when we receive notification updates
        loadUnreadCount();
      }
    });

    return () => {
      clearInterval(interval);
      unsubscribeWebSocket();
    };
  }, [subscribe]);

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
      icon: Users,
      isActive: location.pathname === "/admin/users",
    },
    {
      title: "Requests",
      path: "/admin/requests",
      icon: User,
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
            <SidebarMenuButton 
              tooltip="Notifications"
              onClick={() => setNotificationsOpen(true)}
              className="relative"
            >
              <div className="relative">
                <Bell className="size-4" />
                {unreadCount > 0 && (
                  <div className="absolute -top-2 -right-2 size-5 rounded-full bg-destructive text-destructive-foreground text-xs flex items-center justify-center min-w-[20px] h-5">
                    {unreadCount > 99 ? '99+' : unreadCount}
                  </div>
                )}
              </div>
              <span>Notifications</span>
              {unreadCount > 0 && (
                <div className="ml-auto size-5 rounded-full bg-destructive text-destructive-foreground text-xs flex items-center justify-center">
                  {unreadCount > 99 ? '99+' : unreadCount}
                </div>
              )}
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton asChild tooltip="User Settings">
              <Link to="/settings">
                <User className="size-4" />
                <span>Settings</span>
              </Link>
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
      
      <NotificationsPanel 
        isOpen={notificationsOpen}
        onClose={() => {
          setNotificationsOpen(false);
          // Refresh unread count when panel closes
          backendApi.getUnreadCount().then(data => {
            const count = typeof data === 'number' ? data : (data?.unread_count || data?.UnreadCount || 0);
            setUnreadCount(count);
          }).catch(console.error);
        }}
      />
    </Sidebar>
  );
}
