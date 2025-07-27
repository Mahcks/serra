import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Avatar } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { 
  Users,
  UserPlus,
  Mail, 
  Clock, 
  CheckCircle, 
  XCircle, 
  AlertCircle,
  RefreshCw,
  MoreHorizontal,
  Edit,
  Shield,
  Crown,
  Server,
  User,
  Save,
  Loader2
} from "lucide-react";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";
import { backendApi } from "@/lib/api";
import { CreateInvitationDialog } from "@/components/admin/invitations/CreateInvitationDialog";
import { InvitationDataTable } from "@/components/admin/invitations/invitation-data-table";
import { createInvitationColumns } from "@/components/admin/invitations/invitation-columns";
import UserPermissionSelector from "@/components/admin/UserPermissionSelector";
import { DataTable } from "@/pages/user/data-table";
import { type ColumnDef } from "@tanstack/react-table";
import type { UserWithPermissions } from "@/types";

export default function UsersAndInvitationsPage() {
  const queryClient = useQueryClient();
  const [editingUser, setEditingUser] = useState<UserWithPermissions | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(new Set());
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState<UserWithPermissions | null>(null);
  const [editUserForm, setEditUserForm] = useState({
    username: "",
    email: "",
  });
  const [createUserDialogOpen, setCreateUserDialogOpen] = useState(false);
  const [createUserForm, setCreateUserForm] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [createUserPermissions, setCreateUserPermissions] = useState<Set<string>>(new Set());

  // Fetch users
  const {
    data: usersResponse,
    isLoading: usersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useQuery({
    queryKey: ["users"],
    queryFn: backendApi.getUsers,
    retry: false,
  });

  const users = usersResponse?.users || [];

  // Fetch invitations
  const {
    data: invitations = [],
    isLoading: invitationsLoading,
    error: invitationsError,
    refetch: refetchInvitations,
  } = useQuery({
    queryKey: ["invitations"],
    queryFn: invitationsApi.getAllInvitations,
    retry: false,
  });

  // Fetch invitation statistics
  const {
    data: stats,
    isLoading: statsLoading,
  } = useQuery({
    queryKey: ["invitation-stats"],
    queryFn: invitationsApi.getInvitationStats,
    retry: false,
    staleTime: 30000, // 30 seconds
  });

  // Cancel invitation mutation
  const cancelInvitationMutation = useMutation({
    mutationFn: invitationsApi.cancelInvitation,
    onSuccess: () => {
      toast.success("Invitation cancelled successfully");
      queryClient.invalidateQueries({ queryKey: ["invitations"] });
      queryClient.invalidateQueries({ queryKey: ["invitation-stats"] });
    },
    onError: (error: unknown) => {
      toast.error("Failed to cancel invitation", {
        description: (error as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message || "Please try again.",
      });
    },
  });

  // Delete invitation mutation
  const deleteInvitationMutation = useMutation({
    mutationFn: invitationsApi.deleteInvitation,
    onSuccess: () => {
      toast.success("Invitation deleted successfully");
      queryClient.invalidateQueries({ queryKey: ["invitations"] });
      queryClient.invalidateQueries({ queryKey: ["invitation-stats"] });
    },
    onError: (error: unknown) => {
      toast.error("Failed to delete invitation", {
        description: (error as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message || "Please try again.",
      });
    },
  });

  const handleInvitationSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ["invitations"] });
    queryClient.invalidateQueries({ queryKey: ["invitation-stats"] });
  };

  const handleCancelInvitation = (invitationId: number) => {
    cancelInvitationMutation.mutate(invitationId);
  };

  const handleDeleteInvitation = (invitationId: number) => {
    deleteInvitationMutation.mutate(invitationId);
  };

  // Update user permissions mutation
  const updateUserPermissionsMutation = useMutation({
    mutationFn: async ({ userId, permissions }: { userId: string; permissions: string[] }) => {
      const response = await fetch(`/v1/users/${userId}/permissions`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ permissions }),
      });

      if (!response.ok) {
        throw new Error('Failed to update user permissions');
      }

      return response.json();
    },
    onSuccess: () => {
      toast.success("User permissions updated successfully");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setEditDialogOpen(false);
      setEditingUser(null);
    },
    onError: (error: unknown) => {
      toast.error("Failed to update user permissions", {
        description: (error as Error)?.message || "Please try again.",
      });
    },
  });

  // Update user profile mutation
  const updateUserProfileMutation = useMutation({
    mutationFn: async ({ userId, userData }: { userId: string; userData: { username: string; email: string } }) => {
      const response = await fetch(`/v1/users/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error?.message || 'Failed to update user profile');
      }

      return response.json();
    },
    onSuccess: () => {
      toast.success("User profile updated successfully");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setEditDialogOpen(false);
      setEditingUser(null);
    },
    onError: (error: unknown) => {
      toast.error("Failed to update user profile", {
        description: (error as Error)?.message || "Please try again.",
      });
    },
  });

  // Create local user mutation
  const createLocalUserMutation = useMutation({
    mutationFn: async (userData: { username: string; email?: string; password: string; permissions?: string[] }) => {
      const response = await backendApi.createLocalUser(userData);
      return response;
    },
    onSuccess: () => {
      toast.success("Local user created successfully");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setCreateUserDialogOpen(false);
      setCreateUserForm({
        username: "",
        email: "",
        password: "",
        confirmPassword: "",
      });
      setCreateUserPermissions(new Set());
    },
    onError: (error: unknown) => {
      toast.error("Failed to create local user", {
        description: (error as Error)?.message || "Please try again.",
      });
    },
  });

  // Delete user mutation
  const deleteUserMutation = useMutation({
    mutationFn: async (userId: string) => {
      return await backendApi.deleteUser(userId);
    },
    onSuccess: () => {
      toast.success("User deleted successfully");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setDeleteDialogOpen(false);
      setUserToDelete(null);
    },
    onError: (error: unknown) => {
      toast.error("Failed to delete user", {
        description: (error as Error)?.message || "Please try again.",
      });
    },
  });

  const handleEditUser = (user: UserWithPermissions) => {
    setEditingUser(user);
    setSelectedPermissions(new Set(user.permissions?.map(p => p.id) || []));
    setEditUserForm({
      username: user.username,
      email: user.email || "",
    });
    setEditDialogOpen(true);
  };

  const handleSaveUserProfile = () => {
    if (!editingUser) return;
    updateUserProfileMutation.mutate({
      userId: editingUser.id,
      userData: {
        username: editUserForm.username.trim(),
        email: editUserForm.email.trim(),
      },
    });
  };

  const handleSaveUserPermissions = () => {
    if (!editingUser) return;
    updateUserPermissionsMutation.mutate({
      userId: editingUser.id,
      permissions: Array.from(selectedPermissions),
    });
  };

  const handleSaveAll = async () => {
    if (!editingUser) return;
    
    // Save profile changes if they exist
    const hasProfileChanges = 
      editUserForm.username.trim() !== editingUser.username ||
      editUserForm.email.trim() !== (editingUser.email || "");
    
    if (hasProfileChanges) {
      updateUserProfileMutation.mutate({
        userId: editingUser.id,
        userData: {
          username: editUserForm.username.trim(),
          email: editUserForm.email.trim(),
        },
      });
    }

    // Save permission changes
    const currentPermissions = new Set(editingUser.permissions?.map(p => p.id) || []);
    const hasPermissionChanges = 
      selectedPermissions.size !== currentPermissions.size ||
      !Array.from(selectedPermissions).every(p => currentPermissions.has(p));

    if (hasPermissionChanges) {
      updateUserPermissionsMutation.mutate({
        userId: editingUser.id,
        permissions: Array.from(selectedPermissions),
      });
    }

    // Close dialog if no changes were made
    if (!hasProfileChanges && !hasPermissionChanges) {
      setEditDialogOpen(false);
      setEditingUser(null);
    }
  };

  const handlePermissionChange = (permissionId: string, checked: boolean) => {
    setSelectedPermissions(prev => {
      const newPermissions = new Set(prev);
      if (checked) {
        newPermissions.add(permissionId);
      } else {
        newPermissions.delete(permissionId);
      }
      return newPermissions;
    });
  };

  const handleCreateUserPermissionChange = (permissionId: string, checked: boolean) => {
    setCreateUserPermissions(prev => {
      const newPermissions = new Set(prev);
      if (checked) {
        newPermissions.add(permissionId);
      } else {
        newPermissions.delete(permissionId);
      }
      return newPermissions;
    });
  };

  const handleCreateLocalUser = (e: React.FormEvent) => {
    e.preventDefault();

    // Basic validation
    if (!createUserForm.username.trim()) {
      toast.error("Username is required");
      return;
    }

    if (!createUserForm.password.trim()) {
      toast.error("Password is required");
      return;
    }

    if (createUserForm.password !== createUserForm.confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }

    createLocalUserMutation.mutate({
      username: createUserForm.username.trim(),
      email: createUserForm.email.trim() || undefined,
      password: createUserForm.password,
      permissions: Array.from(createUserPermissions),
    });
  };

  const handleDeleteUser = (user: UserWithPermissions) => {
    setUserToDelete(user);
    setDeleteDialogOpen(true);
  };

  const confirmDeleteUser = () => {
    if (!userToDelete) return;
    deleteUserMutation.mutate(userToDelete.id);
  };

  // Create user columns with edit functionality
  const createUserColumns = (): ColumnDef<UserWithPermissions>[] => {
    return [
      {
        accessorKey: "username",
        header: "User",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <div className="flex items-center gap-3">
              <Avatar 
                src={user.avatar_url ? `v1${user.avatar_url}` : undefined}
                alt={user.username}
                fallback={user.username.charAt(0).toUpperCase()}
                size="sm"
                className="bg-muted"
              />
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{user.username}</span>
                  {user.permissions?.some(p => p.id === 'owner') && (
                    <Crown className="h-3 w-3 text-yellow-500" />
                  )}
                  {user.permissions?.some(p => p.category === 'admin') && !user.permissions?.some(p => p.id === 'owner') && (
                    <Shield className="h-3 w-3 text-blue-500" />
                  )}
                </div>
                <div className="text-xs text-muted-foreground">
                  {user.email || 'No email'}
                </div>
              </div>
            </div>
          );
        },
      },
      {
        accessorKey: "user_type",
        header: "Account Type",
        cell: ({ row }) => {
          const userType = row.getValue("user_type") as string;
          return (
            <Badge variant={userType === 'media_server' ? 'default' : 'secondary'} className="text-xs">
              {userType === 'media_server' ? (
                <>
                  <Server className="h-3 w-3 mr-1" />
                  Media
                </>
              ) : (
                <>
                  <User className="h-3 w-3 mr-1" />
                  Local
                </>
              )}
            </Badge>
          );
        },
      },
      {
        accessorKey: "permissions",
        header: "Permissions",
        cell: ({ row }) => {
          const permissions = row.getValue("permissions") as UserWithPermissions["permissions"];
          
          if (!permissions || permissions.length === 0) {
            return <span className="text-muted-foreground">No permissions</span>;
          }

          const count = permissions.length;
          const displayPermissions = permissions
            .slice(0, 2)
            .map((p) => p.name)
            .join(", ");

          return (
            <div className="flex items-center gap-2">
              <span className="text-sm">
                {displayPermissions}
                {count > 2 && (
                  <span className="text-muted-foreground"> +{count - 2} more</span>
                )}
              </span>
              <Badge variant="outline" className="text-xs">
                {count}
              </Badge>
            </div>
          );
        },
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => handleEditUser(user)}>
                  <Edit className="h-4 w-4 mr-2" />
                  Edit User
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={() => handleDeleteUser(user)}
                  className="text-red-600 focus:text-red-600"
                >
                  <XCircle className="h-4 w-4 mr-2" />
                  Delete User
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ];
  };

  const userColumns = createUserColumns();
  const invitationColumns = createInvitationColumns(handleCancelInvitation, handleDeleteInvitation);

  // Statistics cards data
  const statisticsCards = [
    {
      title: "Total Users",
      value: users.length,
      icon: Users,
      color: "text-blue-600",
    },
    {
      title: "Pending Invitations",
      value: stats?.pending_count || 0,
      icon: Clock,
      color: "text-yellow-600",
    },
    {
      title: "Recent Joins",
      value: stats?.accepted_count || 0,
      icon: CheckCircle,
      color: "text-green-600",
    },
    {
      title: "Expired/Cancelled",
      value: (stats?.expired_count || 0) + (stats?.cancelled_count || 0),
      icon: XCircle,
      color: "text-red-600",
    },
  ];

  const pendingInvitations = (invitations || []).filter(inv => inv.status === 'pending');

  if (usersError || invitationsError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <AlertCircle className="h-12 w-12 text-red-500" />
        <div className="text-center">
          <h3 className="text-lg font-semibold">Failed to load users data</h3>
          <p className="text-muted-foreground mt-1">
            There was an error loading the users and invitations data.
          </p>
          <Button 
            variant="outline" 
            onClick={() => {
              refetchUsers();
              refetchInvitations();
            }}
            className="mt-3"
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary/20 rounded-lg border">
            <Users className="w-6 h-6 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-foreground">User Management</h1>
            <p className="text-muted-foreground">Manage users and invitations. Create invitation links to share directly with new users.</p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          <Button 
            variant="outline" 
            onClick={() => {
              refetchUsers();
              refetchInvitations();
            }}
            disabled={usersLoading || invitationsLoading}
            className="min-w-[100px]"
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${(usersLoading || invitationsLoading) ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Sheet open={createUserDialogOpen} onOpenChange={setCreateUserDialogOpen}>
            <SheetTrigger asChild>
              <Button variant="outline">
                <UserPlus className="mr-2 h-4 w-4" />
                Create Local User
              </Button>
            </SheetTrigger>
            <SheetContent className="w-[800px] sm:max-w-[800px] overflow-hidden p-0">
              <div className="flex flex-col h-full">
                <SheetHeader className="flex-shrink-0 px-6 py-4 border-b">
                  <SheetTitle>Create Local User</SheetTitle>
                  <SheetDescription>
                    Create a new local user account with custom permissions. This user will be able to sign in with their username and password.
                  </SheetDescription>
                </SheetHeader>

                <form onSubmit={handleCreateLocalUser} className="flex-1 flex flex-col min-h-0">
                  <ScrollArea className="flex-1 px-6">
                    <div className="space-y-6 py-4">
                      {/* User Details Section */}
                      <div className="space-y-4">
                        <h3 className="text-lg font-medium border-b pb-2">
                          User Details
                        </h3>

                        <div className="grid grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="create-username">Username *</Label>
                            <Input
                              id="create-username"
                              value={createUserForm.username}
                              onChange={(e) =>
                                setCreateUserForm((prev) => ({
                                  ...prev,
                                  username: e.target.value,
                                }))
                              }
                              placeholder="Enter username"
                              required
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="create-email">Email</Label>
                            <Input
                              id="create-email"
                              type="email"
                              value={createUserForm.email}
                              onChange={(e) =>
                                setCreateUserForm((prev) => ({
                                  ...prev,
                                  email: e.target.value,
                                }))
                              }
                              placeholder="Enter email (optional)"
                            />
                          </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="create-password">Password *</Label>
                            <Input
                              id="create-password"
                              type="password"
                              value={createUserForm.password}
                              onChange={(e) =>
                                setCreateUserForm((prev) => ({
                                  ...prev,
                                  password: e.target.value,
                                }))
                              }
                              placeholder="Enter password"
                              required
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="create-confirmPassword">
                              Confirm Password *
                            </Label>
                            <Input
                              id="create-confirmPassword"
                              type="password"
                              value={createUserForm.confirmPassword}
                              onChange={(e) =>
                                setCreateUserForm((prev) => ({
                                  ...prev,
                                  confirmPassword: e.target.value,
                                }))
                              }
                              placeholder="Confirm password"
                              required
                            />
                          </div>
                        </div>
                      </div>

                      {/* Permissions Section */}
                      <div className="space-y-4">
                        <UserPermissionSelector
                          selectedPermissions={createUserPermissions}
                          onPermissionChange={handleCreateUserPermissionChange}
                          title="User Permissions"
                          description="Select permissions for this user. You can always modify these later."
                          showCard={false}
                          loadDefaults={true}
                        />
                      </div>
                    </div>
                  </ScrollArea>

                  <div className="flex-shrink-0 border-t px-6 py-4">
                    <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
                      <SheetClose asChild>
                        <Button type="button" variant="outline">
                          Cancel
                        </Button>
                      </SheetClose>
                      <Button
                        type="submit"
                        disabled={createLocalUserMutation.isPending}
                      >
                        {createLocalUserMutation.isPending ? (
                          <>
                            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                            Creating...
                          </>
                        ) : (
                          <>
                            <UserPlus className="h-4 w-4 mr-2" />
                            Create User
                          </>
                        )}
                      </Button>
                    </div>
                  </div>
                </form>
              </div>
            </SheetContent>
          </Sheet>
          <CreateInvitationDialog onSuccess={handleInvitationSuccess} />
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statisticsCards.map((card) => {
          const Icon = card.icon;
          return (
            <div key={card.title} className="bg-card rounded-lg border p-4">
              <div className="flex items-center gap-2">
                <Icon className={`w-4 h-4 ${card.color}`} />
                <span className="text-sm font-medium text-muted-foreground">{card.title}</span>
              </div>
              <p className="text-2xl font-bold text-foreground">
                {(statsLoading || usersLoading) ? (
                  <div className="h-8 w-16 bg-muted rounded animate-pulse"></div>
                ) : (
                  card.value
                )}
              </p>
            </div>
          );
        })}
      </div>

      {/* Pending Invitations Section - Show if any exist */}
      {pendingInvitations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Pending Invitations ({pendingInvitations.length})
            </CardTitle>
            <CardDescription>
              Users who have been invited but haven't accepted yet. Click the menu to copy invitation links.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {invitationsLoading ? (
              <div className="space-y-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex items-center space-x-4">
                    <div className="h-10 w-10 bg-gray-200 rounded-full animate-pulse"></div>
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-gray-200 rounded animate-pulse"></div>
                      <div className="h-3 bg-gray-200 rounded animate-pulse w-1/2"></div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <InvitationDataTable 
                columns={invitationColumns}
                data={pendingInvitations}
                compact={true}
              />
            )}
          </CardContent>
        </Card>
      )}

      {/* Users Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Active Users ({users.length})
          </CardTitle>
          <CardDescription>
            All users with accounts on the system. Click edit to manage their permissions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {usersLoading ? (
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center justify-between p-4 border rounded-lg animate-pulse">
                  <div className="flex items-center space-x-3">
                    <div className="h-8 w-8 bg-gray-200 rounded-full"></div>
                    <div className="space-y-2">
                      <div className="h-4 w-24 bg-gray-200 rounded"></div>
                      <div className="h-3 w-16 bg-gray-200 rounded"></div>
                    </div>
                  </div>
                  <div className="h-4 w-20 bg-gray-200 rounded"></div>
                </div>
              ))}
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-12">
              <UserPlus className="mx-auto h-16 w-16 text-gray-400" />
              <h3 className="mt-4 text-lg font-semibold text-foreground">No users yet</h3>
              <p className="mt-2 text-sm text-muted-foreground max-w-sm mx-auto">
                Get started by creating an invitation. Users will appear here once they accept their invitations.
              </p>
              <div className="mt-4">
                <CreateInvitationDialog onSuccess={handleInvitationSuccess} />
              </div>
            </div>
          ) : (
            <DataTable columns={userColumns} data={users} />
          )}
        </CardContent>
      </Card>

      {/* All Invitations History - Collapsible */}
      {(invitations || []).length > pendingInvitations.length && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Invitation History
            </CardTitle>
            <CardDescription>
              All invitations including accepted, expired, and cancelled ones.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <InvitationDataTable 
              columns={invitationColumns}
              data={invitations || []}
            />
          </CardContent>
        </Card>
      )}

      {/* Delete User Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete user "{userToDelete?.username}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end space-x-2 pt-4">
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button 
              variant="destructive" 
              onClick={confirmDeleteUser}
              disabled={deleteUserMutation.isPending}
            >
              {deleteUserMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  Delete User
                </>
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Edit User Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-3">
              <Avatar 
                className="h-8 w-8"
                src={editingUser?.avatar_url ? `v1${editingUser.avatar_url}` : undefined}
                fallback={editingUser?.username.charAt(0).toUpperCase() || "U"}
                size="sm"
              />
              Edit User: {editingUser?.username}
            </DialogTitle>
            <DialogDescription>
              Manage permissions and settings for this user. Changes will take effect immediately.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            {/* User Info */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">User Information</CardTitle>
                <CardDescription>
                  Edit basic user information. Changes will be saved when you click "Save Changes".
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="edit-username" className="text-sm font-medium">
                      Username *
                    </Label>
                    <Input
                      id="edit-username"
                      value={editUserForm.username}
                      onChange={(e) => setEditUserForm(prev => ({ ...prev, username: e.target.value }))}
                      placeholder="Enter username"
                      className="w-full"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="edit-email" className="text-sm font-medium">
                      Email
                    </Label>
                    <Input
                      id="edit-email"
                      type="email"
                      value={editUserForm.email}
                      onChange={(e) => setEditUserForm(prev => ({ ...prev, email: e.target.value }))}
                      placeholder="Enter email address"
                      className="w-full"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label className="text-sm font-medium">Account Type</Label>
                    <div className="flex items-center gap-2">
                      <Badge variant={editingUser?.user_type === 'media_server' ? 'default' : 'secondary'}>
                        {editingUser?.user_type === 'media_server' ? (
                          <>
                            <Server className="h-3 w-3 mr-1" />
                            Media Account
                          </>
                        ) : (
                          <>
                            <User className="h-3 w-3 mr-1" />
                            Local Account
                          </>
                        )}
                      </Badge>
                      <span className="text-xs text-muted-foreground">(Cannot be changed)</span>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label className="text-sm font-medium">Current Permissions</Label>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">
                        {editingUser?.permissions?.length || 0} permissions
                      </Badge>
                      <span className="text-xs text-muted-foreground">Managed below</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Permissions */}
            <UserPermissionSelector
              selectedPermissions={selectedPermissions}
              onPermissionChange={handlePermissionChange}
              title="User Permissions"
              description="Select the permissions this user should have. Changes will be saved immediately."
              showCard={true}
            />

            {/* Actions */}
            <div className="flex items-center justify-between pt-4 border-t">
              <div className="flex items-center gap-3">
                <Button
                  variant="outline"
                  onClick={handleSaveUserProfile}
                  disabled={updateUserProfileMutation.isPending || !editUserForm.username.trim()}
                >
                  {updateUserProfileMutation.isPending ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Saving Profile...
                    </>
                  ) : (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      Save Profile
                    </>
                  )}
                </Button>
                <Button
                  variant="outline"
                  onClick={handleSaveUserPermissions}
                  disabled={updateUserPermissionsMutation.isPending}
                >
                  {updateUserPermissionsMutation.isPending ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <Shield className="h-4 w-4 mr-2" />
                      Save Permissions
                    </>
                  )}
                </Button>
              </div>
              <Button
                variant="ghost"
                onClick={() => {
                  setEditDialogOpen(false);
                  setEditingUser(null);
                }}
              >
                Close
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
