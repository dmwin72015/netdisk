import { browser } from "$app/environment";

export function persistedState<T>(key: string, defaultValue: T) {
  let value = $state<T>(browser ? (localStorage.getItem(key) as T | null) ?? defaultValue : defaultValue);
  return {
    get current() { return value; },
    set current(v: T) {
      value = v;
      if (browser) localStorage.setItem(key, String(v));
    },
  };
}
