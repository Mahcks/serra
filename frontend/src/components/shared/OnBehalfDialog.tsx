import { Users, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { type TMDBMediaItem, type UserWithPermissions } from "@/types";

interface OnBehalfDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedMedia: TMDBMediaItem | null;
  selectedUser: string;
  onSelectedUserChange: (userId: string) => void;
  allUsers?: { users: UserWithPermissions[] };
  onSubmit: () => void;
  isLoading: boolean;
}

export function OnBehalfDialog({
  open,
  onOpenChange,
  selectedMedia,
  selectedUser,
  onSelectedUserChange,
  allUsers,
  onSubmit,
  isLoading,
}: OnBehalfDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Request Media</DialogTitle>
          <DialogDescription>
            Choose who to request "{selectedMedia?.title || selectedMedia?.name}" for.
          </DialogDescription>
        </DialogHeader>
        
        <div className="py-4">
          <Label htmlFor="user-select" className="text-sm font-medium">
            Request for:
          </Label>
          <Select value={selectedUser} onValueChange={onSelectedUserChange}>
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
                      <span className="text-muted-foreground text-sm">({user.email})</span>
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
          <Button
            onClick={onSubmit}
            disabled={isLoading}
          >
            {isLoading && (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            )}
            Submit Request
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}