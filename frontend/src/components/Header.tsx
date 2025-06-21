import { LogOut, Bell, Search, User } from "lucide-react";
import { Link } from "react-router-dom";

interface HeaderProps {
  onLogout: () => void;
  currentPath?: string;
}

export function Header({ onLogout, currentPath = "/dashboard" }: HeaderProps) {
  return (
    <header className="sticky top-0 z-50 backdrop-blur-sm bg-background/95 border-b border-border shadow-xl">
      {/* Subtle gradient overlay */}
      <div className="absolute inset-0 bg-primary/5"></div>

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20">
          {/* Left side - Brand and Navigation */}
          <div className="flex items-center gap-12">
            {/* Brand */}
            <Link to="/dashboard" className="group relative">
              <h1 className="text-2xl font-bold text-primary tracking-tight transition-all duration-300 group-hover:scale-[1.02]">
                Serra
              </h1>
              <div className="absolute -inset-2 bg-primary/20 rounded-lg blur opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            </Link>

            {/* Navigation */}
            <nav className="flex gap-1 p-1 rounded-xl bg-card/40 backdrop-blur-sm border border-border/50">
              <Link
                to="/dashboard"
                className={`group relative px-6 py-3 rounded-lg font-semibold text-sm transition-all duration-300 ${
                  currentPath === "/dashboard" || currentPath === "/"
                    ? "bg-primary text-primary-foreground shadow-lg"
                    : "text-foreground hover:text-foreground hover:bg-muted/60"
                }`}
              >
                <span className="relative z-10">Dashboard</span>
                {(currentPath === "/dashboard" || currentPath === "/") && (
                  <div className="absolute inset-0 bg-primary rounded-lg blur-sm opacity-50"></div>
                )}
              </Link>
              <Link
                to="/requests"
                className={`group relative px-6 py-3 rounded-lg font-semibold text-sm transition-all duration-300 ${
                  currentPath === "/requests"
                    ? "bg-primary text-primary-foreground shadow-lg"
                    : "text-foreground hover:text-foreground hover:bg-muted/60"
                }`}
              >
                <span className="relative z-10">Requests</span>
                {currentPath === "/requests" && (
                  <div className="absolute inset-0 bg-primary rounded-lg blur-sm opacity-50"></div>
                )}
              </Link>
            </nav>
          </div>

          {/* Right side - Actions */}
          <div className="flex items-center gap-4">
            {/* Search Button */}
            <button className="group p-2 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-lg transition-all duration-300 hover:scale-110">
              <Search className="w-5 h-5" />
            </button>

            {/* Notifications */}
            <button className="group relative p-2 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-lg transition-all duration-300 hover:scale-110">
              <Bell className="w-5 h-5" />
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-destructive rounded-full animate-pulse shadow-md"></span>
            </button>

            {/* User Avatar */}
            <div className="flex items-center gap-3 px-4 py-2 rounded-xl bg-card/40 backdrop-blur-sm border border-border/50 transition-all duration-300 hover:bg-card/60">
              <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center shadow-md">
                <User className="w-4 h-4 text-primary-foreground" />
              </div>
              <div className="hidden sm:block">
                <p className="text-sm font-semibold text-foreground">
                  Welcome back
                </p>
                <p className="text-xs text-muted-foreground">
                  Ready to innovate
                </p>
              </div>
            </div>

            {/* Logout Button */}
            <button
              onClick={onLogout}
              className="group flex items-center gap-2 px-4 py-2 rounded-lg border border-border text-foreground hover:text-destructive-foreground hover:bg-destructive hover:border-transparent transition-all duration-300 font-semibold backdrop-blur-sm bg-card/20"
            >
              <LogOut className="w-4 h-4 group-hover:rotate-12 transition-transform duration-300" />
              <span>Sign Out</span>
            </button>
          </div>
        </div>
      </div>

      {/* Animated bottom border */}
      <div className="absolute bottom-0 left-0 w-full h-px bg-primary/50"></div>
    </header>
  );
}
