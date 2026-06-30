import { describe, expect, it } from "vitest";
import {
  isThemePreference,
  normalizeThemePreference,
  resolveThemePreference,
} from "../../src/lib/services/themeManager.svelte";

describe("theme manager helpers", () => {
  it("recognizes valid preferences", () => {
    expect(isThemePreference("system")).toBe(true);
    expect(isThemePreference("light")).toBe(true);
    expect(isThemePreference("dark")).toBe(true);
    expect(isThemePreference("midnight")).toBe(false);
  });

  it("normalizes invalid values to system", () => {
    expect(normalizeThemePreference("dark")).toBe("dark");
    expect(normalizeThemePreference("light")).toBe("light");
    expect(normalizeThemePreference("system")).toBe("system");
    expect(normalizeThemePreference(null)).toBe("system");
    expect(normalizeThemePreference("bad-value")).toBe("system");
  });

  it("resolves system against the current system mode", () => {
    expect(resolveThemePreference("light", "dark")).toBe("light");
    expect(resolveThemePreference("dark", "light")).toBe("dark");
    expect(resolveThemePreference("system", "dark")).toBe("dark");
    expect(resolveThemePreference("system", "light")).toBe("light");
  });
});
