import { useState, useMemo, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useMutation, useQuery } from "@tanstack/react-query";
import { backendApi } from "@/lib/api";
import type { ServerInfo } from "@/types";
import {
  User,
  Lock,
  Eye,
  EyeOff,
  Loader2,
  Film,
  Server,
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
  const [loginMethod, setLoginMethod] = useState<'media-server' | 'local' | null>(null);

  // Get server authentication info
  const { data: serverInfo, isLoading: serverInfoLoading } = useQuery<ServerInfo>({
    queryKey: ['server-info'],
    queryFn: backendApi.getServerInfo,
    retry: false,
  });

  // Generate particle positions only once
  const particles = useMemo(() => 
    Array.from({ length: 30 }).map((_, i) => ({
      id: i,
      left: Math.random() * 100,
      top: Math.random() * 100,
      delay: Math.random() * 15,
      duration: 10 + Math.random() * 20
    })), []
  );

  // Set default login method based on server configuration
  useEffect(() => {
    if (serverInfo && loginMethod === null) {
      // Default to media server if available, otherwise local
      if (serverInfo.media_server_auth_enabled) {
        setLoginMethod('media-server');
      } else if (serverInfo.local_auth_enabled) {
        setLoginMethod('local');
      }
    }
  }, [serverInfo, loginMethod]);

  const mediaServerLoginMutation = useMutation({
    mutationFn: async () => {
      await backendApi.loginMediaServer(credentials.username, credentials.password);
      return onLoginSuccess(credentials.username, credentials.password);
    },
    onError: (error) => {
      setError(error instanceof Error ? error.message : "Media server login failed");
    },
  });

  const localLoginMutation = useMutation({
    mutationFn: async () => {
      await backendApi.loginLocal(credentials.username, credentials.password);
      return onLoginSuccess(credentials.username, credentials.password);
    },
    onError: (error) => {
      setError(error instanceof Error ? error.message : "Local login failed");
    },
  });

  const handleMediaServerSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!credentials.username || !credentials.password) {
      setError("Please enter both username and password");
      return;
    }

    try {
      await mediaServerLoginMutation.mutateAsync();
    } catch (error) {
      console.error("Media server login error:", error);
    }
  };

  const handleLocalSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!credentials.username || !credentials.password) {
      setError("Please enter both username and password");
      return;
    }

    try {
      await localLoginMutation.mutateAsync();
    } catch (error) {
      console.error("Local login error:", error);
    }
  };

  const handleSubmit = loginMethod === 'media-server' ? handleMediaServerSubmit : handleLocalSubmit;
  
  const toggleLoginMethod = () => {
    setLoginMethod(loginMethod === 'media-server' ? 'local' : 'media-server');
    setError(""); // Clear any errors when switching
    setCredentials({ username: "", password: "" }); // Clear form
  };

  const isLoading = mediaServerLoginMutation.isPending || localLoginMutation.isPending;

  return (
    <div className="min-h-screen relative overflow-hidden">
      {/* Elegant Modern Background */}
      <div className="absolute inset-0">
        {/* Dynamic Gradient Base */}
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-950 via-slate-900 to-violet-950" />
        
        {/* Animated Mesh Gradient */}
        <div className="absolute inset-0">
          <div className="absolute top-0 left-0 w-full h-full bg-gradient-to-br from-blue-600/20 via-transparent to-transparent animate-pulse" />
          <div className="absolute bottom-0 right-0 w-full h-full bg-gradient-to-tl from-purple-600/20 via-transparent to-transparent animate-pulse animation-delay-2000" />
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 rounded-full blur-3xl animate-pulse animation-delay-1000" />
        </div>

        {/* Flowing Particles */}
        <div className="absolute inset-0 overflow-hidden">
          {particles.map((particle) => (
            <div
              key={particle.id}
              className="absolute w-1 h-1 bg-white/30 rounded-full animate-float"
              style={{
                left: `${particle.left}%`,
                top: `${particle.top}%`,
                animationDelay: `${particle.delay}s`,
                animationDuration: `${particle.duration}s`
              }}
            />
          ))}
        </div>

        {/* Geometric Overlay */}
        <div className="absolute inset-0 opacity-30">
          <svg className="w-full h-full" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#6366f1" stopOpacity="0.1" />
                <stop offset="50%" stopColor="#8b5cf6" stopOpacity="0.2" />
                <stop offset="100%" stopColor="#a855f7" stopOpacity="0.1" />
              </linearGradient>
            </defs>
            
            {/* Flowing Lines */}
            <path
              d="M0,100 Q300,50 600,200 T1200,150"
              stroke="url(#lineGrad)"
              strokeWidth="2"
              fill="none"
              className="animate-pulse"
            />
            <path
              d="M0,300 Q400,200 800,400 T1600,300"
              stroke="url(#lineGrad)"
              strokeWidth="1.5"
              fill="none"
              className="animate-pulse"
              style={{ animationDelay: '2s' }}
            />
            <path
              d="M0,500 Q500,350 1000,600 T2000,500"
              stroke="url(#lineGrad)"
              strokeWidth="1"
              fill="none"
              className="animate-pulse"
              style={{ animationDelay: '1s' }}
            />
          </svg>
        </div>

        {/* Soft Grid Pattern */}
        <div className="absolute inset-0 opacity-5">
          <div className="absolute inset-0 bg-[linear-gradient(to_right,#ffffff10_1px,transparent_1px),linear-gradient(to_bottom,#ffffff10_1px,transparent_1px)] bg-[size:3rem_3rem]" />
        </div>

        {/* Radial Spotlight */}
        <div className="absolute inset-0 bg-gradient-radial from-transparent via-transparent to-black/40" />
        
        {/* Corner Fade */}
        <div className="absolute inset-0 bg-gradient-to-br from-black/20 via-transparent to-black/20" />
      </div>

      {/* CSS Animations */}
      <style jsx>{`
        @keyframes float {
          0%, 100% { 
            transform: translateY(0px) translateX(0px) rotate(0deg); 
            opacity: 0.3;
          }
          25% { 
            transform: translateY(-20px) translateX(10px) rotate(90deg); 
            opacity: 0.6;
          }
          50% { 
            transform: translateY(-10px) translateX(-5px) rotate(180deg); 
            opacity: 0.2;
          }
          75% { 
            transform: translateY(-30px) translateX(15px) rotate(270deg); 
            opacity: 0.5;
          }
        }
        .animate-float {
          animation: float 15s infinite linear;
        }
        .animation-delay-1000 {
          animation-delay: 1s;
        }
        .animation-delay-2000 {
          animation-delay: 2s;
        }
      `}</style>

      {/* Content */}
      <div className="relative z-10 min-h-screen flex items-center justify-center p-4">
        <div className="w-full max-w-lg">
          {/* Enhanced Logo/Brand Area */}
          <div className="text-center mb-12">
            <h1 className="text-6xl font-bold bg-gradient-to-r from-white via-blue-200 to-purple-200 bg-clip-text text-transparent mb-4 drop-shadow-lg">
              Serra
            </h1>
          </div>

          {/* Ultra-Modern Login Card */}
          <Card className="backdrop-blur-xl bg-white/5 border border-white/10 shadow-2xl shadow-black/20 rounded-3xl overflow-hidden">
            {/* Card Glow Effect */}
            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/10 via-purple-500/10 to-pink-500/10 rounded-3xl blur-xl" />
            
            <CardHeader className="relative text-center space-y-3 pb-8 pt-8">
              <CardTitle className="text-3xl font-bold text-white">Welcome Back</CardTitle>

            </CardHeader>
            
            <CardContent className="relative px-8 pb-8">
              {!serverInfoLoading && serverInfo && loginMethod && (
                <div className="space-y-6">
                  {/* Login Method Header with Toggle */}
                  <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                      {loginMethod === 'media-server' ? (
                        <>
                          <Server className="w-6 h-6 text-primary" />
                          <div>
                            <h3 className="text-xl font-semibold text-white">Sign in with {serverInfo.media_server_name}</h3>
                            <p className="text-sm text-gray-300">Use your {serverInfo.media_server_name} account credentials</p>
                          </div>
                        </>
                      ) : (
                        <>
                          <Shield className="w-6 h-6 text-emerald-500" />
                          <div>
                            <h3 className="text-xl font-semibold text-white">Sign in with Serra</h3>
                            <p className="text-sm text-gray-300">Use your local Serra account credentials</p>
                          </div>
                        </>
                      )}
                    </div>
                    
                    {/* Toggle Button - only show if both methods are enabled */}
                    {(serverInfo.media_server_auth_enabled && serverInfo.local_auth_enabled) && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={toggleLoginMethod}
                        className="bg-white/10 border-white/20 text-white hover:bg-white/15 hover:border-white/30 transition-all duration-300"
                      >
                        {loginMethod === 'media-server' ? 'Use Serra Account' : `Use ${serverInfo.media_server_name}`}
                      </Button>
                    )}
                  </div>
                  
                  {/* Single Login Form */}
                  <form onSubmit={handleSubmit} className="space-y-4">
                    {/* Username Field */}
                    <div className="space-y-2">
                      <Label htmlFor="username" className="text-sm font-medium text-gray-200">
                        {loginMethod === 'media-server' ? `${serverInfo.media_server_name} Username` : 'Serra Username'}
                      </Label>
                      <div className="relative">
                        <User className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                        <Input
                          id="username"
                          type="text"
                          placeholder={loginMethod === 'media-server' 
                            ? `Enter your ${serverInfo.media_server_name} username`
                            : 'Enter your Serra username'
                          }
                          className={`pl-10 h-12 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:bg-white/15 transition-all duration-300 ${
                            loginMethod === 'media-server' 
                              ? 'focus:border-primary/50' 
                              : 'focus:border-emerald-500/50'
                          }`}
                          value={credentials.username}
                          onChange={(e) => {
                            setCredentials((prev) => ({
                              ...prev,
                              username: e.target.value,
                            }));
                            setError("");
                          }}
                          required
                        />
                      </div>
                    </div>

                    {/* Password Field */}
                    <div className="space-y-2">
                      <Label htmlFor="password" className="text-sm font-medium text-gray-200">
                        {loginMethod === 'media-server' ? `${serverInfo.media_server_name} Password` : 'Serra Password'}
                      </Label>
                      <div className="relative">
                        <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                        <Input
                          id="password"
                          type={showPassword ? "text" : "password"}
                          placeholder={loginMethod === 'media-server' 
                            ? `Enter your ${serverInfo.media_server_name} password`
                            : 'Enter your Serra password'
                          }
                          className={`pl-10 pr-12 h-12 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:bg-white/15 transition-all duration-300 ${
                            loginMethod === 'media-server' 
                              ? 'focus:border-primary/50' 
                              : 'focus:border-emerald-500/50'
                          }`}
                          value={credentials.password}
                          onChange={(e) => {
                            setCredentials((prev) => ({
                              ...prev,
                              password: e.target.value,
                            }));
                            setError("");
                          }}
                          required
                        />
                        <button
                          type="button"
                          onClick={() => setShowPassword(!showPassword)}
                          className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white transition-colors duration-300"
                        >
                          {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                        </button>
                      </div>
                    </div>

                    {/* Submit Button */}
                    <Button
                      type="submit"
                      disabled={isLoading}
                      className={`w-full h-12 font-medium rounded-lg transition-all duration-300 hover:scale-[1.02] disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none ${
                        loginMethod === 'media-server'
                          ? 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700'
                          : 'bg-gradient-to-r from-emerald-600 to-green-600 hover:from-emerald-700 hover:to-green-700'
                      } text-white`}
                    >
                      {isLoading ? (
                        <div className="flex items-center justify-center">
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                          {loginMethod === 'media-server' 
                            ? `Connecting to ${serverInfo.media_server_name}...`
                            : 'Signing in...'
                          }
                        </div>
                      ) : (
                        `Sign in with ${loginMethod === 'media-server' ? serverInfo.media_server_name : 'Serra'}`
                      )}
                    </Button>
                  </form>
                  
                  {/* Error Message */}
                  {error && (
                    <div className="bg-red-500/20 border border-red-500/30 rounded-lg p-3">
                      <p className="text-red-300 text-sm font-medium">{error}</p>
                    </div>
                  )}
                </div>
              )}

            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
