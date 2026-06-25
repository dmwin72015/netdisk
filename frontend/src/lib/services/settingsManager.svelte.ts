import { browser } from "$app/environment";
import {
  getUserSettings,
  saveUserSettings,
  type UserSettings,
} from "$lib/api/settings";
import {
  UPLOAD_FILE_CONCURRENCY,
  UPLOAD_FILE_CONCURRENCY_DEFAULT,
} from "$lib/upload-concurrency";

export const DIRECTORY_UNLOCK_TTL_OPTIONS = [1, 2, 6, 24, -1] as const;
export type DirectoryUnlockTtlHours =
  (typeof DIRECTORY_UNLOCK_TTL_OPTIONS)[number];

const LS_SHOW_SYSTEM_DIRS = "nd.files.showSystemDirs";
const LS_UPLOAD_CONCURRENCY = "nd.files.uploadConcurrency";
const LS_DUPLICATE_STRATEGY = "nd.files.duplicateStrategy";
const LS_DIRECTORY_UNLOCK_TTL = "nd.files.directoryUnlockTtlHours";

const DEFAULTS: UserSettings = {
  showSystemDirs: true,
  uploadConcurrency: UPLOAD_FILE_CONCURRENCY_DEFAULT,
  duplicateStrategy: "prompt",
  directoryUnlockTtlHours: 2,
};

function lsGet(key: string): string | null {
  return browser ? localStorage.getItem(key) : null;
}

function lsSet(key: string, value: string) {
  if (browser) localStorage.setItem(key, value);
}

function clampConcurrency(v: number) {
  return Math.max(
    1,
    Math.min(UPLOAD_FILE_CONCURRENCY, v || UPLOAD_FILE_CONCURRENCY_DEFAULT),
  );
}

function normalizeStrategy(v: string) {
  return ["prompt", "overwrite", "keep_both", "skip"].includes(v)
    ? v
    : DEFAULTS.duplicateStrategy;
}

function normalizeTtl(v: number) {
  return DIRECTORY_UNLOCK_TTL_OPTIONS.includes(
    v as DirectoryUnlockTtlHours,
  )
    ? v
    : DEFAULTS.directoryUnlockTtlHours;
}

class SettingsManager {
  showSystemDirs = $state<boolean>(DEFAULTS.showSystemDirs);
  uploadConcurrency = $state<number>(DEFAULTS.uploadConcurrency);
  duplicateStrategy = $state<string>(DEFAULTS.duplicateStrategy);
  directoryUnlockTtlHours = $state<number>(DEFAULTS.directoryUnlockTtlHours);

  private serverLoaded = false;
  private persistTimer: ReturnType<typeof setTimeout> | null = null;

  constructor() {
    // Hydrate from localStorage on construction (browser only)
    const sv = lsGet(LS_SHOW_SYSTEM_DIRS);
    if (sv !== null) this.showSystemDirs = sv !== "false";

    const cv = lsGet(LS_UPLOAD_CONCURRENCY);
    if (cv) this.uploadConcurrency = clampConcurrency(parseInt(cv, 10));

    const dv = lsGet(LS_DUPLICATE_STRATEGY);
    if (dv) this.duplicateStrategy = normalizeStrategy(dv);

    const tv = lsGet(LS_DIRECTORY_UNLOCK_TTL);
    if (tv) this.directoryUnlockTtlHours = normalizeTtl(parseInt(tv, 10));
  }

  /** Load settings from server, merge with local. Call once at app startup. */
  async load(): Promise<void> {
    if (!browser) return;
    try {
      const settings = await getUserSettings();
      this.apply(settings, true);
      this.serverLoaded = true;
    } catch {
      this.serverLoaded = true;
    }
  }

  /** Update a single setting (optimistic). */
  async update<K extends keyof UserSettings>(
    key: K,
    value: UserSettings[K],
  ): Promise<void> {
    const prev = this.snapshot[key];
    this.setLocal(key, value);
    try {
      const saved = await saveUserSettings(this.snapshot);
      this.apply(saved, true);
    } catch {
      this.setLocal(key, prev);
      throw new Error("设置保存失败");
    }
  }

  /** Get current settings as a plain object. */
  get snapshot(): UserSettings {
    return {
      showSystemDirs: this.showSystemDirs,
      uploadConcurrency: this.uploadConcurrency,
      duplicateStrategy: this.duplicateStrategy,
      directoryUnlockTtlHours: this.directoryUnlockTtlHours,
    };
  }

  /** Export settings as JSON string. */
  exportJSON(): string {
    return JSON.stringify(this.snapshot, null, 2);
  }

  /** Import settings from JSON string. */
  async importJSON(json: string): Promise<void> {
    const parsed = JSON.parse(json) as Partial<UserSettings>;
    const merged: UserSettings = { ...this.snapshot, ...parsed };
    this.serverLoaded = true;
    if (this.persistTimer) {
      clearTimeout(this.persistTimer);
      this.persistTimer = null;
    }
    const saved = await saveUserSettings(merged);
    this.apply(saved, true);
  }

  // --- Internal ---

  private apply(settings: Partial<UserSettings>, persistLocal: boolean) {
    if (typeof settings.showSystemDirs === "boolean")
      this.showSystemDirs = settings.showSystemDirs;
    if (settings.uploadConcurrency != null)
      this.uploadConcurrency = clampConcurrency(
        Number(settings.uploadConcurrency),
      );
    if (settings.duplicateStrategy)
      this.duplicateStrategy = normalizeStrategy(
        String(settings.duplicateStrategy),
      );
    if (settings.directoryUnlockTtlHours != null)
      this.directoryUnlockTtlHours = normalizeTtl(
        Number(settings.directoryUnlockTtlHours),
      );

    if (persistLocal) {
      lsSet(LS_SHOW_SYSTEM_DIRS, String(this.showSystemDirs));
      lsSet(LS_UPLOAD_CONCURRENCY, String(this.uploadConcurrency));
      lsSet(LS_DUPLICATE_STRATEGY, this.duplicateStrategy);
      lsSet(LS_DIRECTORY_UNLOCK_TTL, String(this.directoryUnlockTtlHours));
    }
  }

  private setLocal<K extends keyof UserSettings>(key: K, value: UserSettings[K]) {
    switch (key) {
      case "showSystemDirs":
        this.showSystemDirs = value as boolean;
        lsSet(LS_SHOW_SYSTEM_DIRS, String(value));
        break;
      case "uploadConcurrency":
        this.uploadConcurrency = clampConcurrency(value as number);
        lsSet(LS_UPLOAD_CONCURRENCY, String(this.uploadConcurrency));
        break;
      case "duplicateStrategy":
        this.duplicateStrategy = normalizeStrategy(value as string);
        lsSet(LS_DUPLICATE_STRATEGY, this.duplicateStrategy);
        break;
      case "directoryUnlockTtlHours":
        this.directoryUnlockTtlHours = normalizeTtl(value as number);
        lsSet(LS_DIRECTORY_UNLOCK_TTL, String(this.directoryUnlockTtlHours));
        break;
    }
    this.schedulePersist();
  }

  private schedulePersist() {
    if (!browser || !this.serverLoaded) return;
    if (this.persistTimer) clearTimeout(this.persistTimer);
    this.persistTimer = setTimeout(() => {
      this.persistTimer = null;
      void saveUserSettings(this.snapshot)
        .then((s) => this.apply(s, false))
        .catch(() => undefined);
    }, 300);
  }
}

export const settingsManager = new SettingsManager();
