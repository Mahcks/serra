import React from 'react';
import { useWebSocketContext, WebSocketState, WebSocketEvent, useWebSocketEvent } from '@/lib/WebSocketContext';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { 
  Wifi, 
  WifiOff, 
  Loader2, 
  AlertCircle, 
  CheckCircle, 
  Clock 
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface WebSocketStatusProps {
  showDetails?: boolean;
  className?: string;
}

export function WebSocketStatus({ showDetails = false, className = '' }: WebSocketStatusProps) {
  const { connectionInfo, isConnected, isConnecting, isReconnecting } = useWebSocketContext();
  const [lastHeartbeat, setLastHeartbeat] = React.useState<Date | null>(null);

  // Track heartbeats
  useWebSocketEvent(WebSocketEvent.HEARTBEAT, () => {
    setLastHeartbeat(new Date());
  }, []);

  const getStatusConfig = () => {
    switch (connectionInfo.state) {
      case WebSocketState.CONNECTED:
        return {
          icon: CheckCircle,
          color: 'bg-green-500',
          textColor: 'text-green-700',
          bgColor: 'bg-green-50 border-green-200',
          label: 'Connected',
          variant: 'default' as const,
        };
      case WebSocketState.CONNECTING:
        return {
          icon: Loader2,
          color: 'bg-blue-500',
          textColor: 'text-blue-700', 
          bgColor: 'bg-blue-50 border-blue-200',
          label: 'Connecting',
          variant: 'secondary' as const,
        };
      case WebSocketState.RECONNECTING:
        return {
          icon: Loader2,
          color: 'bg-yellow-500',
          textColor: 'text-yellow-700',
          bgColor: 'bg-yellow-50 border-yellow-200', 
          label: `Reconnecting (${connectionInfo.reconnectAttempts})`,
          variant: 'outline' as const,
        };
      case WebSocketState.DISCONNECTED:
        return {
          icon: WifiOff,
          color: 'bg-gray-500',
          textColor: 'text-gray-700',
          bgColor: 'bg-gray-50 border-gray-200',
          label: 'Disconnected',
          variant: 'secondary' as const,
        };
      case WebSocketState.ERROR:
        return {
          icon: AlertCircle,
          color: 'bg-red-500', 
          textColor: 'text-red-700',
          bgColor: 'bg-red-50 border-red-200',
          label: 'Error',
          variant: 'destructive' as const,
        };
      case WebSocketState.CLOSED:
        return {
          icon: WifiOff,
          color: 'bg-gray-500',
          textColor: 'text-gray-700', 
          bgColor: 'bg-gray-50 border-gray-200',
          label: 'Closed',
          variant: 'outline' as const,
        };
      default:
        return {
          icon: WifiOff,
          color: 'bg-gray-500',
          textColor: 'text-gray-700',
          bgColor: 'bg-gray-50 border-gray-200', 
          label: 'Unknown',
          variant: 'outline' as const,
        };
    }
  };

  const config = getStatusConfig();
  const Icon = config.icon;

  const getTooltipContent = () => {
    const details = [];
    
    details.push(`Status: ${config.label}`);
    
    if (connectionInfo.connectedAt) {
      details.push(`Connected: ${formatDistanceToNow(connectionInfo.connectedAt, { addSuffix: true })}`);
    }
    
    if (connectionInfo.disconnectedAt && connectionInfo.state === WebSocketState.DISCONNECTED) {
      details.push(`Disconnected: ${formatDistanceToNow(connectionInfo.disconnectedAt, { addSuffix: true })}`);
    }
    
    if (connectionInfo.latency) {
      details.push(`Latency: ${connectionInfo.latency}ms`);
    }
    
    if (lastHeartbeat) {
      details.push(`Last heartbeat: ${formatDistanceToNow(lastHeartbeat, { addSuffix: true })}`);
    }
    
    if (connectionInfo.serverInfo) {
      details.push(`Server: ${connectionInfo.serverInfo.message || 'Unknown'}`);
    }
    
    if (connectionInfo.reconnectAttempts > 0) {
      details.push(`Reconnect attempts: ${connectionInfo.reconnectAttempts}`);
    }

    return details.join('\n');
  };

  if (!showDetails) {
    // Simple status indicator
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={`flex items-center gap-1 ${className}`}>
              <div className={`w-2 h-2 rounded-full ${config.color} flex-shrink-0`} />
              {(isConnecting || isReconnecting) && (
                <Loader2 className="w-3 h-3 animate-spin text-muted-foreground" />
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <div className="whitespace-pre-line">{getTooltipContent()}</div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Detailed status display
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge variant={config.variant} className={`${className} ${config.bgColor} ${config.textColor}`}>
            <Icon 
              className={`w-3 h-3 mr-1 ${
                (isConnecting || isReconnecting) ? 'animate-spin' : ''
              }`} 
            />
            {config.label}
            {connectionInfo.latency && isConnected && (
              <span className="ml-1 text-xs opacity-75">
                ({connectionInfo.latency}ms)
              </span>
            )}
          </Badge>
        </TooltipTrigger>
        <TooltipContent>
          <div className="whitespace-pre-line">{getTooltipContent()}</div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

// Hook to get connection health score (0-100)
export function useWebSocketHealth(): number {
  const { connectionInfo } = useWebSocketContext();
  
  if (connectionInfo.state === WebSocketState.CONNECTED) {
    let score = 100;
    
    // Reduce score based on latency
    if (connectionInfo.latency) {
      if (connectionInfo.latency > 1000) score -= 30;
      else if (connectionInfo.latency > 500) score -= 15;
      else if (connectionInfo.latency > 200) score -= 5;
    }
    
    // Reduce score based on recent reconnects
    if (connectionInfo.reconnectAttempts > 0) {
      score -= Math.min(connectionInfo.reconnectAttempts * 10, 40);
    }
    
    return Math.max(score, 60); // Connected but poor quality
  }
  
  if (connectionInfo.state === WebSocketState.RECONNECTING) {
    return Math.max(50 - (connectionInfo.reconnectAttempts * 5), 10);
  }
  
  return 0; // Disconnected/Error/Closed
}