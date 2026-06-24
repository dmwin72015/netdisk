import type { NameConflictInfo, NameConflictResult } from "$lib/upload-manager.svelte";

type ConflictRequest = {
  conflicts: NameConflictInfo[];
  resolve: (results: Map<string, NameConflictResult>) => void;
};

export class ConflictManager {
  open = $state(false);
  conflicts = $state<NameConflictInfo[]>([]);

  private queue: ConflictRequest[] = [];
  private active: ConflictRequest | null = null;

  onUploadConflicts: (conflicts: NameConflictInfo[]) => Promise<Map<string, NameConflictResult>> = (
    conflicts,
  ) =>
    new Promise((resolve) => {
      this.queue.push({ conflicts, resolve });
      this.showNext();
    });

  private showNext() {
    if (this.active || this.queue.length === 0) return;
    this.active = this.queue.shift()!;
    this.conflicts = this.active.conflicts;
    this.open = true;
  }

  finish(results: Map<string, NameConflictResult>) {
    const current = this.active;
    this.active = null;
    this.conflicts = [];
    current?.resolve(results);
    this.showNext();
  }

  cancel() {
    const allSkipped = new Map<string, NameConflictResult>(
      (this.active?.conflicts ?? []).map((c) => [
        c.uid,
        { strategy: "skip" as const, applyToAll: false },
      ]),
    );
    this.finish(allSkipped);
  }
}
