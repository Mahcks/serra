import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { 
  Users,
  UserPlus,
  Mail, 
  Clock, 
  CheckCircle, 
  XCircle, 
  AlertCircle,
  RefreshCw
} from "lucide-react";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";
import { backendApi } from "@/lib/api";
import { CreateInvitationDialog } from "@/components/admin/invitations/CreateInvitationDialog";
import { InvitationDataTable } from "@/components/admin/invitations/invitation-data-table";
import { createInvitationColumns } from "@/components/admin/invitations/invitation-columns";
import type { User } from "@/types";

export default function UsersPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState("users");

  // Fetch users
  const {
    data: users = [],
    isLoading: usersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const response = await backendApi.getUsers();
      return response;
    },
    retry: false,
  });

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
    onError: (error: any) => {
      toast.error("Failed to cancel invitation", {
        description: error.response?.data?.error?.message || "Please try again.",
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
    onError: (error: any) => {
      toast.error("Failed to delete invitation", {
        description: error.response?.data?.error?.message || "Please try again.",
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

  const invitationColumns = createInvitationColumns(handleCancelInvitation, handleDeleteInvitation);

  // Statistics cards data
  const statisticsCards = [
    {
      title: "Total Users",
      value: users.length,
      icon: Users,
      description: "Active users",
      color: "text-blue-600",
      bgColor: "bg-blue-50",
    },
    {
      title: "Pending Invitations",
      value: stats?.pending_count || 0,
      icon: Clock,
      description: "Awaiting response",
      color: "text-yellow-600",
      bgColor: "bg-yellow-50",
    },
    {
      title: "Recent Joins",
      value: stats?.accepted_count || 0,
      icon: CheckCircle,
      description: "Successfully joined",
      color: "text-green-600",
      bgColor: "bg-green-50",
    },
    {
      title: "Expired/Cancelled",
      value: (stats?.expired_count || 0) + (stats?.cancelled_count || 0),
      icon: XCircle,
      description: "No longer valid",
      color: "text-red-600",
      bgColor: "bg-red-50",
    },
  ];

  const pendingInvitations = invitations.filter(inv => inv.status === 'pending');

  if (usersError || invitationsError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <AlertCircle className="h-12 w-12 text-red-500" />
        <div className="text-center">
          <h3 className="text-lg font-semibold">Failed to load users data</h3>
          <p className="text-gray-600 mt-1">
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
        <div>
          <h1 className="text-2xl font-bold">Users</h1>
          <p className="text-gray-600 mt-1">
            Manage users and invitations. Create invitation links to share directly with new users.
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button 
            variant="outline" 
            onClick={() => {
              refetchUsers();
              refetchInvitations();
            }}
            disabled={usersLoading || invitationsLoading}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${(usersLoading || invitationsLoading) ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <CreateInvitationDialog onSuccess={handleInvitationSuccess} />
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statisticsCards.map((card) => {
          const Icon = card.icon;
          return (
            <Card key={card.title}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">
                  {card.title}
                </CardTitle>
                <div className={`p-2 rounded-md ${card.bgColor}`}>
                  <Icon className={`h-4 w-4 ${card.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {(statsLoading || usersLoading) ? (
                    <div className="h-8 w-16 bg-gray-200 rounded animate-pulse"></div>
                  ) : (
                    card.value
                  )}
                </div>
                <p className="text-xs text-gray-600 mt-1">
                  {card.description}
                </p>
              </CardContent>
            </Card>
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
            All users with accounts on the system.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {usersLoading ? (
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center space-x-4">
                  <div className="h-10 w-10 bg-gray-200 rounded-full animate-pulse"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-gray-200 rounded animate-pulse"></div>
                    <div className="h-3 bg-gray-200 rounded animate-pulse w-1/2"></div>
                  </div>
                </div>
              ))}
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-8">
              <UserPlus className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-semibold text-gray-900">No users yet</h3>
              <p className="mt-1 text-sm text-gray-500">
                Get started by creating an invitation.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {users.map((user: User) => (
                <div key={user.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="h-8 w-8 bg-gray-200 rounded-full flex items-center justify-center">
                      <Users className="h-4 w-4 text-gray-500" />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900">{user.username}</p>
                      <p className="text-sm text-gray-500">{user.email || 'No email'}</p>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500">
                    {user.user_type === 'media_server' ? 'Media Account' : 'Local Account'}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* All Invitations History - Collapsible */}
      {invitations.length > pendingInvitations.length && (
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
              data={invitations}
            />
          </CardContent>
        </Card>
      )}
    </div>
  );
}