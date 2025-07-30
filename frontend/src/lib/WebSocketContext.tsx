import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from "react";
import { 
  type Message, 
  OpcodeHello, 
  OpcodeHeartbeat, 
  OpcodeReconnect, 
  OpcodeError,
  OpcodeDispatch,
  OpcodeDownloadProgress,
  OpcodeDownloadProgressBatch,
  OpcodeSystemStatus,
  OpcodeUserActivity,
  OpcodeNotification,
  type HelloPayload,
  type ErrorPayload,
  type DownloadProgressBatchPayload,
  type DriveStatsPayload,
  type SystemStatusPayload,
  type UserActivityPayload
} from "@/types";
import { proactiveTokenRefresh } from "@/lib/api";

// Connection states
export enum WebSocketState {
  CONNECTING = 'connecting',
  CONNECTED = 'connected', 
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error',
  CLOSED = 'closed'
}

// WebSocket events
export enum WebSocketEvent {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected', 
  RECONNECTING = 'reconnecting',
  ERROR = 'error',
  MESSAGE = 'message',
  HELLO = 'hello',
  HEARTBEAT = 'heartbeat'
}

// Typed event handlers
export interface WebSocketEventHandlers {
  [WebSocketEvent.CONNECTING]: () => void;
  [WebSocketEvent.CONNECTED]: (payload: HelloPayload) => void;
  [WebSocketEvent.DISCONNECTED]: (reason?: string) => void;
  [WebSocketEvent.RECONNECTING]: (attempt: number) => void;
  [WebSocketEvent.ERROR]: (error: Error | ErrorPayload) => void;
  [WebSocketEvent.MESSAGE]: (message: Message) => void;
  [WebSocketEvent.HELLO]: (payload: HelloPayload) => void;
  [WebSocketEvent.HEARTBEAT]: () => void;
}

// Message type handlers
export interface MessageTypeHandlers {
  downloadProgressBatch: (payload: DownloadProgressBatchPayload) => void;
  driveStats: (payload: DriveStatsPayload) => void;
  systemStatus: (payload: SystemStatusPayload) => void;
  userActivity: (payload: UserActivityPayload) => void;
  [key: string]: (payload: any) => void;
}

interface QueuedMessage {
  data: any;
  timestamp: number;
  retries: number;
}

interface ConnectionInfo {
  state: WebSocketState;
  url: string;
  connectedAt?: Date;
  disconnectedAt?: Date;
  reconnectAttempts: number;
  lastHeartbeat?: Date;
  serverInfo?: HelloPayload;
  latency?: number;
}

interface WebSocketContextType {
  // Connection state
  connectionInfo: ConnectionInfo;
  isConnected: boolean;
  isConnecting: boolean;
  isReconnecting: boolean;
  
  // Core functionality
  send: (data: any) => boolean;
  disconnect: () => void;
  reconnect: () => void;
  
  // Event system
  addEventListener: <T extends WebSocketEvent>(
    event: T, 
    handler: WebSocketEventHandlers[T]
  ) => () => void;
  
  // Message type subscriptions
  subscribeToMessageType: <T extends string>(
    type: T,
    handler: (payload: any) => void
  ) => () => void;
  
