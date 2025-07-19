import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { requestsApi } from "@/lib/api";
import type { Request, RequestStatistics } from "@/types";
import {
  Users,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
} from "lucide-react";
import { RequestsDataTable } from "@/components/admin/requests-data-table";
import { createRequestsColumns } from "@/components/admin/requests-columns";

interface UpdateRequestDialogProps {
  request: Request | null;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: (requestId: number, status: string, notes?: string) => void;
  isLoading: boolean;
}

function UpdateRequestDialog({
  request,
  isOpen,
  onClose,
  onUpdate,
  isLoading,
}: UpdateRequestDialogProps) {
  const [status, setStatus] = useState(request?.status || "pending");
  const [notes, setNotes] = useState(request?.notes || "");

  const handleSubmit = () => {
    if (!request) return;
    onUpdate(request.id, status, notes || undefined);
  };

  if (!request) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Update Request</DialogTitle>
          <DialogDescription>
            Update the status and notes for "{request.title}"
          </DialogDescription>
        </DialogHeader>
        
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Status</label>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="denied">Denied</SelectItem>
                <SelectItem value="fulfilled">Fulfilled</SelectItem>
              </SelectContent>
            </Select>
          </div>
          
          <div className="space-y-2">
            <label className="text-sm font-medium">Notes</label>
            <Textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add notes for this request..."
              rows={3}
            />
          </div>
        </div>
        
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            Update Request
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function RequestsPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [selectedRequest, setSelectedRequest] = useState<Request | null>(null);
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = useState(false);

  // Fetch all requests
  const { data: allRequests = [], isLoading, error } = useQuery({
    queryKey: ["admin-requests"],
    queryFn: requestsApi.getAllRequests,
  });

  // Fetch request statistics
  const { data: stats } = useQuery<RequestStatistics>({
    queryKey: ["request-statistics"],
    queryFn: requestsApi.getRequestStatistics,
  });

  // Update request mutation
  const updateRequestMutation = useMutation({
    mutationFn: ({ id, status, notes }: { id: number; status: string; notes?: string }) =>
      requestsApi.updateRequest(id, { status, notes }),
    onSuccess: (updatedRequest) => {
      toast.success(`Request updated to "${updatedRequest.status}"`);
      queryClient.invalidateQueries({ queryKey: ["admin-requests"] });
      queryClient.invalidateQueries({ queryKey: ["request-statistics"] });
      setIsUpdateDialogOpen(false);
      setSelectedRequest(null);
    },
    onError: (error: unknown) => {
      const errorObj = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = errorObj.response?.data?.error?.message || errorObj.message || "Failed to update request";
      toast.error(message);
    },
  });

  // Mass update mutation
  const massUpdateMutation = useMutation({
    mutationFn: async ({ ids, status }: { ids: number[]; status: string }) => {
      const promises = ids.map(id => requestsApi.updateRequest(id, { status }));
      return Promise.all(promises);
    },
    onSuccess: (results) => {
      toast.success(`Successfully updated ${results.length} request(s)`);
      queryClient.invalidateQueries({ queryKey: ["admin-requests"] });
      queryClient.invalidateQueries({ queryKey: ["request-statistics"] });
    },
    onError: (error: unknown) => {
      const errorObj = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = errorObj.response?.data?.error?.message || errorObj.message || "Failed to update requests";
      toast.error(message);
    },
  });

  const handleUpdateRequest = useCallback((requestId: number, status: string, notes?: string) => {
    updateRequestMutation.mutate({ id: requestId, status, notes });
  }, [updateRequestMutation]);

  const handleMassUpdate = useCallback((selectedIds: number[], status: string) => {
    massUpdateMutation.mutate({ ids: selectedIds, status });
  }, [massUpdateMutation]);

  const openUpdateDialog = useCallback((request: Request) => {
    setSelectedRequest(request);
    setIsUpdateDialogOpen(true);
  }, []);

  const columns = createRequestsColumns({
    onUpdateRequest: handleUpdateRequest,
    onOpenDialog: openUpdateDialog,
    navigate,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="flex items-center gap-3 text-muted-foreground">
          <Loader2 className="w-5 h-5 animate-spin" />
          <span>Loading requests...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <XCircle className="w-12 h-12 text-destructive mx-auto mb-4" />
          <p className="text-destructive">Failed to load requests</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="p-2 bg-primary/20 rounded-lg border">
          <Users className="w-6 h-6 text-primary" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-foreground">Request Management</h1>
          <p className="text-muted-foreground">Manage and review user content requests</p>
        </div>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 text-yellow-600" />
              <span className="text-sm font-medium text-muted-foreground">Pending</span>
            </div>
            <p className="text-2xl font-bold text-foreground">{stats.pending_requests}</p>
          </div>
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-blue-600" />
              <span className="text-sm font-medium text-muted-foreground">Approved</span>
            </div>
            <p className="text-2xl font-bold text-foreground">{stats.approved_requests}</p>
          </div>
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <XCircle className="w-4 h-4 text-red-600" />
              <span className="text-sm font-medium text-muted-foreground">Denied</span>
            </div>
            <p className="text-2xl font-bold text-foreground">{stats.denied_requests}</p>
          </div>
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-green-600" />
              <span className="text-sm font-medium text-muted-foreground">Fulfilled</span>
            </div>
            <p className="text-2xl font-bold text-foreground">{stats.fulfilled_requests}</p>
          </div>
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <Users className="w-4 h-4 text-muted-foreground" />
              <span className="text-sm font-medium text-muted-foreground">Total</span>
            </div>
            <p className="text-2xl font-bold text-foreground">{stats.total_requests}</p>
          </div>
        </div>
      )}

      {/* Requests Data Table */}
      <RequestsDataTable
        columns={columns}
        data={allRequests}
        onMassUpdate={handleMassUpdate}
        isUpdating={massUpdateMutation.isPending}
      />

      {/* Update Request Dialog */}
      <UpdateRequestDialog
        request={selectedRequest}
        isOpen={isUpdateDialogOpen}
        onClose={() => {
          setIsUpdateDialogOpen(false);
          setSelectedRequest(null);
        }}
        onUpdate={handleUpdateRequest}
        isLoading={updateRequestMutation.isPending}
      />
    </div>
  );
}