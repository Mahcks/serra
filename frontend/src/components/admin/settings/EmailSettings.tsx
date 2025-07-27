import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { 
  Mail, 
  Server, 
  Shield, 
  TestTube,
  Save,
  AlertCircle,
  CheckCircle
} from "lucide-react";
import { toast } from "sonner";
import { backendApi } from "@/lib/api";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface EmailSettingsData {
  email_enabled: boolean;
  email_require_user_email: boolean;
  email_sender_name: string;
  email_sender_address: string;
  email_request_alert: boolean;
  email_smtp_host: string;
  email_smtp_port: number;
  email_encryption_method: string;
  email_use_starttls: boolean;
  email_allow_self_signed: boolean;
  email_smtp_username: string;
  email_smtp_password: string;
  email_pgp_private_key?: string;
  email_pgp_password?: string;
}

export default function EmailSettings() {
  const queryClient = useQueryClient();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [formData, setFormData] = useState<EmailSettingsData>({
    email_enabled: false,
    email_require_user_email: false,
    email_sender_name: "Serra",
    email_sender_address: "",
    email_request_alert: false,
    email_smtp_host: "",
    email_smtp_port: 587,
    email_encryption_method: "starttls",
    email_use_starttls: true,
    email_allow_self_signed: false,
    email_smtp_username: "",
    email_smtp_password: "",
    email_pgp_private_key: "",
    email_pgp_password: "",
  });

  // Fetch current settings
  const { data: settings, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: backendApi.getSystemSettings,
    onSuccess: (data) => {
      const emailSettings = {
        email_enabled: data.email_enabled === "true",
        email_require_user_email: data.email_require_user_email === "true",
        email_sender_name: data.email_sender_name || "Serra",
        email_sender_address: data.email_sender_address || "",
        email_request_alert: data.email_request_alert === "true",
        email_smtp_host: data.email_smtp_host || "",
        email_smtp_port: parseInt(data.email_smtp_port) || 587,
        email_encryption_method: data.email_encryption_method || "starttls",
        email_use_starttls: data.email_use_starttls === "true",
        email_allow_self_signed: data.email_allow_self_signed === "true",
        email_smtp_username: data.email_smtp_username || "",
        email_smtp_password: data.email_smtp_password || "",
        email_pgp_private_key: data.email_pgp_private_key || "",
        email_pgp_password: data.email_pgp_password || "",
      };
      setFormData(emailSettings);
    },
  });

  // Save settings mutation
  const saveSettingsMutation = useMutation({
    mutationFn: async (settings: Record<string, string>) => {
      const results = await Promise.all(
        Object.entries(settings).map(([key, value]) =>
          backendApi.updateSystemSetting(key, value)
        )
      );
      return results;
    },
    onSuccess: () => {
      toast.success("Email settings saved successfully");
      setHasUnsavedChanges(false);
      queryClient.invalidateQueries({ queryKey: ["settings"] });
    },
    onError: (error: any) => {
      toast.error("Failed to save email settings", {
        description: error.response?.data?.error?.message || "Please try again.",
      });
    },
  });

  const handleInputChange = (key: keyof EmailSettingsData, value: any) => {
    setFormData(prev => ({ ...prev, [key]: value }));
    setHasUnsavedChanges(true);
  };

  const handleSave = () => {
    const settingsToSave: Record<string, string> = {
      email_enabled: formData.email_enabled.toString(),
      email_require_user_email: formData.email_require_user_email.toString(),
      email_sender_name: formData.email_sender_name,
      email_sender_address: formData.email_sender_address,
      email_request_alert: formData.email_request_alert.toString(),
      email_smtp_host: formData.email_smtp_host,
      email_smtp_port: formData.email_smtp_port.toString(),
      email_encryption_method: formData.email_encryption_method,
      email_use_starttls: formData.email_use_starttls.toString(),
      email_allow_self_signed: formData.email_allow_self_signed.toString(),
      email_smtp_username: formData.email_smtp_username,
      email_smtp_password: formData.email_smtp_password,
    };

    if (formData.email_pgp_private_key) {
      settingsToSave.email_pgp_private_key = formData.email_pgp_private_key;
    }
    if (formData.email_pgp_password) {
      settingsToSave.email_pgp_password = formData.email_pgp_password;
    }

    saveSettingsMutation.mutate(settingsToSave);
  };

  const handleTestEmail = async () => {
    try {
      // TODO: Implement backend endpoint for test email
      // For now, just validate settings locally
      if (!formData.email_smtp_host || !formData.email_sender_address) {
        toast.error("Please configure SMTP host and sender address first");
        return;
      }
      
      toast.info("Test email functionality will be implemented in a future update");
    } catch (error) {
      toast.error("Failed to send test email");
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            Email Settings
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="space-y-2">
                <div className="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
                <div className="h-10 w-full bg-gray-200 rounded animate-pulse"></div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Mail className="h-5 w-5" />
          Email Settings
        </CardTitle>
        <CardDescription>
          Configure email settings for automatic invitation emails and notifications. If disabled, invitation links can still be copied and shared manually from the invitations page.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Enable Email */}
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="email_enabled">Enable Email Service</Label>
            <p className="text-sm text-gray-600">
              Enable email functionality for invitations and notifications
            </p>
          </div>
          <Switch
            id="email_enabled"
            checked={formData.email_enabled}
            onCheckedChange={(checked) => handleInputChange('email_enabled', checked)}
          />
        </div>

        {formData.email_enabled && (
          <>
            <Separator />

            {/* Sender Information */}
            <div className="space-y-4">
              <h3 className="font-medium text-sm">Sender Information</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="email_sender_name">Sender Name</Label>
                  <Input
                    id="email_sender_name"
                    value={formData.email_sender_name}
                    onChange={(e) => handleInputChange('email_sender_name', e.target.value)}
                    placeholder="Serra"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email_sender_address">Sender Email Address</Label>
                  <Input
                    id="email_sender_address"
                    type="email"
                    value={formData.email_sender_address}
                    onChange={(e) => handleInputChange('email_sender_address', e.target.value)}
                    placeholder="noreply@yourdomain.com"
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* SMTP Configuration */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <Server className="h-4 w-4" />
                <h3 className="font-medium text-sm">SMTP Configuration</h3>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="email_smtp_host">SMTP Host</Label>
                  <Input
                    id="email_smtp_host"
                    value={formData.email_smtp_host}
                    onChange={(e) => handleInputChange('email_smtp_host', e.target.value)}
                    placeholder="smtp.gmail.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email_smtp_port">SMTP Port</Label>
                  <Input
                    id="email_smtp_port"
                    type="number"
                    value={formData.email_smtp_port}
                    onChange={(e) => handleInputChange('email_smtp_port', parseInt(e.target.value))}
                    placeholder="587"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="email_smtp_username">SMTP Username</Label>
                  <Input
                    id="email_smtp_username"
                    value={formData.email_smtp_username}
                    onChange={(e) => handleInputChange('email_smtp_username', e.target.value)}
                    placeholder="your-email@domain.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email_smtp_password">SMTP Password</Label>
                  <Input
                    id="email_smtp_password"
                    type="password"
                    value={formData.email_smtp_password}
                    onChange={(e) => handleInputChange('email_smtp_password', e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* Security Settings */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <Shield className="h-4 w-4" />
                <h3 className="font-medium text-sm">Security Settings</h3>
              </div>

              <div className="space-y-2">
                <Label htmlFor="email_encryption_method">Encryption Method</Label>
                <Select
                  value={formData.email_encryption_method}
                  onValueChange={(value) => handleInputChange('email_encryption_method', value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="starttls">STARTTLS (Recommended)</SelectItem>
                    <SelectItem value="implicit_tls">Implicit TLS/SSL</SelectItem>
                    <SelectItem value="none">None (Not Recommended)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="email_allow_self_signed">Allow Self-Signed Certificates</Label>
                  <p className="text-sm text-gray-600">
                    Only enable if using self-signed certificates (not recommended for production)
                  </p>
                </div>
                <Switch
                  id="email_allow_self_signed"
                  checked={formData.email_allow_self_signed}
                  onCheckedChange={(checked) => handleInputChange('email_allow_self_signed', checked)}
                />
              </div>
            </div>

            <Separator />

            {/* Notification Settings */}
            <div className="space-y-4">
              <h3 className="font-medium text-sm">Notification Settings</h3>
              
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="email_require_user_email">Require User Email</Label>
                  <p className="text-sm text-gray-600">
                    Require email addresses for all user accounts
                  </p>
                </div>
                <Switch
                  id="email_require_user_email"
                  checked={formData.email_require_user_email}
                  onCheckedChange={(checked) => handleInputChange('email_require_user_email', checked)}
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="email_request_alert">Request Alerts</Label>
                  <p className="text-sm text-gray-600">
                    Send email notifications for new requests
                  </p>
                </div>
                <Switch
                  id="email_request_alert"
                  checked={formData.email_request_alert}
                  onCheckedChange={(checked) => handleInputChange('email_request_alert', checked)}
                />
              </div>
            </div>

            {/* Common Provider Examples */}
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="text-sm">
                <strong>Popular SMTP Providers:</strong><br />
                • Gmail: smtp.gmail.com:587 (STARTTLS)<br />
                • ImprovMX: smtp.improvmx.com:587 (STARTTLS)<br />
                • Outlook: smtp-mail.outlook.com:587 (STARTTLS)<br />
                • SendGrid: smtp.sendgrid.net:587 (STARTTLS)
              </AlertDescription>
            </Alert>
          </>
        )}

        {/* Action Buttons */}
        <div className="flex items-center justify-between pt-4">
          <div className="flex items-center space-x-2">
            {formData.email_enabled && (
              <Button
                variant="outline"
                onClick={handleTestEmail}
                disabled={!formData.email_smtp_host || !formData.email_sender_address}
              >
                <TestTube className="mr-2 h-4 w-4" />
                Send Test Email
              </Button>
            )}
          </div>
          
          <div className="flex items-center space-x-2">
            {hasUnsavedChanges && (
              <div className="flex items-center text-sm text-orange-600">
                <AlertCircle className="mr-1 h-4 w-4" />
                Unsaved changes
              </div>
            )}
            <Button
              onClick={handleSave}
              disabled={saveSettingsMutation.isPending || !hasUnsavedChanges}
            >
              {saveSettingsMutation.isPending ? (
                <>
                  <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-transparent border-t-current" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Settings
                </>
              )}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}