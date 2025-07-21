import { useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { backendApi } from "@/lib/api";
import { type UserWithPermissions } from "@/types";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { Avatar } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useState, useEffect } from "react";

export default function UserSettingsPage() {
  const { userId } = useParams<{ userId: string }>();
  const queryClient = useQueryClient();
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [passwordData, setPasswordData] = useState({
    newPassword: "",
    confirmPassword: "",
  });

  // Get current user to check if they're viewing their own profile
  const {
    data: currentUser,
  } = useQuery({
    queryKey: ["me"],
    queryFn: backendApi.getCurrentUser,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  });

  const {
    data: user,
    isLoading: userLoading,
    error: userError,
    refetch: refetchUser,
  } = useQuery({
    queryKey: ["user", userId],
    queryFn: async (): Promise<UserWithPermissions> => {
      return await backendApi.getUser(userId!);
    },
    enabled: !!userId,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  });

  const {
    data: allPermissions,
    isLoading: permissionsLoading,
    error: permissionsError,
    refetch: refetchPermissions,
  } = useQuery({
    queryKey: ["permissions"],
    queryFn: backendApi.getPermissions,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  });

  const handleRetry = () => {
    refetchUser();
    refetchPermissions();
  };

  const updatePermissionsMutation = useMutation({
    mutationFn: ({ userId, permissions }: { userId: string; permissions: string[] }) =>
      backendApi.updateUserPermissions(userId, permissions),
    onSuccess: () => {
      console.log("Permissions updated successfully");
      queryClient.invalidateQueries({ queryKey: ["user", userId] });
    },
    onError: (error: any) => {
      console.error("Failed to update permissions:", error);
    },
  });

  const changePasswordMutation = useMutation({
    mutationFn: ({ userId, newPassword }: { userId: string; newPassword: string }) =>
      backendApi.changeUserPassword(userId, newPassword),
    onSuccess: () => {
      console.log("Password changed successfully");
      setPasswordData({ newPassword: "", confirmPassword: "" });
    },
    onError: (error: any) => {
      console.error("Failed to change password:", error);
    },
  });

  useEffect(() => {
    if (user) {
      setSelectedPermissions(user.permissions ? user.permissions.map(p => p.id) : []);
      console.log("User data loaded:", user);
      console.log("Avatar URL:", user.avatar_url);
    }
  }, [user]);

  const handlePermissionToggle = (permissionId: string, checked: boolean) => {
    setSelectedPermissions(prev => 
      checked 
        ? [...prev, permissionId]
        : prev.filter(id => id !== permissionId)
    );
  };

  const handleSavePermissions = () => {
    if (userId) {
      updatePermissionsMutation.mutate({ userId, permissions: selectedPermissions });
    }
  };

  const handleChangePassword = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (passwordData.newPassword !== passwordData.confirmPassword) {
      alert("Passwords do not match");
      return;
    }
    
    if (passwordData.newPassword.length < 6) {
      alert("Password must be at least 6 characters long");
      return;
    }

    if (userId) {
      changePasswordMutation.mutate({ 
        userId, 
        newPassword: passwordData.newPassword 
      });
    }
  };

  const hasPermissionsChanged = () => {
    if (!user) return false;
    const currentPermissionIds = user.permissions ? user.permissions.map(p => p.id).sort() : [];
    const selectedPermissionIds = [...selectedPermissions].sort();
    return JSON.stringify(currentPermissionIds) !== JSON.stringify(selectedPermissionIds);
  };

  // Check if current user is viewing their own profile and has owner permission
  const isViewingOwnProfile = currentUser && userId === currentUser.id;
  const hasOwnerPermission = user?.permissions?.some(p => p.id === "owner");
  const isOwnerViewingOwnProfile = isViewingOwnProfile && hasOwnerPermission;
  
  // Check if password change is allowed for this user
  const isLocalUser = user?.user_type === "local";
  const canChangePassword = isLocalUser && (isViewingOwnProfile || currentUser?.is_admin);

  const isLoading = userLoading || permissionsLoading;
  const error = userError || permissionsError;

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">User Settings</h1>
          <p className="text-muted-foreground">
            Configure user account and permissions
          </p>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="text-muted-foreground">Loading user...</div>
        </div>
      </div>
    );
  }

  if (error) {
    const is403Error = error instanceof Error && error.message.includes('403');
    
    return (
      <div>
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">User Settings</h1>
          <p className="text-muted-foreground">
            Configure user account and permissions
          </p>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="text-red-500 text-center">
            {is403Error ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <p className="font-medium">Access Denied</p>
                  <p className="text-sm text-muted-foreground">
                    You don't have permission to manage users. Please contact an administrator.
                  </p>
                </div>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={handleRetry}
                  disabled={isLoading}
                >
                  {isLoading ? "Checking..." : "Retry"}
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <p className="font-medium">Error loading data</p>
                  <p className="text-sm text-muted-foreground">
                    {error instanceof Error ? error.message : 'Unknown error'}
                  </p>
                </div>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={handleRetry}
                  disabled={isLoading}
                >
                  {isLoading ? "Loading..." : "Retry"}
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="py-10">
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">User Settings</h1>
          <p className="text-muted-foreground">
            Configure user account and permissions
          </p>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="text-muted-foreground">User not found</div>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight">
          Settings for {user.username}
        </h1>
        <p className="text-muted-foreground">
          Configure user account and permissions
        </p>
      </div>

      <div className="space-y-6">
        {/* User Information Section */}
        <div className="bg-card p-6 rounded-lg border">
          <h2 className="text-xl font-semibold mb-4">User Information</h2>
          <div className="flex items-start gap-6">
            <div className="flex-shrink-0">
              <Avatar 
                src={user.avatar_url ? `http://localhost:9090/v1${user.avatar_url}` : undefined}
                alt={user.username}
                fallback={user.username.charAt(0)}
                size="xl"
                className="border-2 border-border"
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 flex-1">
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Username
                </label>
                <p className="text-sm font-mono">{user.username}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Email
                </label>
                <p className="text-sm font-mono">
                  {user.email || "No email set"}
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  User Type
                </label>
                <div className="flex items-center">
                  <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                    user.user_type === "local" 
                      ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300"
                      : "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300"
                  }`}>
                    {user.user_type === "local" ? "Local" : "Media Server"}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Permissions Section */}
        <div className="bg-card p-6 rounded-lg border">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Permissions</h2>
            {!isOwnerViewingOwnProfile && (
              <Button 
                onClick={handleSavePermissions}
                disabled={updatePermissionsMutation.isPending || !hasPermissionsChanged()}
              >
                {updatePermissionsMutation.isPending ? "Saving..." : "Save Changes"}
              </Button>
            )}
          </div>
          
          {isOwnerViewingOwnProfile && (
            <div className="mb-4 p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg">
              <p className="text-sm text-amber-800 dark:text-amber-200">
                <strong>Owner permissions cannot be modified.</strong> As the system owner, your permissions are protected and cannot be changed to prevent accidental lockout.
              </p>
            </div>
          )}
          
          <div className="space-y-6">
            {allPermissions?.permissions?.length > 0 ? (() => {
              // Group permissions by category
              const permissionsByCategory = allPermissions.permissions.reduce((acc: any, permission: any) => {
                const category = permission.category || 'Other';
                if (!acc[category]) {
                  acc[category] = [];
                }
                acc[category].push(permission);
                return acc;
              }, {});

              // Define category order for better UX
              const categoryOrder = [
                'Owner',
                'Administrative', 
                'Request Content',
                'Auto-Approve Requests',
                'Manage Requests',
                'Other'
              ];

              return categoryOrder.map(category => {
                if (!permissionsByCategory[category] || permissionsByCategory[category].length === 0) {
                  return null;
                }

                return (
                  <div key={category} className="space-y-3">
                    <h3 className="text-sm font-semibold text-foreground border-b pb-2">
                      {category}
                    </h3>
                    <div className="space-y-2 pl-2">
                      {permissionsByCategory[category].map((permission: any) => (
                        <div key={permission.id} className="flex items-center space-x-3 p-3 rounded-lg border bg-background/50">
                          <Checkbox
                            id={`permission-${permission.id}`}
                            checked={selectedPermissions.includes(permission.id)}
                            disabled={isOwnerViewingOwnProfile}
                            onCheckedChange={(checked) => 
                              handlePermissionToggle(permission.id, !!checked)
                            }
                          />
                          <div className="flex-1">
                            <label 
                              htmlFor={`permission-${permission.id}`}
                              className="text-sm font-medium cursor-pointer"
                            >
                              {permission.name}
                            </label>
                            {permission.description && (
                              <p className="text-xs text-muted-foreground mt-1">
                                {permission.description}
                              </p>
                            )}
                            {permission.dangerous && (
                              <span className="inline-flex items-center px-2 py-1 mt-1 text-xs font-medium text-red-700 bg-red-100 rounded-full dark:bg-red-900 dark:text-red-300">
                                Dangerous
                              </span>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                );
              }).filter(Boolean);
            })() : (
              <p className="text-sm text-muted-foreground">
                No permissions available
              </p>
            )}
          </div>
        </div>

        {/* Password Change Section - Only for local users */}
        {canChangePassword && (
          <div className="bg-card p-6 rounded-lg border">
            <h2 className="text-xl font-semibold mb-4">Change Password</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Update the password for this local user account.
            </p>
            
            <form onSubmit={handleChangePassword} className="space-y-4 max-w-md">
              <div className="space-y-2">
                <Label htmlFor="newPassword">New Password</Label>
                <Input
                  id="newPassword"
                  type="password"
                  value={passwordData.newPassword}
                  onChange={(e) => setPasswordData(prev => ({ 
                    ...prev, 
                    newPassword: e.target.value 
                  }))}
                  placeholder="Enter new password"
                  required
                  minLength={6}
                />
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm Password</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={passwordData.confirmPassword}
                  onChange={(e) => setPasswordData(prev => ({ 
                    ...prev, 
                    confirmPassword: e.target.value 
                  }))}
                  placeholder="Confirm new password"
                  required
                  minLength={6}
                />
              </div>
              
              <Button 
                type="submit" 
                disabled={
                  changePasswordMutation.isPending || 
                  !passwordData.newPassword || 
                  !passwordData.confirmPassword ||
                  passwordData.newPassword !== passwordData.confirmPassword
                }
              >
                {changePasswordMutation.isPending ? "Changing..." : "Change Password"}
              </Button>
              
              {passwordData.newPassword && passwordData.confirmPassword && 
               passwordData.newPassword !== passwordData.confirmPassword && (
                <p className="text-sm text-red-500">Passwords do not match</p>
              )}
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
