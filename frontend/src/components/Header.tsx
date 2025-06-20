import { Button } from "@/components/ui/button";
import { LogOut, Bell, Search, User } from "lucide-react";
import { Link, useLocation } from "react-router-dom";

interface HeaderProps {
  onLogout: () => void;
}

export function Header({ onLogout }: HeaderProps) {
  const location = useLocation();
  const currentPath = location.pathname;

  return (
    <header className="sticky top-0 z-50 backdrop-blur-xl bg-gradient-to-r from-slate-900/95 via-gray-900/95 to-slate-900/95 border-b border-white/10 shadow-2xl">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20">
          {/* Left side - Brand and Navigation */}
          <div className="flex items-center space-x-12">
            {/* Brand */}
            <Link to="/dashboard" className="flex items-center group">
              <div className="relative">
                <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent tracking-tight">
                  Serra
                </h1>
                <div className="absolute -inset-1 bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 rounded-lg blur opacity-20 group-hover:opacity-40 transition duration-300"></div>
              </div>
            </Link>

            {/* Navigation */}
            <nav className="flex space-x-1 bg-white/5 rounded-xl p-1 backdrop-blur-sm">
              <Link
                to="/dashboard"
                className={`relative px-6 py-3 rounded-lg font-medium text-sm transition-all duration-300 ${
                  currentPath === "/dashboard"
                    ? "bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg shadow-blue-500/25 transform scale-105"
                    : "text-gray-300 hover:text-white hover:bg-white/10"
                }`}
              >
                <span className="relative z-10">Dashboard</span>
                {currentPath === "/dashboard" && (
                  <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg blur-sm opacity-50"></div>
                )}
              </Link>
              <Link
                to="/requests"
                className={`relative px-6 py-3 rounded-lg font-medium text-sm transition-all duration-300 ${
                  currentPath === "/requests"
                    ? "bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg shadow-blue-500/25 transform scale-105"
                    : "text-gray-300 hover:text-white hover:bg-white/10"
                }`}
              >
                <span className="relative z-10">Requests</span>
                {currentPath === "/requests" && (
                  <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg blur-sm opacity-50"></div>
                )}
              </Link>
            </nav>
          </div>

          {/* Right side - Actions */}
          <div className="flex items-center space-x-4">
            {/* Search Button */}
            <button className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all duration-200 hover:scale-110">
              <Search className="w-5 h-5" />
            </button>

            {/* Notifications */}
            <button className="relative p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all duration-200 hover:scale-110">
              <Bell className="w-5 h-5" />
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-gradient-to-r from-pink-500 to-red-500 rounded-full animate-pulse"></span>
            </button>

            {/* User Avatar */}
            <div className="flex items-center space-x-3 bg-white/5 rounded-xl px-4 py-2 backdrop-blur-sm">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <User className="w-4 h-4 text-white" />
              </div>
              <div className="hidden sm:block">
                <p className="text-sm font-medium text-white">Welcome back</p>
                <p className="text-xs text-gray-400">Ready to innovate</p>
              </div>
            </div>

            {/* Logout Button */}
            <Button
              variant="outline"
              onClick={onLogout}
              className="border-white/20 text-gray-300 hover:text-white hover:bg-red-500/20 hover:border-red-500/50 transition-all duration-300 group"
            >
              <LogOut className="w-4 h-4 mr-2 group-hover:rotate-12 transition-transform duration-200" />
              Sign Out
            </Button>
          </div>
        </div>
      </div>

      {/* Animated bottom border */}
      <div className="absolute bottom-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent opacity-50"></div>
    </header>
  );
}
