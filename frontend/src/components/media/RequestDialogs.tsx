import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
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
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Users, Loader2 } from "lucide-react";
import { discoverApi } from "@/lib/api";
import type { TMDBMediaItem, UserWithPermissions, SeasonDetails } from "@/types";

interface OnBehalfRequestDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedMedia: TMDBMediaItem | null;
  selectedUser: string;
  onUserChange: (userId: string) => void;
  allUsers?: { users: UserWithPermissions[] };
  onSubmit: () => void;
  isLoading: boolean;
}

export function OnBehalfRequestDialog({
  open,
  onOpenChange,
  selectedMedia,
  selectedUser,
  onUserChange,
  allUsers,
  onSubmit,
  isLoading,
}: OnBehalfRequestDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Request Media</DialogTitle>
          <DialogDescription>
            Choose who to request "
            {selectedMedia?.title || selectedMedia?.name}" for.
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          <Label htmlFor="user-select" className="text-sm font-medium">
            Request for:
          </Label>
          <Select value={selectedUser} onValueChange={onUserChange}>
            <SelectTrigger className="w-full mt-2">
              <SelectValue placeholder="Select a user or request for yourself" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="myself">Myself</SelectItem>
              {allUsers?.users?.map((user: UserWithPermissions) => (
                <SelectItem key={user.id} value={user.id}>
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    <span>{user.username}</span>
                    {user.email && (
                      <span className="text-muted-foreground text-sm">
                        ({user.email})
                      </span>
                    )}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isLoading}>
            {isLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            Submit Request
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface SeasonSelectionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedMedia: TMDBMediaItem | null;
  selectedSeasons: number[];
  onSeasonsChange: (seasons: number[]) => void;
  onSubmit: () => void;
  isLoading: boolean;
}

export function SeasonSelectionDialog({
  open,
  onOpenChange,
  selectedMedia,
  selectedSeasons,
  onSeasonsChange,
  onSubmit,
  isLoading,
}: SeasonSelectionDialogProps) {
  const [seasonDetails, setSeasonDetails] = useState<Record<number, SeasonDetails>>({});
  const [loadingSeasons, setLoadingSeasons] = useState<Record<number, boolean>>({});

  // Get TV show details to fetch seasons
  const { data: tvDetails } = useQuery({
    queryKey: ["tvDetails", selectedMedia?.id],
    queryFn: () => {
      if (!selectedMedia) throw new Error("No media selected");
      return discoverApi.getMediaDetails(selectedMedia.id.toString(), "tv");
    },
    enabled: !!(selectedMedia && open),
  });

  const seasons = tvDetails?.seasons?.filter((season) => season.season_number > 0) || [];

  const handleSeasonToggle = (seasonNumber: number, checked: boolean) => {
    if (checked) {
      onSeasonsChange([...selectedSeasons, seasonNumber]);
    } else {
      onSeasonsChange(selectedSeasons.filter(s => s !== seasonNumber));
    }
  };

  const handleSelectAll = () => {
    const allSeasonNumbers = seasons.map(s => s.season_number);
    onSeasonsChange(allSeasonNumbers);
  };

  const handleSelectNone = () => {
    onSeasonsChange([]);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Select Seasons to Request</DialogTitle>
          <DialogDescription>
            Choose which seasons of "{selectedMedia?.title || selectedMedia?.name}" you want to request.
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          {seasons.length > 0 && (
            <div className="mb-4 flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleSelectAll}
                disabled={isLoading}
              >
                Select All
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleSelectNone}
                disabled={isLoading}
              >
                Select None
              </Button>
            </div>
          )}

          <div className="space-y-3">
            {seasons.map((season) => (
              <div
                key={season.season_number}
                className="flex items-start space-x-3 p-3 border rounded-lg"
              >
                <Checkbox
                  id={`season-${season.season_number}`}
                  checked={selectedSeasons.includes(season.season_number)}
                  onCheckedChange={(checked) =>
                    handleSeasonToggle(season.season_number, checked as boolean)
                  }
                  disabled={isLoading}
                />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <Label
                      htmlFor={`season-${season.season_number}`}
                      className="font-medium cursor-pointer"
                    >
                      Season {season.season_number}
                    </Label>
                    <Badge variant="secondary" className="text-xs">
                      {season.episode_count} episodes
                    </Badge>
                  </div>
                  
                  {season.overview && (
                    <p className="text-sm text-muted-foreground mt-1">
                      {season.overview}
                    </p>
                  )}
                  
                  {season.air_date && (
                    <p className="text-xs text-muted-foreground mt-1">
                      Aired: {new Date(season.air_date).getFullYear()}
                    </p>
                  )}
                </div>
                
                {season.poster_path && (
                  <img
                    src={`https://image.tmdb.org/t/p/w92${season.poster_path}`}
                    alt={`Season ${season.season_number}`}
                    className="w-12 h-18 object-cover rounded"
                  />
                )}
              </div>
            ))}
          </div>

          {seasons.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <p>No seasons available for selection.</p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={onSubmit}
            disabled={selectedSeasons.length === 0 || isLoading}
          >
            {isLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            Request {selectedSeasons.length} Season{selectedSeasons.length !== 1 ? 's' : ''}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}