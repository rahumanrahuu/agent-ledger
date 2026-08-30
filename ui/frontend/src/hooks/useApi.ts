import { useState, useEffect, useCallback, useRef } from 'react';

export type ApiState<T> = { status: 'loading' } | { status: 'ok'; data: T } | { status: 'error'; error: string };

export function useApi<T>(fetcher: () => Promise<T>, deps: unknown[] = []) {
  const [state, setState] = useState<ApiState<T>>({ status: 'loading' });
  const mountedRef = useRef(true);

  const load = useCallback(async () => {
    setState({ status: 'loading' });
    try {
      const data = await fetcher();
      if (mountedRef.current) setState({ status: 'ok', data });
    } catch (err) {
      if (mountedRef.current)
        setState({ status: 'error', error: err instanceof Error ? err.message : String(err) });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  useEffect(() => {
    mountedRef.current = true;
    load();
    return () => { mountedRef.current = false; };
  }, [load]);

  // Live WebSocket update subscription for silent background refresh
  useEffect(() => {
    const handleLiveUpdate = async () => {
      try {
        const data = await fetcher();
        if (mountedRef.current) {
          setState({ status: 'ok', data });
        }
      } catch (err) {
        console.warn('Live update fetch failed:', err);
      }
    };

    window.addEventListener('al-websocket-message', handleLiveUpdate);
    return () => window.removeEventListener('al-websocket-message', handleLiveUpdate);
  }, [fetcher]);

  return { state, reload: load };
}
