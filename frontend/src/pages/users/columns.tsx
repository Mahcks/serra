"use client";

import { type ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Avatar } from "@/components/ui/avatar";
import { type UserWithPermissions } from "@/types";
import { useNavigate } from "react-router-dom";

export const createColumns = (): ColumnDef<UserWithPermissions>[] => {
  const navigate = useNavigate();
  
  return [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "username",
    header: "User",
    cell: ({ row }) => {
      const user = row.original;
      return (
        <div className="flex items-center gap-3">
          <Avatar 
            src={user.avatar_url ? `http://localhost:9090/v1${user.avatar_url}` : undefined}
            alt={user.username}
            fallback={user.username.charAt(0)}
            size="sm"
          />
          <span className="font-medium">{user.username}</span>
        </div>
      );
    },
  },
  {
    accessorKey: "user_type",
    header: "Type",
    cell: ({ row }) => {
      const userType = row.getValue("user_type") as string;
      return (
        <div className="flex items-center">
          <span className={`px-2 py-1 text-xs rounded-full font-medium ${
            userType === "local" 
              ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300"
              : "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300"
          }`}>
            {userType === "local" ? "Local" : "Media Server"}
          </span>
        </div>
      );
    },
  },
  {
    accessorKey: "email",
    header: "Email",
    cell: ({ row }) => {
      const email = row.getValue("email") as string;
      return (
        <div className="max-w-[200px] truncate">
          {email || <span className="text-muted-foreground">No email</span>}
        </div>
      );
    },
  },
  {
    accessorKey: "permissions",
    header: "Permissions",
    cell: ({ row }) => {
      const permissions = row.getValue(
        "permissions"
      ) as UserWithPermissions["permissions"];
      
      if (!permissions || permissions.length === 0) {
        return (
          <div className="max-w-[200px] truncate">
            <span className="text-muted-foreground">No permissions</span>
          </div>
        );
      }

      const count = permissions.length;
      const displayPermissions = permissions
        .slice(0, 3)
        .map((p) => p.name)
        .join(", ");

      return (
        <div className="max-w-[200px] truncate">
          {displayPermissions}
          {count > 3 && (
            <span className="text-muted-foreground"> +{count - 3} more</span>
          )}
        </div>
      );
    },
  },
  {
    id: "actions",
    cell: ({ row }) => {
      const user = row.original;

      return (
        <Button 
          aria-label="Actions"
          onClick={() => navigate(`/admin/users/${user.id}/settings`)}
        >
          Edit
        </Button>
      );
    },
  },
];
};
