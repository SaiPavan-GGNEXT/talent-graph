import { useCallback, useEffect, useRef, useState } from "react";

// Tiny data-fetching hook: tracks loading/error/data and ignores stale
// responses when the inputs change mid-flight.
export function useApi<T>(fn: () => Promise<T>, deps: unknown[], enabled = true) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<unknown>(null);
  const seq = useRef(0);

  const run = useCallback(() => {
    if (!enabled) {
      setData(null);
      setLoading(false);
      setError(null);
      return;
    }
    const id = ++seq.current;
    setLoading(true);
    setError(null);
    fn().then(
      (result) => {
        if (seq.current === id) {
          setData(result);
          setLoading(false);
        }
      },
      (err) => {
        if (seq.current === id) {
          setError(err);
          setLoading(false);
        }
      },
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, enabled]);

  useEffect(run, [run]);

  return { data, loading, error, retry: run };
}
