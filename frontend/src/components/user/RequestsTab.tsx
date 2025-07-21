import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Users, Plus, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { requestsApi } from "@/lib/api";
import { UserRequestsDataTable } from "@/components/user/requests/user-requests-data-table";
import { createUserRequestsColumns } from "@/components/user/requests/user-requests-columns";

export function RequestsTab() {
  const navigate = useNavigate();

  // Fetch user requests
  const { data: userRequests = [], isLoading: isLoadingRequests } = useQuery({
    queryKey: ["user-requests"],
    queryFn: requestsApi.getUserRequests,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-2 bg-muted/50 rounded-lg border">
          <Users className="w-6 h-6 text-muted-foreground" />
        </div>
        <div>
          <h2 className="text-2xl font-bold text-foreground">My Requests</h2>
          <p className="text-muted-foreground">Track your content requests</p>
        </div>
      </div>
      
      {isLoadingRequests ? (
        <div className="flex justify-center py-12">
          <div className="flex items-center gap-3 text-muted-foreground">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>Loading your requests...</span>
          </div>
        </div>
      ) : userRequests.length === 0 ? (
        <div className="bg-muted/50 rounded-xl p-8 text-center border">
          <Users className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-foreground mb-2">No Requests Yet</h3>
          <p className="text-muted-foreground mb-6">Start by requesting some content you'd like to watch</p>
          <Button onClick={() => navigate('/requests?tab=discover')}>
            <Plus className="w-4 h-4 mr-2" />
            Browse Content
          </Button>
        </div>
      ) : (
        <UserRequestsDataTable
          columns={createUserRequestsColumns(navigate)}
          data={userRequests}
        />
      )}
    </div>
  );
}