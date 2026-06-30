/**
 * Compatibility layer — delegates to settingsManager.
 * Existing consumers can keep importing from this file.
 * New code should import from $lib/services/settingsManager.svelte directly.
 */
import { settingsManager } from "$lib/services/settingsManager.svelte";
import type { UserSettings } from "$lib/api/settings";

export {
  DIRECTORY_UNLOCK_TTL_OPTIONS,
  type DirectoryUnlockTtlHours,
} from "$lib/services/settingsManager.svelte";

// --- Getters (read from settingsManager's $state) ---

export function getShowSystemDirs() {
  return settingsManager.showSystemDirs;
}

export function getUploadConcurrency() {
  return settingsManager.uploadConcurrency;
}

export function getDuplicateStrategy() {
  return settingsManager.duplicateStrategy;
}

export function getDirectoryUnlockTtlHours() {
  return settingsManager.directoryUnlockTtlHours;
}

// --- Setters (delegate to settingsManager.update) ---

export function setShowSystemDirs(value: boolean) {
  void settingsManager.update("showSystemDirs", value);
}

export function setUploadConcurrency(value: number) {
  void settingsManager.update("uploadConcurrency", value);
}

export function setDuplicateStrategy(value: string) {
  void settingsManager.update("duplicateStrategy", value);
}

export function setDirectoryUnlockTtlHours(value: number) {
  void settingsManager.update("directoryUnlockTtlHours", value);
}

// --- Lifecycle ---

export async function loadPreferencesFromServer() {
  await settingsManager.load();
  return settingsManager.snapshot;
}

export function exportPreferences(): UserSettings {
  return settingsManager.snapshot;
}

export function getThemePreference() {
  return settingsManager.theme;
}

export function setThemePreference(value: UserSettings["theme"]) {
  void settingsManager.update("theme", value);
}

export async function importPreferences(settings: Partial<UserSettings>) {
  await settingsManager.importJSON(JSON.stringify(settings));
  return settingsManager.snapshot;
}
