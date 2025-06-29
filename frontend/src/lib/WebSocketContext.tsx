import type { Message } from "@/types";
import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
} from "react";

type WebSocketContextType = {
  ws: WebSocket | null;
  connected: boolean;
  isConnected: boolean;
  send: (data: any) => void;
  subscribe: (handler: (msg: Message) => void) => () => void;
};

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const ws = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const handlers = useRef<((msg: Message) => void)[]>([]);

  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  // TODO: Change this to use config maybe
  const wsUrl = `${wsProtocol}://${window.location.host}/v1/ws`;

  useEffect(() => {
    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => setConnected(true);
    ws.current.onclose = () => setConnected(false);
    ws.current.onerror = () => setConnected(false);

    ws.current.onmessage = (event) => {
      let data: Message;
      try {
        data = JSON.parse(event.data);
        console.log(data)
      } catch {
        return;
      }
      handlers.current.forEach((h) => h(data));
    };

    return () => {
      ws.current?.close();
    };
  }, [wsUrl]);

  const send = useCallback((data: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data));
    }
  }, []);

  // Subscribe returns an unsubscribe function
  const subscribe = useCallback((handler: (msg: any) => void) => {
    handlers.current.push(handler);
    return () => {
      handlers.current = handlers.current.filter((h) => h !== handler);
    };
  }, []);

  return (
    <WebSocketContext.Provider
      value={{ ws: ws.current, connected, isConnected: connected, send, subscribe }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};

export function useWebSocketContext() {
  const ctx = useContext(WebSocketContext);
  if (!ctx)
    throw new Error(
      "useWebSocketContext must be used within a WebSocketProvider"
    );
  return ctx;
}
