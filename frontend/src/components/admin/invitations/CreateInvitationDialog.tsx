import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Plus, Mail, User, Clock, Server } from "lucide-react";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";
import type { CreateInvitationRequest } from "@/types";

interface CreateInvitationDialogProps {
  onSuccess?: () => void;
}

export function CreateInvitationDialog({ onSuccess }: CreateInvitationDialogProps) {
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    email: "",
    username: "",
    expires_in_days: 7,
    create_media_user: true,
  });

  // Note: We no longer allow permission selection for invitations - default permissions are used

  const createInvitationMutation = useMutation({
    mutationFn: invitationsApi.createInvitation,
    onSuccess: (data) => {
      toast.success("Invitation created successfully");
      setOpen(false);
      resetForm();
      onSuccess?.();
      
      // Copy invitation link to clipboard
      navigator.clipboard.writeText(data.invite_url);
      toast.success("Invitation link copied to clipboard", {
        description: "You can now share this link with the invited user.",
      });
    },
    onError: (error: any) => {
      toast.error("Failed to create invitation", {
        description: error.response?.data?.error?.message || "Please try again.",
      });
    },
  });

  const resetForm = () => {
    setFormData({
      email: "",
      username: "",
      expires_in_days: 7,
      create_media_user: true,
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.email || !formData.username) {
      toast.error("Please fill in all required fields");
      return;
    }

    const request: CreateInvitationRequest = {
      email: formData.email,
      username: formData.username,
      permissions: [], // Empty array - default permissions will be used
      create_media_user: formData.create_media_user,
      expires_in_days: formData.expires_in_days,
    };

    createInvitationMutation.mutate(request);
  };


  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Invitation
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            Create New Invitation
          </DialogTitle>
          <DialogDescription>
            Create an invitation link for a new Serra account. The link will be copied to your clipboard so you can share it directly with the user. Default permissions will be assigned automatically.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          <ScrollArea className="max-h-[60vh] pr-4">
            <div className="space-y-6">
              {/* Basic Information */}
              <div className="space-y-4">
                <div className="flex items-center gap-2 mb-3">
                  <User className="h-4 w-4" />
                  <h3 className="font-medium">User Information</h3>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="email">Email Address *</Label>
                    <Input
                      id="email"
                      type="email"
                      placeholder="user@example.com"
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                      required
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <Label htmlFor="username">Username *</Label>
                    <Input
                      id="username"
                      type="text"
                      placeholder="username"
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                      required
                    />
                  </div>
                </div>
              </div>

              <Separator />

              {/* Settings */}
              <div className="space-y-4">
                <div className="flex items-center gap-2 mb-3">
                  <Clock className="h-4 w-4" />
                  <h3 className="font-medium">Invitation Settings</h3>
                </div>
                
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="expires_in_days">Expires In (Days)</Label>
                    <Select
                      value={formData.expires_in_days.toString()}
                      onValueChange={(value) => setFormData({ ...formData, expires_in_days: parseInt(value) })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1">1 Day</SelectItem>
                        <SelectItem value="3">3 Days</SelectItem>
                        <SelectItem value="7">7 Days (Default)</SelectItem>
                        <SelectItem value="14">14 Days</SelectItem>
                        <SelectItem value="30">30 Days</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <Server className="h-4 w-4" />
                        <Label htmlFor="create_media_user">Media Server Account</Label>
                      </div>
                      <p className="text-sm text-gray-600">
                        Creates a media server account for the user so they can login to Jellyfin/Emby. If disabled, it will create a local account instead.
                      </p>
                    </div>
                    <Switch
                      id="create_media_user"
                      checked={formData.create_media_user}
                      onCheckedChange={(checked) => setFormData({ ...formData, create_media_user: checked })}
                    />
                  </div>
                </div>
              </div>

              {/* Note: Permission selection removed - invitations now use default permissions */}
            </div>
          </ScrollArea>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createInvitationMutation.isPending}
            >
              {createInvitationMutation.isPending ? "Creating..." : "Create Invitation"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}