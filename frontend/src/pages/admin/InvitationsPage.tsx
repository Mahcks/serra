import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { 
  Mail, 
  Users, 
  Clock, 
  CheckCircle, 
  XCircle, 
  AlertCircle,
  RefreshCw
} from "lucide-react";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";
import { CreateInvitationDialog } from "@/components/admin/invitations/CreateInvitationDialog";
import { InvitationDataTable } from "@/components/admin/invitations/invitation-data-table";
import { createInvitationColumns } from "@/components/admin/invitations/invitation-columns";

export default function InvitationsPage() {
  const queryClient = useQueryClient();

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
    staleTime: 0,
    gcTime: 0,
  });

  console.log("Invitations data:", invitations);

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
    onSuccess: (data, invitationId) => {
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
    onSuccess: (data, invitationId) => {
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

  const handleCancelInvitation = (invitationId: number) => {
    cancelInvitationMutation.mutate(invitationId);
  };

  const handleDeleteInvitation = (invitationId: number) => {
    deleteInvitationMutation.mutate(invitationId);
  };

  const handleInvitationSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ["invitations"] });
    queryClient.invalidateQueries({ queryKey: ["invitation-stats"] });
  };

  const columns = createInvitationColumns(handleCancelInvitation, handleDeleteInvitation);

  // Statistics cards data
  const statisticsCards = [
    {
      title: "Total Invitations",
      value: stats?.total_count || 0,
      icon: Mail,
      description: "All time invitations sent",
      color: "text-blue-600",
      bgColor: "bg-blue-50",
    },
    {
      title: "Pending",
      value: stats?.pending_count || 0,
      icon: Clock,
      description: "Awaiting response",
      color: "text-yellow-600",
      bgColor: "bg-yellow-50",
    },
    {
      title: "Accepted",
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

  if (invitationsError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <AlertCircle className="h-12 w-12 text-red-500" />
        <div className="text-center">
          <h3 className="text-lg font-semibold">Failed to load invitations</h3>
          <p className="text-gray-600 mt-1">
            There was an error loading the invitations data.
          </p>
          <Button 
            onClick={() => refetchInvitations()} 
            className="mt-4"
            variant="outline"
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
          <h1 className="text-2xl font-bold">Invitation Management</h1>
          <p className="text-gray-600 mt-1">
            Create invitation links and share them directly with users. Email setup is optional.
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button 
            variant="outline" 
            onClick={() => refetchInvitations()}
            disabled={invitationsLoading}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${invitationsLoading ? 'animate-spin' : ''}`} />
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
                  {statsLoading ? (
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

      {/* Invitations Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                All Invitations
              </CardTitle>
              <CardDescription>
                Manage and track all user invitations in your system.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {invitationsLoading ? (
            <div className="space-y-4">
              {/* Loading skeleton */}
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center space-x-4">
                  <div className="h-4 w-48 bg-gray-200 rounded animate-pulse"></div>
                  <div className="h-4 w-32 bg-gray-200 rounded animate-pulse"></div>
                  <div className="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
                  <div className="h-4 w-16 bg-gray-200 rounded animate-pulse"></div>
                </div>
              ))}
            </div>
          ) : (
            <InvitationDataTable columns={columns} data={invitations || []} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}