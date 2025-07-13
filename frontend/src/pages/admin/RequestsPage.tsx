import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { requestsApi } from "@/lib/api";
import type { Request, RequestStatistics } from "@/types";
import {
  Users,
  Search,
  Filter,
  MoreHorizontal,
  CheckCircle,
  XCircle,
  Clock,
  User,
  Calendar,
  Film,
  Tv,
  Loader2,
} from "lucide-react";

const statusColors = {
  pending: "bg-yellow-500/20 text-yellow-700 dark:text-yellow-300",
  approved: "bg-blue-500/20 text-blue-700 dark:text-blue-300",
  denied: "bg-red-500/20 text-red-700 dark:text-red-300",
  fulfilled: "bg-green-500/20 text-green-700 dark:text-green-300",
};

const statusIcons = {
  pending: Clock,
  approved: CheckCircle,
  denied: XCircle,
  fulfilled: CheckCircle,
};

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
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
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
    onError: (error: any) => {
      const message = error.response?.data?.error?.message || error.message || "Failed to update request";
      toast.error(message);
    },
  });

  // Filter requests based on search and status
  const filteredRequests = (allRequests || []).filter((request) => {
    const matchesSearch = (request.title || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
                         (request.user_id || '').toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter === "all" || request.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const handleUpdateRequest = useCallback((requestId: number, status: string, notes?: string) => {
    updateRequestMutation.mutate({ id: requestId, status, notes });
  }, [updateRequestMutation]);

  const openUpdateDialog = useCallback((request: Request) => {
    setSelectedRequest(request);
    setIsUpdateDialogOpen(true);
  }, []);

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

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search requests..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-muted-foreground" />
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="approved">Approved</SelectItem>
              <SelectItem value="denied">Denied</SelectItem>
              <SelectItem value="fulfilled">Fulfilled</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Requests Table */}
      <div className="bg-card rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Content</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Requested</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredRequests.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8">
                  <div className="text-muted-foreground">
                    {searchQuery || statusFilter !== "all" 
                      ? "No requests found matching your filters" 
                      : "No requests yet"
                    }
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              filteredRequests.map((request) => {
                const StatusIcon = statusIcons[request.status as keyof typeof statusIcons] || Clock;
                return (
                  <TableRow key={request.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        {request.poster_url && (
                          <img
                            src={request.poster_url}
                            alt={request.title}
                            className="w-10 h-15 object-cover rounded"
                          />
                        )}
                        <div>
                          <div className="font-medium text-foreground">{request.title}</div>
                          {request.notes && (
                            <div className="text-sm text-muted-foreground mt-1">
                              {request.notes}
                            </div>
                          )}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {request.media_type === 'movie' ? (
                          <Film className="w-4 h-4 text-muted-foreground" />
                        ) : (
                          <Tv className="w-4 h-4 text-muted-foreground" />
                        )}
                        <span className="capitalize">{request.media_type}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-muted-foreground" />
                        <span>{request.username || request.user_id}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="secondary"
                        className={statusColors[request.status as keyof typeof statusColors]}
                      >
                        <StatusIcon className="w-3 h-3 mr-1" />
                        {request.status ? request.status.charAt(0).toUpperCase() + request.status.slice(1) : 'Unknown'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Calendar className="w-4 h-4" />
                        <span>{new Date(request.created_at).toLocaleDateString()}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => openUpdateDialog(request)}
                        disabled={updateRequestMutation.isPending}
                      >
                        <MoreHorizontal className="w-4 h-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

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