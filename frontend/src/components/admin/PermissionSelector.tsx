import { useState, useEffect } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
  dangerous: boolean;
}

interface PermissionCategories {
  [category: string]: Permission[];
}

interface PermissionSelectorProps {
  selectedPermissions: Set<string>;
  onPermissionChange: (permissionId: string, checked: boolean) => void;
  title?: string;
  description?: string;
  showCard?: boolean;
  className?: string;
}

export default function PermissionSelector({
  selectedPermissions,
  onPermissionChange,
  title = "Permissions",
  description = "Select permissions to assign",
  showCard = true,
  className = "",
}: PermissionSelectorProps) {
  const [categories, setCategories] = useState<PermissionCategories>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPermissions();
  }, []);

  const loadPermissions = async () => {
    try {
      setLoading(true);
      const response = await fetch('/v1/permissions/categories', {
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error('Failed to load permissions');
      }
      
      const data: PermissionCategories = await response.json();
      setCategories(data);
    } catch (error) {
      console.error('Failed to load permissions:', error);
      toast.error("Failed to load permissions. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const PermissionContent = () => {
    if (loading) {
      return (
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          <span className="text-sm text-muted-foreground">Loading permissions...</span>
        </div>
      );
    }

    if (Object.keys(categories).length === 0) {
      return (
        <div className="py-8 text-center">
          <p className="text-sm text-muted-foreground">No permissions available</p>
        </div>
      );
    }

    return (
      <div className="space-y-6">
        {Object.entries(categories).map(([categoryName, permissions]) => (
          <div key={categoryName} className="space-y-4">
            <h4 className="text-sm font-semibold text-foreground">{categoryName}</h4>
            <div className="space-y-3 pl-4">
              {permissions.map((permission) => (
                <div key={permission.id} className="space-y-2">
                  <div className="flex items-center space-x-3">
                    <Checkbox
                      id={`perm-${permission.id}`}
                      checked={selectedPermissions.has(permission.id)}
                      onCheckedChange={(checked) => onPermissionChange(permission.id, checked === true)}
                    />
                    <Label 
                      htmlFor={`perm-${permission.id}`} 
                      className={`text-sm font-medium ${permission.dangerous ? 'text-amber-600 dark:text-amber-400' : ''}`}
                    >
                      {permission.name}
                      {permission.dangerous && (
                        <span className="ml-1 text-xs font-normal text-amber-500">(Admin)</span>
                      )}
                    </Label>
                  </div>
                  <p className="text-xs text-muted-foreground ml-6">
                    {permission.description}
                  </p>
                </div>
              ))}
            </div>
          </div>
        ))}
        
        {selectedPermissions.size > 0 && (
          <div className="mt-4 p-3 bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800/50 rounded-lg">
            <p className="text-xs text-blue-600 dark:text-blue-400 font-medium">
              Selected permissions ({selectedPermissions.size}): {Array.from(selectedPermissions).join(', ')}
            </p>
          </div>
        )}
      </div>
    );
  };

  if (showCard) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
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
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-medium">{title}</h3>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
        <PermissionContent />
      </div>
    </div>
  );
}