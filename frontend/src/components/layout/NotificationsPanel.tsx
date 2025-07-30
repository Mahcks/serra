import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Bell, X, Check, Trash2 } from 'lucide-react';
import { backendApi } from '@/lib/api';
import { toast } from 'sonner';
import { useDateTime } from '@/lib/date-format';

interface Notification {
  id: string;
  title: string;
  message: string;
  type: string;
  priority: string;
  read_at: string | null;
  created_at: string;
}

interface NotificationsPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function NotificationsPanel({ isOpen, onClose }: NotificationsPanelProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);
  const { formatRelativeTime } = useDateTime();

  useEffect(() => {
    if (isOpen) {
      loadNotifications();
    }
  }, [isOpen]);

  const loadNotifications = async () => {
    setLoading(true);
    try {
      const data = await backendApi.getNotifications();
      console.log('Notifications API response:', data);
      // Handle both array format and response object format
      const notificationsArray = Array.isArray(data) ? data : (data?.notifications || []);
      setNotifications(notificationsArray);
    } catch (error) {
      console.error('Failed to load notifications:', error);
      toast.error('Failed to load notifications');
    } finally {
      setLoading(false);
    }
  };

  const handleMarkAsRead = async (id: string) => {
    try {
      await backendApi.markNotificationAsRead(id);
      setNotifications(prev => 
        prev.map(n => n.id === id ? { ...n, read_at: new Date().toISOString() } : n)
      );
    } catch (error) {
      console.error('Failed to mark as read:', error);
      toast.error('Failed to mark notification as read');
    }
  };

  const handleDismiss = async (id: string) => {
    try {
      await backendApi.dismissNotification(id);
      setNotifications(prev => prev.filter(n => n.id !== id));
      toast.success('Notification dismissed');
    } catch (error) {
      console.error('Failed to dismiss notification:', error);
      toast.error('Failed to dismiss notification');
    }
  };

  const formatDate = (dateString: string) => {
    return formatRelativeTime(dateString);
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent': return 'bg-red-500';
      case 'high': return 'bg-orange-500';
      case 'normal': return 'bg-blue-500';
      case 'low': return 'bg-gray-500';
      default: return 'bg-blue-500';
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/20" onClick={onClose}>
      <div 
        className="fixed right-4 top-16 w-96 bg-background border rounded-lg shadow-lg"
        onClick={e => e.stopPropagation()}
      >
        <Card className="border-0">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg flex items-center gap-2">
                <Bell className="h-5 w-5" />
                Notifications
              </CardTitle>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="h-8 w-8 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          
          <CardContent className="p-0">
            <ScrollArea className="h-96">
              {loading ? (
                <div className="p-4 text-center text-muted-foreground">
                  Loading notifications...
                </div>
              ) : notifications.length === 0 ? (
                <div className="p-4 text-center text-muted-foreground">
                  No notifications
                </div>
              ) : (
                <div className="space-y-0">
                  {notifications.map((notification, index) => (
                    <div key={notification.id}>
                      <div className={`p-4 hover:bg-muted/50 ${!notification.read_at ? 'bg-blue-50/50 dark:bg-blue-950/20' : ''}`}>
                        <div className="flex items-start gap-3">
                          <div className={`w-2 h-2 rounded-full mt-2 ${getPriorityColor(notification.priority)}`} />
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-2">
                              <div className="flex-1">
                                <h4 className="font-medium text-sm">{notification.title}</h4>
                                <p className="text-sm text-muted-foreground mt-1">
                                  {notification.message}
                                </p>
                                <div className="flex items-center gap-2 mt-2">
                                  <span className="text-xs text-muted-foreground">
                                    {formatDate(notification.created_at)}
                                  </span>
                                  <Badge variant="outline" className="text-xs">
                                    {notification.type}
                                  </Badge>
                                </div>
                              </div>
                              <div className="flex gap-1">
                                {!notification.read_at && (
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => handleMarkAsRead(notification.id)}
                                    className="h-8 w-8 p-0"
                                    title="Mark as read"
                                  >
                                    <Check className="h-3 w-3" />
                                  </Button>
                                )}
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleDismiss(notification.id)}
                                  className="h-8 w-8 p-0"
                                  title="Dismiss"
                                >
                                  <Trash2 className="h-3 w-3" />
                                </Button>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                      {index < notifications.length - 1 && <Separator />}
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}