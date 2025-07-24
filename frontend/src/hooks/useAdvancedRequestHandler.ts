import { useState, useCallback, useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { requestsApi, backendApi, discoverApi } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { handleApiError, ERROR_CODES, getErrorCode } from "@/utils/errorHandling";
import type { 
  TMDBMediaItem, 
  CreateRequestRequest, 
  UserWithPermissions, 
  MediaStatusResponse 
} from "@/types";

interface UseAdvancedRequestHandlerOptions {
  onSuccess?: () => void;
  queryKeysToInvalidate?: string[][];
}

export function useAdvancedRequestHandler({
  onSuccess,
  queryKeysToInvalidate = []
}: UseAdvancedRequestHandlerOptions = {}) {
  const { user } = useAuth();
  const queryClient = useQueryClient();

  // State for request dialogs
  const [currentRequestItem, setCurrentRequestItem] = useState<TMDBMediaItem | null>(null);
  const [showOnBehalfDialog, setShowOnBehalfDialog] = useState(false);
  const [selectedMedia, setSelectedMedia] = useState<TMDBMediaItem | null>(null);
  const [selectedUser, setSelectedUser] = useState<string>("myself");
  const [showSeasonDialog, setShowSeasonDialog] = useState(false);
  const [selectedSeasons, setSelectedSeasons] = useState<number[]>([]);

  // Check if user can request on behalf of others
  const { data: currentUserPermissions } = useQuery({
    queryKey: ["current-user-permissions"],
    queryFn: backendApi.getCurrentUserPermissions,
    enabled: !!user,
  });

  const canRequestOnBehalf = useMemo(() => {
    if (!user) return false;
    // Admin users can always request on behalf of others
    if (user.isAdmin) return true;
    
    // Check if user has owner or requests.manage permission
    const userPermissions = currentUserPermissions?.permissions || [];
    return userPermissions.some(
      (perm) => perm.id === "owner" || perm.id === "requests.manage"
    );
  }, [user, currentUserPermissions]);

  // Fetch all users for on-behalf requests
  const { data: allUsers } = useQuery({
    queryKey: ["users"],
    queryFn: backendApi.getUsers,
    enabled: canRequestOnBehalf,
  });

  // Get media status for a specific item
  const getMediaStatus = useCallback(async (tmdbId: number, mediaType: 'movie' | 'tv') => {
    if (!user) return null;
    try {
      return await discoverApi.getMediaStatus(tmdbId, mediaType);
    } catch (error) {
      console.error("Failed to get media status:", error);
      return null;
    }
  }, [user]);

  // Create request mutation
  const createRequestMutation = useMutation({
    mutationFn: (data: CreateRequestRequest) => {
      return requestsApi.createRequest(data);
    },
    onSuccess: (newRequest) => {
      const displayTitle =
        newRequest.title ||
        currentRequestItem?.title ||
        currentRequestItem?.name ||
        "the requested content";

      if (newRequest.status === "approved") {
        toast.success(`🎉 Request Approved!`, {
          description: `"${displayTitle}" was automatically approved and will be downloaded soon.`,
          duration: 5000,
        });
      } else if (newRequest.status === "fulfilled") {
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

      // Invalidate provided query keys
      queryKeysToInvalidate.forEach(queryKey => {
        queryClient.invalidateQueries({ queryKey });
      });

      // Also invalidate media status for this specific item
      if (currentRequestItem) {
        const mediaType = currentRequestItem.media_type || 
          (currentRequestItem.first_air_date ? 'tv' : 'movie');
        queryClient.invalidateQueries({
          queryKey: ["mediaStatus", currentRequestItem.id, mediaType],
        });
      }

      // Reset state
      setCurrentRequestItem(null);
      setShowOnBehalfDialog(false);
      setShowSeasonDialog(false);
      setSelectedMedia(null);
      setSelectedSeasons([]);

      onSuccess?.();
    },
    onError: (error: any) => {
      console.error("Request creation failed:", error);

      const { message } = handleApiError(error);
      const errorCode = getErrorCode(error);

      if (errorCode === ERROR_CODES.DUPLICATE_REQUEST) {
        toast.error(`🔄 Already Requested`, {
          description: `You've already requested this content.`,
          duration: 4000,
        });
      } else {
        const title = errorCode ? `❌ Request Failed (${errorCode})` : `❌ Request Failed`;
        toast.error(title, {
          description: message,
          duration: 4000,
        });
      }

      // Reset state
      setCurrentRequestItem(null);
      setShowOnBehalfDialog(false);
      setShowSeasonDialog(false);
      setSelectedMedia(null);
      setSelectedSeasons([]);
    },
  });

  // Main request handler - entry point for all requests
  const handleRequest = useCallback(
    (item: TMDBMediaItem) => {
      if (!user) {
        toast.error("Authentication required", {
          description: "Please log in to make requests.",
        });
        return;
      }

      const mediaType = item.media_type || (item.first_air_date ? "tv" : "movie");

      // For TV series, show season selection dialog first
      if (mediaType === "tv") {
        setSelectedMedia(item);
        setShowSeasonDialog(true);
        setSelectedSeasons([]); // Reset selection
        return;
      }

      // For movies, check if we need on-behalf dialog
      if (canRequestOnBehalf && allUsers?.users && allUsers.users.length > 0) {
        setSelectedMedia(item);
        setSelectedUser("myself"); // Default to myself
        setShowOnBehalfDialog(true);
        return;
      }

      // Otherwise create request directly
      submitRequest(item);
    },
    [user, canRequestOnBehalf, allUsers]
  );

  // Submit the actual request
  const submitRequest = useCallback((item: TMDBMediaItem, seasons?: number[], onBehalfOf?: string) => {
    const mediaType = item.media_type || (item.first_air_date ? "tv" : "movie");
    const title = item.title || item.name || "Unknown Title";
    const posterUrl = item.poster_path
      ? `https://image.tmdb.org/t/p/w500${item.poster_path}`
      : undefined;

    const requestData: CreateRequestRequest = {
      media_type: mediaType,
      tmdb_id: item.id,
      title: title,
      poster_url: posterUrl,
      // Only include seasons for TV shows and when seasons are selected
      ...(mediaType === "tv" && seasons && seasons.length > 0 && { seasons }),
      // Include on-behalf user if specified and not "myself"
      ...(onBehalfOf && onBehalfOf !== "myself" && { on_behalf_of: onBehalfOf }),
    };

    setCurrentRequestItem(item);
    createRequestMutation.mutate(requestData);
  }, [createRequestMutation]);

  // Handle season selection submission
  const handleSeasonRequestSubmit = useCallback(() => {
    if (!selectedMedia || selectedSeasons.length === 0) return;

    // Check if we need on-behalf dialog
    if (canRequestOnBehalf && allUsers?.users && allUsers.users.length > 0) {
      setShowSeasonDialog(false);
      setSelectedUser("myself"); // Default to myself
      setShowOnBehalfDialog(true);
      return;
    }

    // Otherwise submit directly
    submitRequest(selectedMedia, selectedSeasons);
    setShowSeasonDialog(false);
  }, [selectedMedia, selectedSeasons, canRequestOnBehalf, allUsers, submitRequest]);

  // Handle on-behalf request submission
  const handleOnBehalfSubmit = useCallback(() => {
    if (!selectedMedia) return;

    const userToRequestFor = selectedUser === "myself" ? undefined : selectedUser;
    const seasons = selectedSeasons.length > 0 ? selectedSeasons : undefined;

    submitRequest(selectedMedia, seasons, userToRequestFor);
    setShowOnBehalfDialog(false);
  }, [selectedMedia, selectedUser, selectedSeasons, submitRequest]);

  return {
    // State
    showOnBehalfDialog,
    setShowOnBehalfDialog,
    selectedMedia,
    selectedUser,
    setSelectedUser,
    showSeasonDialog,
    setShowSeasonDialog,
    selectedSeasons,
    setSelectedSeasons,
    
    // Data
    allUsers,
    canRequestOnBehalf,
    
    // Functions
    handleRequest,
    handleSeasonRequestSubmit,
    handleOnBehalfSubmit,
    getMediaStatus,
    
    // Mutation state
    isRequestLoading: createRequestMutation.isPending,
  };
}