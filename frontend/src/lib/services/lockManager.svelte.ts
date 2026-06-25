import type { NormalizedFile } from "$lib/types/file";
import {
  setDirectoryLock,
  clearDirectoryLock,
  unlockDirectory,
} from "$lib/api/files";
import { settingsManager } from "./settingsManager.svelte";
import { promptInput } from "$lib/dialog";

class LockManager {
  /** Slugs temporarily unlocked in the current session. */
  unlockedSlugs = $state<Set<string>>(new Set());

  /** Whether a directory is temporarily unlocked in this session. */
  isUnlocked(slug: string): boolean {
    return this.unlockedSlugs.has(slug);
  }

  /**
   * Check if a file is effectively locked (has password AND not temporarily unlocked).
   * This is the main check for blocking operations.
   */
  isEffectivelyLocked(file: NormalizedFile): boolean {
    return file.hasPassword && !this.unlockedSlugs.has(file.slug);
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
    this.unlockedSlugs.add(file.slug);
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
    this.unlockedSlugs.delete(file.slug);
  }
}

export const lockManager = new LockManager();
