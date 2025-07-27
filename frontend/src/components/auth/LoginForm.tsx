import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useMutation } from "@tanstack/react-query";
import {
  User,
  Lock,
  Eye,
  EyeOff,
  Loader2,
  Film,
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

  const loginMutation = useMutation({
    mutationFn: () =>
      onLoginSuccess(credentials.username, credentials.password),
    onError: (error) => {
      setError(error instanceof Error ? error.message : "Login failed");
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
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
            <p className="text-gray-300 text-xl font-medium">Next-Gen Media Management</p>
            <div className="flex items-center justify-center mt-3 text-sm text-gray-400">
              <div className="w-2 h-2 bg-green-400 rounded-full mr-2 animate-pulse" />
              System Online • Secure Portal
            </div>
          </div>

          {/* Ultra-Modern Login Card */}
          <Card className="backdrop-blur-xl bg-white/5 border border-white/10 shadow-2xl shadow-black/20 rounded-3xl overflow-hidden">
            {/* Card Glow Effect */}
            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/10 via-purple-500/10 to-pink-500/10 rounded-3xl blur-xl" />
            
            <CardHeader className="relative text-center space-y-3 pb-8 pt-8">
              <CardTitle className="text-3xl font-bold text-white">Welcome Back</CardTitle>
              <CardDescription className="text-gray-300 text-lg">
                Enter the digital realm of your media universe
              </CardDescription>
            </CardHeader>
            
            <CardContent className="relative space-y-8 px-8 pb-8">
              <form onSubmit={handleSubmit} className="space-y-6">
                {/* Enhanced Username Field */}
                <div className="space-y-3">
                  <Label htmlFor="username" className="text-sm font-semibold text-gray-200 tracking-wide">
                    USERNAME
                  </Label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-xl blur opacity-0 group-focus-within:opacity-100 transition-opacity duration-500" />
                    <User className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400 group-focus-within:text-blue-400 transition-colors duration-300 z-10" />
                    <Input
                      id="username"
                      type="text"
                      placeholder="Enter your username"
                      className="relative pl-12 pr-4 py-4 h-14 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 backdrop-blur-sm focus:bg-white/15 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/25 transition-all duration-300"
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

                {/* Enhanced Password Field */}
                <div className="space-y-3">
                  <Label htmlFor="password" className="text-sm font-semibold text-gray-200 tracking-wide">
                    PASSWORD
                  </Label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-gradient-to-r from-purple-500/20 to-pink-500/20 rounded-xl blur opacity-0 group-focus-within:opacity-100 transition-opacity duration-500" />
                    <Lock className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400 group-focus-within:text-purple-400 transition-colors duration-300 z-10" />
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="Enter your password"
                      className="relative pl-12 pr-14 py-4 h-14 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 backdrop-blur-sm focus:bg-white/15 focus:border-purple-500/50 focus:ring-2 focus:ring-purple-500/25 transition-all duration-300"
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
                      className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white transition-colors duration-300 z-10 p-1 rounded-lg hover:bg-white/10"
                    >
                      {showPassword ? (
                        <EyeOff className="w-5 h-5" />
                      ) : (
                        <Eye className="w-5 h-5" />
                      )}
                    </button>
                  </div>
                </div>

                {/* Enhanced Error Message */}
                {error && (
                  <div className="relative overflow-hidden bg-red-500/20 border border-red-500/30 rounded-xl p-4 backdrop-blur-sm animate-in slide-in-from-top-3 duration-500">
                    <div className="absolute inset-0 bg-gradient-to-r from-red-500/10 to-pink-500/10 animate-pulse" />
                    <p className="relative text-red-300 text-sm font-semibold">{error}</p>
                  </div>
                )}

                {/* Epic Submit Button */}
                <div className="pt-4">
                  <Button
                    type="submit"
                    disabled={loginMutation.isPending}
                    className="relative w-full h-16 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 hover:from-blue-700 hover:via-purple-700 hover:to-pink-700 text-white font-bold text-lg rounded-xl shadow-2xl shadow-purple-500/25 border-0 transition-all duration-300 hover:scale-105 hover:shadow-purple-500/40 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none group overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 -skew-x-12 translate-x-[-100%] group-hover:translate-x-[200%] transition-transform duration-1000" />
                    {loginMutation.isPending ? (
                      <div className="relative flex items-center justify-center">
                        <Loader2 className="w-6 h-6 mr-3 animate-spin" />
                        <span className="text-lg">Authenticating...</span>
                      </div>
                    ) : (
                      <div className="relative flex items-center justify-center">
                        <span className="group-hover:scale-110 transition-transform duration-200">
                          Launch Dashboard
                        </span>
                        <div className="ml-3 opacity-70 group-hover:opacity-100 group-hover:translate-x-1 transition-all duration-200">
                          →
                        </div>
                      </div>
                    )}
                  </Button>
                </div>
              </form>

            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
