import type { ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Copy, Ban, Trash2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Invitation } from "@/types";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { invitationsApi } from "@/lib/invitations-api";

interface InvitationActionsProps {
  invitation: Invitation;
  onCancel: (id: number) => void;
  onDelete: (id: number) => void;
}

function InvitationActions({ invitation, onCancel, onDelete }: InvitationActionsProps) {
  const copyInviteLink = async () => {
    try {
      const response = await invitationsApi.getInvitationLink(invitation.id);
      navigator.clipboard.writeText(response.invite_url);
      toast.success("Invitation link copied to clipboard");
    } catch (error) {
      toast.error("Failed to get invitation link");
    }
  };


  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <span className="sr-only">Open menu</span>
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>Actions</DropdownMenuLabel>
        {invitation.status === 'pending' && (
          <>
            <DropdownMenuItem onClick={copyInviteLink}>
              <Copy className="mr-2 h-4 w-4" />
              Copy invite link
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem 
              onClick={() => onCancel(invitation.id)}
              className="text-orange-600"
            >
              <Ban className="mr-2 h-4 w-4" />
              Cancel invitation
            </DropdownMenuItem>
          </>
        )}
        <DropdownMenuItem 
          onClick={() => onDelete(invitation.id)}
          className="text-red-600"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete invitation
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function createInvitationColumns(
  onCancel: (id: number) => void,
  onDelete: (id: number) => void
): ColumnDef<Invitation>[] {
  return [
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("email")}</div>
      ),
    },
    {
      accessorKey: "username",
      header: "Username",
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("username")}</div>
      ),
    },
    {
      accessorKey: "inviter_username",
      header: "Invited By",
      cell: ({ row }) => {
        const inviterUsername = row.getValue("inviter_username") as string;
        return (
          <div className="text-sm text-gray-600">
            {inviterUsername || "Unknown"}
          </div>
        );
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = row.getValue("status") as string;
        const getStatusColor = (status: string) => {
          switch (status) {
            case 'pending':
              return 'bg-yellow-100 text-yellow-800 border-yellow-200';
            case 'accepted':
              return 'bg-green-100 text-green-800 border-green-200';
            case 'expired':
              return 'bg-gray-100 text-gray-800 border-gray-200';
            case 'cancelled':
              return 'bg-red-100 text-red-800 border-red-200';
            default:
              return 'bg-gray-100 text-gray-800 border-gray-200';
          }
        };
        
        return (
          <Badge variant="outline" className={getStatusColor(status)}>
            {status.charAt(0).toUpperCase() + status.slice(1)}
          </Badge>
        );
      },
    },
    {
      accessorKey: "create_media_user",
      header: "Media Account",
      cell: ({ row }) => {
        const createMediaUser = row.getValue("create_media_user") as boolean;
        return (
          <Badge variant={createMediaUser ? "default" : "secondary"}>
            {createMediaUser ? "Yes" : "No"}
          </Badge>
        );
      },
    },
    {
      accessorKey: "expires_at",
      header: "Expires",
      cell: ({ row }) => {
        const expiresAt = row.getValue("expires_at") as string;
        const expirationDate = new Date(expiresAt);
        const now = new Date();
        const isExpired = expirationDate < now;
        
        return (
          <div className={`text-sm ${isExpired ? 'text-red-600' : 'text-gray-600'}`}>
            {isExpired ? 'Expired ' : 'Expires '}
            {formatDistanceToNow(expirationDate, { addSuffix: true })}
          </div>
        );
      },
    },
    {
      accessorKey: "created_at",
      header: "Created",
      cell: ({ row }) => {
        const createdAt = row.getValue("created_at") as string;
        return (
          <div className="text-sm text-gray-600">
            {formatDistanceToNow(new Date(createdAt), { addSuffix: true })}
          </div>
        );
      },
    },
    {
      id: "actions",
      cell: ({ row }) => (
        <InvitationActions
          invitation={row.original}
          onCancel={onCancel}
          onDelete={onDelete}
        />
      ),
    },
  ];
}