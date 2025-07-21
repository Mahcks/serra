import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import {
  User,
  Lock,
  Eye,
  EyeOff,
  Loader2,
  Sparkles,
  Shield,
} from "lucide-react";

interface LoginFormProps {
  onLoginSuccess: (username: string, password: string) => Promise<void>;
}

export function LoginForm({ onLoginSuccess }: LoginFormProps) {
  const [credentials, setCredentials] = useState({
    username: "",
    password: "",
  });
  const [error, setError] = useState<string>("");
  const [showPassword, setShowPassword] = useState(false);
  const [focusedField, setFocusedField] = useState<string | null>(null);

  const loginMutation = useMutation({
    mutationFn: () =>
      onLoginSuccess(credentials.username, credentials.password),
    onError: (error) => {
      setError(error instanceof Error ? error.message : "Login failed");
    },
  });

  const handleSubmit = async () => {
    setError("");

    if (!credentials.username || !credentials.password) {
      setError("Please enter both username and password");
      return;
    }

    try {
      await loginMutation.mutateAsync();
    } catch (error) {
      console.error("Login error:", error);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSubmit();
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 p-4">
      {/* Background Effects */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse"></div>
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse delay-1000"></div>
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-80 h-80 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl opacity-10 animate-pulse delay-500"></div>
      </div>

      <div className="relative w-full max-w-md">
        {/* Main Card */}
        <div className="backdrop-blur-xl bg-white/10 border border-white/20 rounded-3xl p-8 shadow-2xl">
          {/* Header */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl mb-4 shadow-lg shadow-blue-500/25">
              <Shield className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-white via-blue-100 to-purple-100 bg-clip-text text-transparent mb-2">
              Welcome to Serra
            </h1>
            <p className="text-gray-300 text-lg">
              Sign in to access your media server
            </p>
            <div className="flex items-center justify-center mt-2 text-sm text-gray-400">
              <Sparkles className="w-4 h-4 mr-1" />
              Secure & Modern Dashboard
            </div>
          </div>

          <div className="space-y-6">
            {/* Username Field */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-200 mb-2 ml-1">
                Username
              </label>
              <div className="relative group">
                <div
                  className={`absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-xl blur transition-all duration-300 ${
                    focusedField === "username"
                      ? "opacity-100 scale-105"
                      : "opacity-0 scale-100"
                  }`}
                ></div>
                <div className="relative">
                  <User
                    className={`absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 transition-colors duration-200 ${
                      focusedField === "username"
                        ? "text-blue-400"
                        : "text-gray-400"
                    }`}
                  />
                  <input
                    type="text"
                    required
                    className="w-full pl-12 pr-4 py-4 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 backdrop-blur-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all duration-200"
                    placeholder="Enter your username"
                    value={credentials.username}
                    onFocus={() => setFocusedField("username")}
                    onBlur={() => setFocusedField(null)}
                    onKeyPress={handleKeyPress}
                    onChange={(e) => {
                      setCredentials((prev) => ({
                        ...prev,
                        username: e.target.value,
                      }));
                      setError("");
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Password Field */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-200 mb-2 ml-1">
                Password
              </label>
              <div className="relative group">
                <div
                  className={`absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-xl blur transition-all duration-300 ${
                    focusedField === "password"
                      ? "opacity-100 scale-105"
                      : "opacity-0 scale-100"
                  }`}
                ></div>
                <div className="relative">
                  <Lock
                    className={`absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 transition-colors duration-200 ${
                      focusedField === "password"
                        ? "text-blue-400"
                        : "text-gray-400"
                    }`}
                  />
                  <input
                    type={showPassword ? "text" : "password"}
                    required
                    className="w-full pl-12 pr-12 py-4 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 backdrop-blur-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all duration-200"
                    placeholder="Enter your password"
                    value={credentials.password}
                    onFocus={() => setFocusedField("password")}
                    onBlur={() => setFocusedField(null)}
                    onKeyPress={handleKeyPress}
                    onChange={(e) => {
                      setCredentials((prev) => ({
                        ...prev,
                        password: e.target.value,
                      }));
                      setError("");
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white transition-colors duration-200"
                  >
                    {showPassword ? (
                      <EyeOff className="w-5 h-5" />
                    ) : (
                      <Eye className="w-5 h-5" />
                    )}
                  </button>
                </div>
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <div className="bg-red-500/20 border border-red-500/30 rounded-xl p-4 backdrop-blur-sm">
                <p className="text-red-300 text-sm font-medium">{error}</p>
              </div>
            )}

            {/* Submit Button */}
            <Button
              onClick={handleSubmit}
              disabled={loginMutation.isPending}
              className="w-full h-14 bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-semibold rounded-xl shadow-lg shadow-blue-500/25 border-0 transition-all duration-300 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none group"
            >
              {loginMutation.isPending ? (
                <div className="flex items-center">
                  <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                  Signing in...
                </div>
              ) : (
                <div className="flex items-center justify-center">
                  <span className="group-hover:scale-110 transition-transform duration-200">
                    Sign In
                  </span>
                  <div className="ml-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                    →
                  </div>
                </div>
              )}
            </Button>
          </div>

          {/* Footer */}
          <div className="mt-8 text-center">
            <p className="text-gray-400 text-sm">
              Powered by modern authentication
            </p>
          </div>
        </div>

        {/* Decorative Elements */}
        <div className="absolute -top-6 -left-6 w-12 h-12 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full opacity-20 animate-bounce delay-300"></div>
        <div className="absolute -bottom-6 -right-6 w-8 h-8 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full opacity-20 animate-bounce delay-700"></div>
      </div>
    </div>
  );
}
