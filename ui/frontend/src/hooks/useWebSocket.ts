import { useState, useEffect } from 'react';
import { createWebSocket, type WsStatus } from '../api/client';

export function useWebSocket() {
  const [status, setStatus] = useState<WsStatus>('connecting');

  useEffect(() => {
    const cleanup = createWebSocket(
      (_e) => { /* future: dispatch live events */ },
      setStatus
    );
    return cleanup;
  }, []);

  return status;
}
