import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
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

interface DefaultPermissionsData {
  permissions: Record<string, boolean>;
}

export default function DynamicDefaultPermissions() {
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [defaultSettings, setDefaultSettings] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    Promise.all([loadPermissions(), loadDefaultSettings()]).finally(() => {
      setLoading(false);
    });
  }, []);

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
    }
  };

  const loadDefaultSettings = async () => {
    try {
      const response = await fetch('/v1/settings/default-permissions', {
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error('Failed to load default permission settings');
      }
      
      const data: DefaultPermissionsData = await response.json();
      setDefaultSettings(data.permissions);
    } catch (error) {
      console.error('Failed to load default settings:', error);
      toast.error("Failed to load default permission settings.");
    }
  };

  const handlePermissionChange = (permissionId: string, checked: boolean) => {
    setDefaultSettings(prev => ({
      ...prev,
      [permissionId]: checked
    }));
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      const response = await fetch('/v1/settings/default-permissions', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          permissions: defaultSettings
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Failed to save default permissions' }));
        throw new Error(errorData.message || 'Failed to save default permissions');
      }

      toast.success("Default permissions saved successfully.");
    } catch (error) {
      console.error('Failed to save default permissions:', error);
      const errorMessage = error instanceof Error ? error.message : "Failed to save default permissions. Please try again.";
      toast.error(errorMessage);
    } finally {
      setSaving(false);
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin mr-3" />
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

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ShieldCheck className="h-5 w-5" />
          Default User Permissions
        </CardTitle>
        <CardDescription>
          Configure which permissions are automatically assigned to new users when they register or are created.
          These permissions can always be modified per-user later.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
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
                    checked={defaultSettings[permission.id] || false}
                    onCheckedChange={(checked) => 
                      handlePermissionChange(permission.id, checked === true)
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
        
        {permissions.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            No permissions available or failed to load permissions.
          </div>
        )}

        <div className="flex items-center justify-end gap-3 pt-6 border-t border-border/60">
          <Button 
            onClick={handleSave}
            disabled={saving}
            className="min-w-[140px]"
          >
            {saving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Saving...
              </>
            ) : (
              <>
                <ShieldCheck className="h-4 w-4 mr-2" />
                Save Default Permissions
              </>
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}