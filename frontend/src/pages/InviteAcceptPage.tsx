import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { 
  Mail, 
  Eye, 
  EyeOff, 
  CheckCircle, 
  XCircle, 
  Clock, 
  Server,
  Shield,
  AlertTriangle,
  Loader2
} from "lucide-react";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";
import type { AcceptInvitationRequest } from "@/types";
import { formatDistanceToNow } from "date-fns";

// Password validation function matching backend requirements
const validatePasswordStrength = (password: string): string | null => {
  if (password.length < 8) {
    return "Password must be at least 8 characters long";
  }
  
  if (password.length > 128) {
    return "Password must be no more than 128 characters long";
  }
  
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasNumber = /\d/.test(password);
  const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(password);
  
  if (!hasUpper) {
    return "Password must contain at least one uppercase letter";
  }
  
  if (!hasLower) {
    return "Password must contain at least one lowercase letter";  
  }
  
  if (!hasNumber) {
    return "Password must contain at least one number";
  }
  
  if (!hasSpecial) {
    return "Password must contain at least one special character";
  }
  
  // Check for common weak patterns
  const lowercasePassword = password.toLowerCase();
  const weakPatterns = ["password", "123456", "qwerty", "admin", "letmein", "welcome", "monkey", "dragon", "master", "login"];
  
  for (const pattern of weakPatterns) {
    if (lowercasePassword.includes(pattern)) {
      return "Password contains common weak patterns";
    }
  }
  
  // Check for repetitive characters
  if (/(.)\1{3,}/.test(password)) {
    return "Password cannot contain repetitive characters";
  }
  
  // Check for sequential patterns
  if (containsSequentialPattern(password)) {
    return "Password cannot contain sequential patterns";
  }
  
  return null;
};

// Helper function to check for sequential patterns
const containsSequentialPattern = (password: string): boolean => {
  // Check for ascending sequences (at least 4 chars)
  for (let i = 0; i <= password.length - 4; i++) {
    if (isSequential(password.slice(i, i + 4), true)) {
      return true;
    }
  }
  
  // Check for descending sequences (at least 4 chars)
  for (let i = 0; i <= password.length - 4; i++) {
    if (isSequential(password.slice(i, i + 4), false)) {
      return true;
    }
  }
  
  return false;
};

// Helper function to check if string is sequential
const isSequential = (s: string, ascending: boolean): boolean => {
  for (let i = 1; i < s.length; i++) {
    const diff = s.charCodeAt(i) - s.charCodeAt(i - 1);
    if (ascending ? diff !== 1 : diff !== -1) {
      return false;
    }
  }
  return true;
};

