import { useState, useEffect } from 'react';
import { createWebSocket, type WsStatus } from '../api/client';

export function useWebSocket(onMessage?: (e: MessageEvent) => void) {
  const [status, setStatus] = useState<WsStatus>('connecting');

  useEffect(() => {
    const cleanup = createWebSocket(
      (e) => {
        window.dispatchEvent(new CustomEvent('al-websocket-message', { detail: e.data }));
        if (onMessage) onMessage(e);
      },
      setStatus
    );
    return cleanup;
  }, [onMessage]);

  return status;
}
