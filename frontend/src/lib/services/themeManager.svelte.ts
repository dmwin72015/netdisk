import { browser } from "$app/environment";
import type { ThemePreference } from "$lib/api/settings";

const LS_THEME = "nd.appearance.theme";
export type ResolvedTheme = "light" | "dark";

export function isThemePreference(value: unknown): value is ThemePreference {
  return value === "system" || value === "light" || value === "dark";
}

export function normalizeThemePreference(value: unknown): ThemePreference {
  return isThemePreference(value) ? value : "system";
}

export function resolveThemePreference(
  preference: ThemePreference,
  systemTheme: ResolvedTheme,
): ResolvedTheme {
  return preference === "system" ? systemTheme : preference;
}

function readSystemTheme(): ResolvedTheme {
  if (!browser) return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function readStoredTheme(): ThemePreference {
  if (!browser) return "system";
  return normalizeThemePreference(localStorage.getItem(LS_THEME));
}

class ThemeManager {
  theme = $state<ThemePreference>(readStoredTheme());
  resolvedTheme = $state<ResolvedTheme>("light");
  private mediaQuery: MediaQueryList | null = null;
  private onSystemChange = () => {
    if (this.theme === "system") this.applyToDocument();
  };

  init() {
    if (!browser) return;
    this.mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    this.mediaQuery.addEventListener("change", this.onSystemChange);
    this.applyToDocument();
  }

  async setTheme(
    theme: ThemePreference,
    options: { persist?: (theme: ThemePreference) => Promise<void> | void } = {},
  ) {
    this.setLocal(theme);
    if (options.persist) await options.persist(this.theme);
  }

  applyFromServer(theme: ThemePreference) {
    this.setLocal(theme);
  }

  private setLocal(theme: ThemePreference) {
    this.theme = normalizeThemePreference(theme);
    if (browser) localStorage.setItem(LS_THEME, this.theme);
    this.applyToDocument();
  }

  private applyToDocument() {
    const resolved = resolveThemePreference(this.theme, readSystemTheme());
    this.resolvedTheme = resolved;
    if (!browser) return;
    document.documentElement.dataset.theme = resolved;
    document.documentElement.style.colorScheme = resolved;
  }
}

export const themeManager = new ThemeManager();
