/**
 * Creates a debounced version of the given function.
 * The debounced function delays invoking `fn` until `ms` milliseconds
 * have elapsed since the last invocation.
 */
export function debounce<T extends (...args: any[]) => void>(
  fn: T,
  ms: number,
): T & { cancel: () => void } {
  let timer: ReturnType<typeof setTimeout> | undefined;

  const debounced = (...args: Parameters<T>) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), ms);
  };

  debounced.cancel = () => {
    clearTimeout(timer);
    timer = undefined;
  };

  return debounced as T & { cancel: () => void };
}
