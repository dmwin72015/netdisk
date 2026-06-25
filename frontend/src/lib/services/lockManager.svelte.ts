import type { NormalizedFile } from "$lib/types/file";
import type { FileItem } from "$lib/api/files";
import { SvelteSet } from "svelte/reactivity";
import { browser } from "$app/environment";
import { user } from "$lib/stores/auth";
import {
  setDirectoryLock,
  clearDirectoryLock,
  unlockDirectory,
} from "$lib/api/files";
import { settingsManager } from "./settingsManager.svelte";
import { promptInput } from "$lib/dialog";

const LS_UNLOCKED_DIRS = "nd.files.unlockedDirs";

class LockManager {
  /** Slugs currently locked (has password set). */
  private lockedSlugs = $state<Set<string>>(new Set());
  /** Slugs temporarily unlocked in the current session. */
  unlockedSlugs = $state<SvelteSet<string>>(new SvelteSet());

  /** Persisted unlocks with timestamps: slug → unlock time (ms). */
  private persistedUnlocks = $state<Map<string, number>>(new Map());

  constructor() {
    this.loadPersistedUnlocks();
    this.unsubscribeAuth = user.subscribe((u) => {
      if (!u) this.reset();
    });
  }

  private unsubscribeAuth: (() => void) | null = null;

  dispose() {
    this.unsubscribeAuth?.();
  }

  /** Clear all local lock state (called on logout). */
  reset() {
    this.lockedSlugs = new Set();
    this.unlockedSlugs = new SvelteSet();
    this.persistedUnlocks = new Map();
    if (browser) localStorage.removeItem(LS_UNLOCKED_DIRS);
  }

  isLocked(slug: string): boolean {
    return this.lockedSlugs.has(slug);
  }

  /** Whether a directory is effectively locked (locked AND not temporarily unlocked). */
  isEffectivelyLocked(file: NormalizedFile): boolean {
    return this.lockedSlugs.has(file.slug) && !this.unlockedSlugs.has(file.slug);
  }

  /** Prompt for password and temporarily unlock a directory. */
  async unlock(slug: string, name?: string): Promise<boolean> {
    const password = await promptInput(
      "目录密码",
      `请输入${name ? `「${name}」` : "目录"}的密码`,
      undefined,
      128,
    );
    if (!password) return false;
    try {
      await unlockDirectory(
        slug,
        password,
        settingsManager.directoryUnlockTtlHours,
      );
      this.unlockedSlugs.add(slug);
      this.persistedUnlocks.set(slug, Date.now());
      this.persistUnlocks();
      return true;
    } catch {
      throw new Error("目录密码错误");
    }
  }

  /** Set a lock password on a directory. */
  async lock(file: NormalizedFile): Promise<void> {
    const password = await promptInput(
      "设置目录密码",
      `请输入「${file.name}」的目录密码（至少 4 位）`,
      undefined,
      128,
    );
    if (!password) return;
    await setDirectoryLock(file.id, password);
    this.lockedSlugs.add(file.slug);
    this.unlockedSlugs.add(file.slug);
    this.persistedUnlocks.set(file.slug, Date.now());
    this.persistUnlocks();
  }

  /** Clear the lock password on a directory. */
  async clearLock(file: NormalizedFile): Promise<void> {
    const password = await promptInput(
      "取消目录密码",
      `请输入「${file.name}」的目录密码`,
      undefined,
      128,
    );
    if (!password) return;
    await clearDirectoryLock(file.id, password);
    this.lockedSlugs.delete(file.slug);
    this.unlockedSlugs.delete(file.slug);
    this.persistedUnlocks.delete(file.slug);
    this.persistUnlocks();
  }

  /** Sync locked state from file list data (call after refresh). */
  syncFromFiles(files: FileItem[]) {
    this.lockedSlugs = new Set(
      files.filter((f) => f.hasPassword).map((f) => f.slug),
    );
    this.cleanupExpiredUnlocks();
  }

  /** Re-evaluate all persisted unlocks against current TTL setting. */
  recheckExpiration() {
    this.cleanupExpiredUnlocks();
  }

  // --- Private ---

  private loadPersistedUnlocks() {
    if (!browser) return;
    try {
      const raw = localStorage.getItem(LS_UNLOCKED_DIRS);
      if (!raw) return;
      const parsed = JSON.parse(raw) as Record<string, number>;
      const now = Date.now();
      const ttlMs = settingsManager.directoryUnlockTtlHours * 3600 * 1000;
      for (const [slug, ts] of Object.entries(parsed)) {
        if (settingsManager.directoryUnlockTtlHours !== -1 && now - ts > ttlMs) {
          continue;
        }
        this.persistedUnlocks.set(slug, ts);
        this.unlockedSlugs.add(slug);
      }
    } catch {
      // ignore corrupt data
    }
  }

  private persistUnlocks() {
    if (!browser) return;
    const obj: Record<string, number> = {};
    for (const [slug, ts] of this.persistedUnlocks) {
      obj[slug] = ts;
    }
    localStorage.setItem(LS_UNLOCKED_DIRS, JSON.stringify(obj));
  }

  private cleanupExpiredUnlocks() {
    if (!browser) return;
    const now = Date.now();
    const ttlMs = settingsManager.directoryUnlockTtlHours * 3600 * 1000;
    let changed = false;
    for (const [slug, ts] of this.persistedUnlocks) {
      if (
        settingsManager.directoryUnlockTtlHours !== -1 &&
        now - ts > ttlMs
      ) {
        this.persistedUnlocks.delete(slug);
        this.unlockedSlugs.delete(slug);
        changed = true;
      }
    }
    if (changed) this.persistUnlocks();
  }
}

export const lockManager = new LockManager();
