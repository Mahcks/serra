import { useQuery, useMutation } from "@tanstack/react-query";
import { createColumns } from "./columns";
import { DataTable } from "./data-table";
import { type GetAllUsersResponse } from "@/types";
import { backendApi } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetFooter,
  SheetClose,
} from "@/components/ui/sheet";
import { useState } from "react";
import { Plus } from "lucide-react";

export default function UsersPage() {
  const columns = createColumns();
  const [isCreateUserOpen, setIsCreateUserOpen] = useState(false);
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const {
    data: users,
    isLoading,
    error,
    refetch,
  } = useQuery<GetAllUsersResponse>({
    queryKey: ["users"],
    queryFn: backendApi.getUsers,
    retry: false,
    staleTime: 0,
    cacheTime: 0,
  });

  const createUserMutation = useMutation({
    mutationFn: backendApi.createLocalUser,
    onSuccess: () => {
      refetch();
      setIsCreateUserOpen(false);
      setFormData({ username: "", email: "", password: "", confirmPassword: "" });
    },
    onError: (error) => {
      console.error("Failed to create user:", error);
    },
  });

  const handleCreateUser = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (formData.password !== formData.confirmPassword) {
      alert("Passwords do not match");
      return;
    }
    
    if (formData.username.trim() === "" || formData.password.trim() === "") {
      alert("Username and password are required");
      return;
    }

    createUserMutation.mutate({
      username: formData.username,
      email: formData.email || undefined,
      password: formData.password,
    });
  };

  if (isLoading) {
    return (
      <div className="container  py-10">
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">
            Manage user accounts and permissions
          </p>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="text-muted-foreground">Loading users...</div>
        </div>
      </div>
    );
  }

  if (error) {
    const is403Error = error instanceof Error && error.message.includes('403');
    
    return (
      <div className="py-10">
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">
            Manage user accounts and permissions
          </p>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="text-red-500 text-center">
            {is403Error ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <p className="font-medium">Access Denied</p>
                  <p className="text-sm text-muted-foreground">
                    You don't have permission to manage users. Please contact an administrator.
                  </p>
                </div>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => refetch()}
                  disabled={isLoading}
                >
                  {isLoading ? "Checking..." : "Retry"}
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <p className="font-medium">Error loading users</p>
                  <p className="text-sm text-muted-foreground">
                    {error.message}
                  </p>
                </div>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => refetch()}
                  disabled={isLoading}
                >
                  {isLoading ? "Loading..." : "Retry"}
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div >
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">
            Manage user accounts and permissions
          </p>
        </div>
        <Sheet open={isCreateUserOpen} onOpenChange={setIsCreateUserOpen}>
          <SheetTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create Local User
            </Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>Create Local User</SheetTitle>
              <SheetDescription>
                Create a new local user account that can authenticate independently of your media server.
              </SheetDescription>
            </SheetHeader>
            <form onSubmit={handleCreateUser} className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="username">Username *</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                  placeholder="Enter username"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                  placeholder="Enter email (optional)"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password *</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                  placeholder="Enter password"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm Password *</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData(prev => ({ ...prev, confirmPassword: e.target.value }))}
                  placeholder="Confirm password"
                  required
                />
              </div>
              <SheetFooter>
                <SheetClose asChild>
                  <Button type="button" variant="outline">
                    Cancel
                  </Button>
                </SheetClose>
                <Button type="submit" disabled={createUserMutation.isPending}>
                  {createUserMutation.isPending ? "Creating..." : "Create User"}
                </Button>
              </SheetFooter>
            </form>
          </SheetContent>
        </Sheet>
      </div>
      <DataTable columns={columns} data={users?.users || []} />
    </div>
  );
}
