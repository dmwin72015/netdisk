import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router';

export interface TableUrlState {
  page: number;
  pageSize: number;
  [key: string]: string | number | undefined;
}

export function useTableUrlState<T extends Record<string, unknown>>(defaults: T) {
  const [searchParams, setSearchParams] = useSearchParams();

  const params = useMemo(() => {
    const result = { ...defaults } as Record<string, unknown>;
    for (const [key, value] of searchParams.entries()) {
      if (key in defaults) {
        const defaultVal = defaults[key];
        result[key] = typeof defaultVal === 'number' ? Number(value) : value;
      }
    }
    return result as T & TableUrlState;
  }, [searchParams, defaults]);

  const setParams = useCallback(
    (updates: Partial<T & TableUrlState>) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        Object.entries(updates).forEach(([key, value]) => {
          if (value === undefined || value === '' || value === null) {
            next.delete(key);
          } else {
            next.set(key, String(value));
          }
        });
        return next;
      });
    },
    [setSearchParams],
  );

  const resetParams = useCallback(() => {
    setSearchParams(new URLSearchParams());
  }, [setSearchParams]);

  return { params, setParams, resetParams };
}