export default function InviteAcceptPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    password: "",
    confirm_password: "",
  });

  // Fetch invitation details
  const {
    data: invitation,
    isLoading: invitationLoading,
    error: invitationError,
  } = useQuery({
    queryKey: ["invitation", token],
    queryFn: () => invitationsApi.getInvitationByToken(token!),
    enabled: !!token,
    retry: false,
  });

  // Accept invitation mutation
  const acceptInvitationMutation = useMutation({
    mutationFn: invitationsApi.acceptInvitation,
    onSuccess: () => {
      toast.success("Welcome to Serra!", {
        description: "Your account has been created successfully. You can now log in.",
      });
      // Redirect to login page after a brief delay
      setTimeout(() => {
        navigate("/login");
      }, 2000);
    },
    onError: (error: unknown) => {
      const errorMessage = error && typeof error === 'object' && 'response' in error 
        ? (error.response as { data?: { error?: { message?: string } } })?.data?.error?.message 
        : "Please check your information and try again.";
      
      toast.error("Failed to accept invitation", {
        description: errorMessage,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.password || !formData.confirm_password) {
      toast.error("Please fill in all fields");
      return;
    }

    if (formData.password !== formData.confirm_password) {
      toast.error("Passwords do not match");
      return;
    }

    // Validate password strength
    const passwordError = validatePasswordStrength(formData.password);
    if (passwordError) {
      toast.error(passwordError);
      return;
    }

    if (!token) {
      toast.error("Invalid invitation token");
      return;
    }

    const request: AcceptInvitationRequest = {
      token,
      password: formData.password,
      confirm_password: formData.confirm_password,
    };

    acceptInvitationMutation.mutate(request);
  };

  const getExpirationStatus = (expiresAt: string) => {
    const expirationDate = new Date(expiresAt);
    const now = new Date();
    const isExpired = expirationDate < now;
    
    return {
      isExpired,
      timeUntilExpiry: formatDistanceToNow(expirationDate, { addSuffix: true }),
    };
  };

  // Loading state
  if (invitationLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="flex flex-col items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600 mb-4" />
            <p className="text-gray-600">Loading invitation...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Error state
  if (invitationError || !invitation) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-red-100 flex items-center justify-center">
              <XCircle className="h-6 w-6 text-red-600" />
            </div>
            <CardTitle className="text-red-600">Invalid Invitation</CardTitle>
            <CardDescription>
              This invitation link is invalid, expired, or has already been used.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button 
              onClick={() => navigate("/login")} 
              variant="outline" 
              className="w-full"
            >
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const expirationStatus = getExpirationStatus(invitation.expires_at);

  // Expired invitation
  if (expirationStatus.isExpired) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-orange-100 flex items-center justify-center">
              <Clock className="h-6 w-6 text-orange-600" />
            </div>
            <CardTitle className="text-orange-600">Invitation Expired</CardTitle>
            <CardDescription>
              This invitation expired {expirationStatus.timeUntilExpiry}. Please contact an administrator for a new invitation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button 
              onClick={() => navigate("/login")} 
              variant="outline" 
              className="w-full"
            >
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
            <Mail className="h-6 w-6 text-blue-600" />
          </div>
          <CardTitle className="text-2xl">Welcome to Serra</CardTitle>
          <CardDescription>
            You've been invited to join Serra. Complete your account setup below.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Invitation Details */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Email:</span>
              <span className="font-medium">{invitation.email}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Username:</span>
              <span className="font-medium">{invitation.username}</span>
            </div>
            {invitation.create_media_user && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Media Server:</span>
                <Badge variant="secondary" className="text-xs">
                  <Server className="mr-1 h-3 w-3" />
                  Account will be created
                </Badge>
              </div>
            )}
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Expires:</span>
              <span className="text-sm text-orange-600">
                {expirationStatus.timeUntilExpiry}
              </span>
            </div>
          </div>

          {/* Permissions */}
          {invitation.permissions && invitation.permissions.length > 0 && (
            <>
              <Separator />
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-gray-600" />
                  <span className="text-sm font-medium">Assigned Permissions</span>
                </div>
                <div className="flex flex-wrap gap-2">
                  {invitation.permissions.map((permission) => (
                    <Badge key={permission} variant="outline" className="text-xs">
                      {permission.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, l => l.toUpperCase())}
                    </Badge>
                  ))}
                </div>
              </div>
            </>
          )}

          <Separator />

          {/* Password Setup Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="Enter your password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  required
                  minLength={8}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <p className="text-xs text-gray-600">
                Password must be at least 8 characters long
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirm_password">Confirm Password</Label>
              <Input
                id="confirm_password"
                type="password"
                placeholder="Confirm your password"
                value={formData.confirm_password}
                onChange={(e) => setFormData({ ...formData, confirm_password: e.target.value })}
                required
                minLength={8}
              />
            </div>

            {invitation.create_media_user && (
              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription className="text-sm">
                  This password will be used for both your Serra account and your media server account.
                </AlertDescription>
              </Alert>
            )}

            <Button
              type="submit"
              className="w-full"
              disabled={acceptInvitationMutation.isPending}
            >
              {acceptInvitationMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Setting up account...
                </>
              ) : (
                <>
                  <CheckCircle className="mr-2 h-4 w-4" />
                  Complete Account Setup
                </>
              )}
            </Button>
          </form>

          <div className="text-center text-xs text-gray-600">
            By accepting this invitation, you agree to the terms of service.
          </div>
        </CardContent>
      </Card>
    </div>
  );
}