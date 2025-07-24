import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Loader2, ShieldCheck, AlertTriangle, Users, Settings, Crown, Play } from "lucide-react";
import { toast } from "sonner";

interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
  dangerous: boolean;
}

interface UserPermissionSelectorProps {
  selectedPermissions: Set<string>;
  onPermissionChange: (permissionId: string, checked: boolean) => void;
  title?: string;
  description?: string;
  showCard?: boolean;
  className?: string;
  loadDefaults?: boolean; // Whether to automatically load default permissions
}

export default function UserPermissionSelector({
  selectedPermissions,
  onPermissionChange,
  title = "User Permissions",
  description = "Select permissions for this user",
  showCard = true,
  className = "",
  loadDefaults = false
}: UserPermissionSelectorProps) {
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPermissions();
  }, []);

  useEffect(() => {
    if (loadDefaults) {
      loadDefaultPermissions();
    }
  }, [loadDefaults, permissions]);

  const loadPermissions = async () => {
    try {
      const response = await fetch('/v1/permissions/categories', {
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error('Failed to load permissions');
      }
      
      const data = await response.json();
      
      // Flatten permissions from categories
      const allPermissions: Permission[] = [];
      Object.entries(data).forEach(([categoryName, categoryPerms]: [string, any]) => {
        if (Array.isArray(categoryPerms)) {
          categoryPerms.forEach((perm: Permission) => {
            allPermissions.push({
              ...perm,
              category: categoryName
            });
          });
        }
      });
      
      setPermissions(allPermissions);
    } catch (error) {
      console.error('Failed to load permissions:', error);
      toast.error("Failed to load available permissions.");
    } finally {
      setLoading(false);
    }
  };

  const loadDefaultPermissions = async () => {
    try {
      const response = await fetch('/v1/settings/default-permissions', {
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error('Failed to load default permissions');
      }
      
      const data = await response.json();
      
      // Apply default permissions
      Object.entries(data.permissions).forEach(([permissionId, enabled]: [string, any]) => {
        if (enabled && permissions.some(p => p.id === permissionId)) {
          onPermissionChange(permissionId, true);
        }
      });
    } catch (error) {
      console.error('Failed to load default permissions:', error);
      // Don't show error toast here as it's not critical
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category.toLowerCase()) {
      case 'owner':
        return <Crown className="h-4 w-4 text-yellow-500" />;
      case 'administrative':
        return <Settings className="h-4 w-4 text-red-500" />;
      case 'request content':
        return <Play className="h-4 w-4 text-blue-500" />;
      case 'manage requests':
        return <Users className="h-4 w-4 text-green-500" />;
      default:
        return <ShieldCheck className="h-4 w-4 text-gray-500" />;
    }
  };

  const PermissionContent = () => {
    if (loading) {
      return (
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          <span>Loading permissions...</span>
        </div>
      );
    }

    // Group permissions by category
    const permissionsByCategory = permissions.reduce((acc, permission) => {
      if (!acc[permission.category]) {
        acc[permission.category] = [];
      }
      acc[permission.category].push(permission);
      return acc;
    }, {} as Record<string, Permission[]>);

    if (Object.keys(permissionsByCategory).length === 0) {
      return (
        <div className="text-center py-8 text-muted-foreground">
          No permissions available or failed to load permissions.
        </div>
      );
    }

    return (
      <div className="space-y-6">
        {Object.entries(permissionsByCategory).map(([category, categoryPermissions]) => (
          <div key={category} className="space-y-3">
            <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
              {getCategoryIcon(category)}
              <span>{category}</span>
              {category.toLowerCase() === 'owner' && (
                <AlertTriangle className="h-4 w-4 text-amber-500 ml-1" />
              )}
            </div>
            
            <div className="grid grid-cols-1 gap-3 pl-6">
              {categoryPermissions.map((permission) => (
                <div key={permission.id} className="flex items-start space-x-3">
                  <Checkbox
                    id={permission.id}
                    checked={selectedPermissions.has(permission.id)}
                    onCheckedChange={(checked) => 
                      onPermissionChange(permission.id, checked === true)
                    }
                  />
                  <div className="flex-1 space-y-1">
                    <Label 
                      htmlFor={permission.id} 
                      className={`text-sm font-medium cursor-pointer flex items-center gap-2 ${
                        permission.dangerous ? 'text-amber-700 dark:text-amber-400' : 'text-foreground'
                      }`}
                    >
                      {permission.name}
                      {permission.dangerous && (
                        <AlertTriangle className="h-3 w-3" />
                      )}
                    </Label>
                    <p className="text-xs text-muted-foreground">
                      {permission.description}
                    </p>
                    {permission.dangerous && (
                      <p className="text-xs text-amber-600 dark:text-amber-400 font-medium">
                        ⚠️ Administrative permission - use with caution
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  };

  if (showCard) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            {title}
          </CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent>
          <PermissionContent />
        </CardContent>
      </Card>
    );
  }

  return (
    <div className={className}>
      <div className="mb-4">
        <h3 className="text-lg font-medium flex items-center gap-2">
          <ShieldCheck className="h-5 w-5" />
          {title}
        </h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <PermissionContent />
    </div>
  );
}