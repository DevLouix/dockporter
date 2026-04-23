import { useState, useEffect, useRef } from 'react';
import type { MigrationEvent, HostConfig } from '../types/docker';

export function useMigrationWebSocket(host: HostConfig) {
  const [events, setEvents] = useState<Record<string, MigrationEvent>>({});
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!host.ip || !host.token) return;

    // 🛡️ ENTERPRISE GUARD 2: Secure WebSocket Handshake
    let cleanIp = host.ip.replace(/^(https?:\/\/|wss?:\/\/)/, '').replace(/\/$/, '');
    const isSecureContext = window.location.protocol === 'https:';
    const isLocalhost = cleanIp.startsWith('localhost') || cleanIp.startsWith('127.0.0.1');
    
    const protocol = (isSecureContext && !isLocalhost) ? 'wss' : 'ws';

    const socket = new WebSocket(`${protocol}://${cleanIp}/ws?token=${host.token}`);
    ws.current = socket;

    socket.onmessage = (event) => {
      const data: MigrationEvent = JSON.parse(event.data);
      setEvents((prev) => ({ ...prev, [data.container_id]: data }));

      if (data.status === 'Success' || data.status === 'Failed') {
        setTimeout(() => {
          setEvents(prev => {
            const newState = { ...prev };
            delete newState[data.container_id];
            return newState;
          });
        }, 8000);
      }
    };

    return () => socket.close();
  }, [host.ip, host.token]);

  return events;
}