  // Legacy compatibility
  subscribe: (handler: (msg: Message) => void) => () => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

interface WebSocketProviderProps {
  children: React.ReactNode;
  url?: string;
  autoReconnect?: boolean;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  messageQueueSize?: number;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
  url,
  autoReconnect = true,
  reconnectInterval = 1000,
  maxReconnectAttempts = 10,
  heartbeatInterval = 45000,
  messageQueueSize = 100,
}) => {
  // Connection management
  const ws = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<NodeJS.Timeout | null>(null);
  const heartbeatTimer = useRef<NodeJS.Timeout | null>(null);
  const heartbeatResponseTimer = useRef<NodeJS.Timeout | null>(null);
  const lastHeartbeatSent = useRef<number>(0);
  
  // State
  const [connectionInfo, setConnectionInfo] = useState<ConnectionInfo>(() => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    // Use direct backend connection for WebSocket (bypass Vite proxy)
    const wsUrl = url || `${wsProtocol}://localhost:9090/v1/ws`;
    
    return {
      state: WebSocketState.DISCONNECTED,
      url: wsUrl,
      reconnectAttempts: 0,
    };
  });
  
  // Event handlers and message queue
  const eventHandlers = useRef<Map<WebSocketEvent, Set<Function>>>(new Map());
  const messageTypeHandlers = useRef<Map<string, Set<Function>>>(new Map());
  const messageQueue = useRef<QueuedMessage[]>([]);
  const isUnmounting = useRef(false);
  const isInitialized = useRef(false);

  // Helper: Add event listener
  const addEventListener = useCallback(<T extends WebSocketEvent>(
    event: T,
    handler: WebSocketEventHandlers[T]
  ) => {
    if (!eventHandlers.current.has(event)) {
      eventHandlers.current.set(event, new Set());
    }
    eventHandlers.current.get(event)!.add(handler);
    
    return () => {
      eventHandlers.current.get(event)?.delete(handler);
    };
  }, []);

  // Helper: Emit event
  const emitEvent = useCallback((event: WebSocketEvent, ...args: any[]) => {
    const handlers = eventHandlers.current.get(event);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(...args);
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${event}:`, error);
        }
      });
    }
  }, []);

  // Helper: Add message type handler
  const subscribeToMessageType = useCallback(<T extends string>(
    type: T,
    handler: (payload: any) => void
  ) => {
    if (!messageTypeHandlers.current.has(type)) {
      messageTypeHandlers.current.set(type, new Set());
    }
    messageTypeHandlers.current.get(type)!.add(handler);
    
    return () => {
      messageTypeHandlers.current.get(type)?.delete(handler);
    };
  }, []);

  // Helper: Process message by type
  const processMessageByType = useCallback((message: Message) => {
    let messageType: string = 'unknown';
    
    // Map opcodes to message types using the typed constants
    if (message.op === OpcodeDownloadProgress) {
      messageType = 'downloadProgress';
    } else if (message.op === OpcodeDownloadProgressBatch) {
      messageType = 'downloadProgressBatch';
    } else if (message.op === OpcodeSystemStatus) {
      messageType = 'systemStatus';
    } else if (message.op === OpcodeUserActivity) {
      messageType = 'userActivity';
    } else {
      messageType = `opcode_${message.op}`;
    }
    
    const handlers = messageTypeHandlers.current.get(messageType);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message.d);
        } catch (error) {
          console.error(`Error in message type handler for ${messageType}:`, error);
        }
      });
    }
  }, []);

  // Helper: Update connection state
  const updateConnectionState = useCallback((newState: WebSocketState, extra?: Partial<ConnectionInfo>) => {
    if (newState === WebSocketState.DISCONNECTED) {
      console.log("🚨 Connection state being set to DISCONNECTED:", {
        newState,
        extra,
        stack: new Error().stack
      });
    }
    setConnectionInfo(prev => ({
      ...prev,
      state: newState,
      ...extra,
      ...(newState === WebSocketState.CONNECTED ? { connectedAt: new Date() } : {}),
      ...(newState === WebSocketState.DISCONNECTED ? { disconnectedAt: new Date() } : {}),
    }));
  }, []);

  // Helper: Clear timers
  const clearTimers = useCallback(() => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
    if (heartbeatTimer.current) {
      clearInterval(heartbeatTimer.current);
      heartbeatTimer.current = null;
    }
    if (heartbeatResponseTimer.current) {
      clearTimeout(heartbeatResponseTimer.current);
      heartbeatResponseTimer.current = null;
    }
  }, []);

  // Helper: Start heartbeat
  const startHeartbeat = useCallback(() => {
    if (heartbeatTimer.current) return;
    
    heartbeatTimer.current = setInterval(() => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        const heartbeatMessage = {
          op: OpcodeHeartbeat,
          t: Date.now(),
          d: null,
        };
        
        lastHeartbeatSent.current = Date.now();
        ws.current.send(JSON.stringify(heartbeatMessage));
        
        // Set timeout for heartbeat response
        heartbeatResponseTimer.current = setTimeout(() => {
          console.warn('Heartbeat response timeout, connection may be stale');
          // Don't immediately disconnect, but log the issue
        }, 15000); // 15 second timeout for heartbeat response
        
        emitEvent(WebSocketEvent.HEARTBEAT);
      }
    }, heartbeatInterval);
  }, [heartbeatInterval, emitEvent]);

  // Helper: Send queued messages
  const sendQueuedMessages = useCallback(() => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) return;
    
    const now = Date.now();
    const messagesToSend = messageQueue.current.splice(0); // Remove all queued messages
    
    messagesToSend.forEach(queuedMessage => {
      // Skip messages that are too old (older than 5 minutes)
      if (now - queuedMessage.timestamp > 5 * 60 * 1000) {
        console.warn('Discarding old queued message');
        return;
      }
      
      try {
        ws.current!.send(JSON.stringify(queuedMessage.data));
      } catch (error) {
        console.error('Failed to send queued message:', error);
        // Re-queue if under retry limit
        if (queuedMessage.retries < 3) {
          messageQueue.current.push({
            ...queuedMessage,
            retries: queuedMessage.retries + 1
          });
        }
      }
    });
  }, []);

  // Helper: Calculate reconnect delay with exponential backoff
  const getReconnectDelay = useCallback((attempt: number) => {
    const baseDelay = reconnectInterval;
    const maxDelay = 30000; // 30 seconds max
    const delay = Math.min(baseDelay * Math.pow(2, attempt), maxDelay);
    // Add jitter (±25%)
    const jitter = delay * 0.25 * (Math.random() - 0.5);
    return delay + jitter;
  }, [reconnectInterval]);

  // Core: Connect function
  const connect = useCallback(async () => {
    if (isUnmounting.current) return;
    
    clearTimers();
    
    if (ws.current?.readyState === WebSocket.OPEN) {
      console.log("🔌 WebSocket already connected");
      return; // Already connected
    }
    
    // Refresh token before connecting to ensure we have a valid token
    try {
      console.log("🔄 Refreshing token before WebSocket connection...");
      await proactiveTokenRefresh();
    } catch (error) {
      console.warn("⚠️ Token refresh failed before WebSocket connection:", error);
      // If token refresh fails, don't attempt to connect
      updateConnectionState(WebSocketState.ERROR);
      emitEvent(WebSocketEvent.ERROR, new Error('Authentication failed: Unable to refresh token'));
      return;
    }
    
    console.log("🔌 Starting WebSocket connection to:", connectionInfo.url);
    updateConnectionState(WebSocketState.CONNECTING);
    emitEvent(WebSocketEvent.CONNECTING);
    
    try {
      ws.current = new WebSocket(connectionInfo.url);
      
      ws.current.onopen = () => {
        if (isUnmounting.current) return;
        
        console.log("✅ WebSocket connection opened successfully");
        updateConnectionState(WebSocketState.CONNECTED, {
          reconnectAttempts: 0,
          connectedAt: new Date(),
        });
        
        // Don't start client heartbeat - server will send heartbeats and we'll respond
        sendQueuedMessages();
      };
      
      ws.current.onmessage = (event) => {
        if (isUnmounting.current) return;
        
        // Add debug logging for ALL WebSocket messages
        console.log("🔌 WebSocket message received:", {
          raw: event.data,
          timestamp: new Date().toISOString()
        });
        
        try {
          const message: Message = JSON.parse(event.data);
          
          console.log("📦 Parsed WebSocket message:", {
            opcode: message.op,
            opcodeType: message.op === 1 ? 'Hello' : 
                       message.op === 2 ? 'Heartbeat' :
                       message.op === 12 ? 'DownloadProgressBatch' : 
                       `Unknown(${message.op})`,
            timestamp: message.t,
            timeSinceLastHeartbeat: message.op === 2 ? (Date.now() - (lastHeartbeatSent.current || 0)) : null,
            dataKeys: message.d ? Object.keys(message.d) : null,
            data: message.d
          });
          
          // Handle specific opcodes
          switch (message.op) {
            case OpcodeHello:
              const helloPayload = message.d as HelloPayload;
              updateConnectionState(WebSocketState.CONNECTED, {
                serverInfo: helloPayload,
                lastHeartbeat: new Date(),
              });
              emitEvent(WebSocketEvent.CONNECTED, helloPayload);
              emitEvent(WebSocketEvent.HELLO, helloPayload);
              break;
              
            case OpcodeHeartbeat:
              try {
                console.log("💓 Processing heartbeat from server, current readyState:", ws.current?.readyState);
                
                // Calculate latency
                if (lastHeartbeatSent.current > 0) {
                  const latency = Date.now() - lastHeartbeatSent.current;
                  updateConnectionState(connectionInfo.state, { latency });
                }
                
                // Clear heartbeat response timeout
                if (heartbeatResponseTimer.current) {
                  clearTimeout(heartbeatResponseTimer.current);
                  heartbeatResponseTimer.current = null;
                }
                
                // Send heartbeat response back to server (server expects this)
                if (ws.current && ws.current.readyState === WebSocket.OPEN) {
                  const heartbeatResponse = {
                    op: OpcodeHeartbeat,
                    t: Date.now(),
                    d: null,
                  };
                  console.log("💓 Sending heartbeat response to server:", heartbeatResponse);
                  ws.current.send(JSON.stringify(heartbeatResponse));
                  console.log("✅ Heartbeat response sent successfully, readyState:", ws.current.readyState);
                } else {
                  console.warn("⚠️ Cannot send heartbeat response - WebSocket not open:", ws.current?.readyState);
                }
                
                // Update last heartbeat without changing connection state
                setConnectionInfo(prev => ({
                  ...prev,
                  lastHeartbeat: new Date(),
                }));
                emitEvent(WebSocketEvent.HEARTBEAT);
                console.log("✅ Heartbeat processing completed successfully");
              } catch (error) {
                console.error("❌ Error processing heartbeat:", error);
                throw error;
              }
              break;
              
            case OpcodeReconnect:
              console.log('🔄 Server requested reconnection - this should not happen during normal operation!', {
                timestamp: new Date().toISOString(),
                messageData: message.d
              });
              disconnect();
              setTimeout(() => reconnect(), 1000);
              break;
              
            case OpcodeError:
              { const errorPayload = message.d as ErrorPayload;
              console.error('WebSocket error from server:', errorPayload);
              
              // Check if it's an authentication error
              if (errorPayload.message && (
                errorPayload.message.includes('auth') || 
                errorPayload.message.includes('token') ||
                errorPayload.message.includes('Invalid auth token') ||
                errorPayload.message.includes('Missing auth token') ||
                errorPayload.message.includes('Auth token expired')
              )) {
                console.log('🔐 Authentication error detected, will attempt token refresh on reconnect');
                // For token expiration, trigger an immediate reconnect with token refresh
                if (errorPayload.message.includes('expired')) {
                  console.log('🔄 Token expired, triggering immediate reconnect with token refresh');
                  disconnect();
                  setTimeout(() => reconnect(), 500);
                }
              }
              
              emitEvent(WebSocketEvent.ERROR, errorPayload);
              break; }
              
            case OpcodeDispatch:
              console.log('Dispatching message:', message.d);
              processMessageByType(message);
              // Also emit generic message event
              emitEvent(WebSocketEvent.MESSAGE, message);
              break;
              
            default:
              processMessageByType(message);
              // Also emit generic message event
              emitEvent(WebSocketEvent.MESSAGE, message);
              break;
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
          emitEvent(WebSocketEvent.ERROR, error as Error);
        }
      };
      
      ws.current.onclose = (event) => {
        console.log("🔴 WebSocket onclose handler called", {
          isUnmounting: isUnmounting.current,
          event: event
        });
        if (isUnmounting.current) return;
        
        clearTimers();
        
        const reason = event.code === 1000 ? 'Normal closure' : 
                      event.code === 1001 ? 'Going away' :
                      event.code === 1006 ? 'Abnormal closure' :
                      event.code === 1008 ? 'Policy violation (likely auth error)' :
                      event.code === 1011 ? 'Unexpected condition (server error)' :
                      `Code ${event.code}: ${event.reason}`;
        
        console.log("❌ WebSocket connection closed:", {
          code: event.code,
          reason: event.reason || 'No reason provided',
          wasClean: event.wasClean,
          readyState: ws.current?.readyState,
          timestamp: new Date().toISOString(),
          parsedReason: reason
        });
        
        // Check if close was due to authentication issues
        const isAuthError = event.code === 1008 || 
                          (event.reason && event.reason.toLowerCase().includes('auth'));
        
        if (isAuthError) {
          console.log('🔐 WebSocket closed due to authentication error, will refresh token on reconnect');
        }
        
        updateConnectionState(WebSocketState.DISCONNECTED, {
          disconnectedAt: new Date(),
        });
        
        emitEvent(WebSocketEvent.DISCONNECTED, reason);
        
        // Auto-reconnect if enabled and not a normal closure
        if (autoReconnect && event.code !== 1000 && connectionInfo.reconnectAttempts < maxReconnectAttempts) {
          const nextAttempt = connectionInfo.reconnectAttempts + 1;
          const delay = getReconnectDelay(nextAttempt);
          
          updateConnectionState(WebSocketState.RECONNECTING, {
            reconnectAttempts: nextAttempt,
          });
          
          emitEvent(WebSocketEvent.RECONNECTING, nextAttempt);
          
          reconnectTimer.current = setTimeout(() => {
            if (!isUnmounting.current) {
              connect();
            }
          }, delay);
        } else if (connectionInfo.reconnectAttempts >= maxReconnectAttempts) {
          console.error('Max reconnection attempts reached');
          updateConnectionState(WebSocketState.ERROR);
        }
      };
      
      ws.current.onerror = (event) => {
        if (isUnmounting.current) return;
        
        console.error('❌ WebSocket error:', {
          event,
          url: connectionInfo.url,
          readyState: ws.current?.readyState,
          timestamp: new Date().toISOString()
        });
        emitEvent(WebSocketEvent.ERROR, new Error('WebSocket connection error'));
      };
      
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      updateConnectionState(WebSocketState.ERROR);
      emitEvent(WebSocketEvent.ERROR, error as Error);
    }
  }, [
    connectionInfo.url, 
    connectionInfo.reconnectAttempts, 
    connectionInfo.state,
    autoReconnect, 
    maxReconnectAttempts, 
    updateConnectionState, 
    emitEvent, 
    startHeartbeat, 
    sendQueuedMessages, 
    getReconnectDelay, 
    clearTimers
  ]);

  // Core: Disconnect function
  const disconnect = useCallback(() => {
    console.log("🔌 Disconnect called, current readyState:", ws.current?.readyState);
    clearTimers();
    
    if (ws.current) {
      // Don't close if we're still connecting - let it finish
      if (ws.current.readyState === WebSocket.CONNECTING) {
        console.log("⏳ WebSocket still connecting, not closing");
        return;
      }
      
      ws.current.close(1000, 'Manual disconnect');
      ws.current = null;
    }
    
    updateConnectionState(WebSocketState.CLOSED);
  }, [clearTimers, updateConnectionState]);

  // Core: Reconnect function
  const reconnect = useCallback(() => {
    console.log("🔄 Manual reconnect called", {
      currentState: connectionInfo.state,
      stack: new Error().stack
    });
    disconnect();
    updateConnectionState(WebSocketState.DISCONNECTED, {
      reconnectAttempts: 0,
    });
    setTimeout(connect, 100);
  }, [disconnect, connect, updateConnectionState, connectionInfo.state]);

  // Core: Send function
  const send = useCallback((data: any): boolean => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      // Queue message if not connected and queue isn't full
      if (messageQueue.current.length < messageQueueSize) {
        messageQueue.current.push({
          data,
          timestamp: Date.now(),
          retries: 0,
        });
        console.log('Message queued (WebSocket not connected)');
        return false;
      } else {
        console.warn('Message queue is full, discarding message');
        return false;
      }
    }
    
    try {
      ws.current.send(JSON.stringify(data));
      return true;
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
      return false;
    }
  }, [messageQueueSize]);

  // Legacy compatibility: subscribe method
  const subscribe = useCallback((handler: (msg: Message) => void) => {
    return addEventListener(WebSocketEvent.MESSAGE, handler);
  }, [addEventListener]);

  // Initialize connection
  useEffect(() => {
    console.log("🔄 WebSocketProvider useEffect - initializing connection", {
      isInitialized: isInitialized.current,
      isUnmounting: isUnmounting.current
    });
    
    // Reset unmounting flag on new mount
    isUnmounting.current = false;
    
    // Prevent React StrictMode double initialization
    if (isInitialized.current) {
      console.log("⏭️ Already initialized, skipping");
      return;
    }
    
    isInitialized.current = true;
    
    // Don't close existing connection if it's already working
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      console.log("🔄 WebSocket already connected, keeping connection");
      // Update connection state to reflect the actual connection status
      updateConnectionState(WebSocketState.CONNECTED, {
        connectedAt: connectionInfo.connectedAt || new Date(),
      });
      return;
    }
    
    // Force fresh connection by disconnecting first
    if (ws.current) {
      console.log("🔄 Closing existing WebSocket connection");
      ws.current.close();
      ws.current = null;
    }
    
    connect();
    
    return () => {
      console.log("🔄 WebSocketProvider cleanup - NOT disconnecting during hot reload");
      // Don't disconnect during hot reload - only mark as unmounting
      isUnmounting.current = true;
      isInitialized.current = false;
    };
  }, []); // Empty dependency array - only run on mount/unmount

  // Cleanup on page unload (not hot reload)
  useEffect(() => {
    const handleBeforeUnload = () => {
      console.log("🔄 Page unload detected, disconnecting WebSocket");
      isUnmounting.current = true;
      disconnect();
    };
    
    window.addEventListener('beforeunload', handleBeforeUnload);
    
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
      clearTimers();
    };
  }, [clearTimers, disconnect]);

  // Update connection info and reconnect when url prop changes
  useEffect(() => {
    if (!url) return;
    setConnectionInfo(prev => ({
      ...prev,
      url,
      reconnectAttempts: 0,
    }));
    // Force a reconnect to the new URL
    disconnect();
    setTimeout(connect, 100);
  }, [url]);

  // Derived state
  const isConnected = connectionInfo.state === WebSocketState.CONNECTED;
  const isConnecting = connectionInfo.state === WebSocketState.CONNECTING;
  const isReconnecting = connectionInfo.state === WebSocketState.RECONNECTING;

  const contextValue: WebSocketContextType = {
    connectionInfo,
    isConnected,
    isConnecting,
    isReconnecting,
    send,
    disconnect,
    reconnect,
    addEventListener,
    subscribeToMessageType,
    subscribe, // Legacy compatibility
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};

export function useWebSocketContext() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider');
  }
  return context;
}

// Convenience hook for specific message types
export function useWebSocketMessage<T extends string>(
  messageType: T,
  handler: (payload: any) => void,
  dependencies: React.DependencyList = []
) {
  const { subscribeToMessageType } = useWebSocketContext();
  
  useEffect(() => {
    return subscribeToMessageType(messageType, handler);
  }, [subscribeToMessageType, messageType, ...dependencies]);
}

// Convenience hook for connection events
export function useWebSocketEvent<T extends WebSocketEvent>(
  event: T,
  handler: WebSocketEventHandlers[T],
  dependencies: React.DependencyList = []
) {
  const { addEventListener } = useWebSocketContext();
  
  useEffect(() => {
    return addEventListener(event, handler);
  }, [addEventListener, event, ...dependencies]);
}