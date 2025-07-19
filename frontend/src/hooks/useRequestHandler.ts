import { useState, useCallback, useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useAuth } from '@/lib/auth';
import { requestsApi, backendApi } from '@/lib/api';
import { type TMDBMediaItem, type UserWithPermissions, type CreateRequestRequest } from '@/types';

export function useRequestHandler() {
  const { user } = useAuth();
  const queryClient = useQueryClient();

  // On-behalf request state
  const [showOnBehalfDialog, setShowOnBehalfDialog] = useState(false);
  const [selectedMedia, setSelectedMedia] = useState<TMDBMediaItem | null>(null);
  const [selectedUser, setSelectedUser] = useState<string>('');
  const [currentRequestItem, setCurrentRequestItem] = useState<TMDBMediaItem | null>(null);

  // Fetch current user's detailed permissions
  const { data: currentUserPermissions } = useQuery({
    queryKey: ["current-user-permissions"],
    queryFn: backendApi.getCurrentUserPermissions,
    enabled: !!user,
  });

  // Check if user can request on behalf of others
  const canRequestOnBehalf = useMemo(() => {
    if (!user) return false;
    
    // Admin users can always request on behalf of others
    if (user.is_admin) return true;
    
    // Check if user has owner or requests.manage permission
    const userPermissions = currentUserPermissions?.permissions || [];
    return userPermissions.some((perm: any) => 
      perm.id === 'owner' || perm.id === 'requests.manage'
    );
  }, [user, currentUserPermissions]);

  // Fetch all users for on-behalf requests (only if user has permission)
  const { data: allUsers } = useQuery<{ users: UserWithPermissions[] }>({
    queryKey: ["all-users"],
    queryFn: backendApi.getUsers,
    enabled: canRequestOnBehalf,
  });

  // Create request mutation
  const createRequestMutation = useMutation({
    mutationFn: (data: CreateRequestRequest) => {
      console.log("🚀 Mutation function called with:", data);
      return requestsApi.createRequest(data);
    },
    onSuccess: (newRequest) => {
      console.log("✅ Request creation successful:", newRequest);
      
      // Use title from response or fallback to the original request data
      const displayTitle = newRequest.title || 
                          currentRequestItem?.title || 
                          currentRequestItem?.name || 
                          selectedMedia?.title || 
                          selectedMedia?.name || 
                          "the requested content";
      
      if (newRequest.status === 'approved') {
        toast.success(`🎉 Request Approved!`, {
          description: `"${displayTitle}" was automatically approved and will be downloaded soon.`,
          duration: 5000,
        });
      } else if (newRequest.status === 'fulfilled') {
        toast.success(`✅ Request Fulfilled!`, {
          description: `"${displayTitle}" is already available in your library.`,
          duration: 4000,
        });
      } else {
        toast.success(`📝 Request Submitted!`, {
          description: `Your request for "${displayTitle}" has been submitted for review.`,
          duration: 4000,
        });
      }
      
      // Refresh queries
      queryClient.invalidateQueries({ queryKey: ["user-requests"] });
      queryClient.invalidateQueries({ queryKey: ["movies"] });
      queryClient.invalidateQueries({ queryKey: ["series"] });
      queryClient.invalidateQueries({ queryKey: ["trending"] });
      queryClient.invalidateQueries({ queryKey: ["popular"] });
      queryClient.invalidateQueries({ queryKey: ["upcoming"] });
      
      // Clear current request item
      setCurrentRequestItem(null);
    },
    onError: (error: any) => {
      console.error("❌ Request creation failed:", error);
      
      const statusCode = error.response?.status;
      const errorData = error.response?.data;
      const errorMessage = errorData?.error?.message || errorData?.message || error.message;
      
      // Handle specific error cases with more helpful messages
      if (statusCode === 400) {
        if (errorMessage?.toLowerCase().includes('already requested')) {
          toast.error(`🔄 Already Requested`, {
            description: `You've already requested this content. Check your requests page for status.`,
            duration: 4000,
          });
        } else if (errorMessage?.toLowerCase().includes('already in library')) {
          toast.error(`📚 Already Available`, {
            description: `This content is already available in your library.`,
            duration: 4000,
          });
        } else {
          toast.error(`❌ Invalid Request`, {
            description: errorMessage || "The request contains invalid data. Please try again.",
            duration: 4000,
          });
        }
      } else if (statusCode === 401) {
        toast.error(`🔐 Authentication Required`, {
          description: "Please log in again to make requests.",
          duration: 4000,
        });
      } else if (statusCode === 403) {
        toast.error(`🚫 Permission Denied`, {
          description: "You don't have permission to request this type of content.",
          duration: 4000,
        });
      } else if (statusCode === 429) {
        toast.error(`⏰ Too Many Requests`, {
          description: "You're making requests too quickly. Please wait a moment and try again.",
          duration: 5000,
        });
      } else if (statusCode >= 500) {
        toast.error(`🔧 Server Error`, {
          description: "There was a problem with the server. Please try again later.",
          duration: 4000,
        });
      } else {
        toast.error(`❌ Request Failed`, {
          description: errorMessage || "Failed to create request. Please try again.",
          duration: 4000,
        });
      }
      
      // Clear current request item on error too
      setCurrentRequestItem(null);
    },
  });

  const handleRequest = useCallback((item: TMDBMediaItem) => {
    console.log("🎬 Request submitted for:", item);
    console.log("🔍 handleRequest debug:", { 
      canRequestOnBehalf, 
      hasUsers: allUsers?.users?.length > 0,
      allUsers: allUsers 
    });
    
    // If user can request on behalf of others, show dialog to choose
    if (canRequestOnBehalf && allUsers?.users && allUsers.users.length > 0) {
      console.log("✅ Showing on-behalf dialog");
      setSelectedMedia(item);
      setSelectedUser('myself'); // Default to myself
      setShowOnBehalfDialog(true);
      return;
    }
    
    console.log("➡️ Creating request directly");
    
    // Otherwise create request directly
    const mediaType = item.media_type || (item.first_air_date ? 'tv' : 'movie');
    const title = item.title || item.name || 'Unknown Title';
    const posterUrl = item.poster_path ? `https://image.tmdb.org/t/p/w500${item.poster_path}` : undefined;

    const requestData = {
      media_type: mediaType,
      tmdb_id: item.id,
      title: title,
      poster_url: posterUrl,
    };
    
    console.log("📤 Sending request data:", requestData);
    
    // Set current item for toast display
    setCurrentRequestItem(item);
    
    createRequestMutation.mutate(requestData);
  }, [createRequestMutation, canRequestOnBehalf, allUsers]);

  // Handle on-behalf request submission
  const handleOnBehalfSubmit = useCallback(() => {
    if (!selectedMedia) return;
    
    const mediaType = selectedMedia.media_type || (selectedMedia.first_air_date ? 'tv' : 'movie');
    const title = selectedMedia.title || selectedMedia.name || 'Unknown Title';
    const posterUrl = selectedMedia.poster_path ? `https://image.tmdb.org/t/p/w500${selectedMedia.poster_path}` : undefined;

    const requestData = {
      media_type: mediaType,
      tmdb_id: selectedMedia.id,
      title: title,
      poster_url: posterUrl,
      on_behalf_of: selectedUser && selectedUser !== 'myself' ? selectedUser : undefined,
    };
    
    console.log("📤 Sending on-behalf request data:", requestData);
    
    // Set current item for toast display
    setCurrentRequestItem(selectedMedia);
    
    createRequestMutation.mutate(requestData);
    setShowOnBehalfDialog(false);
    setSelectedMedia(null);
    setSelectedUser('');
  }, [selectedMedia, selectedUser, createRequestMutation]);

  return {
    // State
    showOnBehalfDialog,
    setShowOnBehalfDialog,
    selectedMedia,
    setSelectedMedia,
    selectedUser,
    setSelectedUser,
    
    // Computed values
    canRequestOnBehalf,
    allUsers,
    
    // Mutations
    createRequestMutation,
    
    // Handlers
    handleRequest,
    handleOnBehalfSubmit,
  };
}