export interface ListQueryResult<T> {
  data: T[];
  total: number;
  isLoading: boolean;
  isFetching: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
}

export interface DetailQueryResult<T> {
  data: T | null;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
}