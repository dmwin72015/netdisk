<script lang="ts">
  import { Dialog } from '$lib/ui/dialog';
  import * as m from '$lib/paraglide/messages';

  export type FileConflict = {
    uid: string;
    fileName: string;
    existingSlug: string;
  };

  export type ConflictDialogResult = {
    strategy: 'overwrite' | 'skip' | 'keep_both';
    applyToAll: boolean;
  };

  let {
    conflicts = $bindable([]),
    open = $bindable(false),
    onResolve,
    onCancel,
  }: {
    conflicts: FileConflict[];
    open?: boolean;
    onResolve: (results: Map<string, ConflictDialogResult>) => void;
    onCancel: () => void;
  } = $props();

  let applyToAll = $state(false);
  let currentIndex = $state(0);
  let resolved = $state<Map<string, ConflictDialogResult>>(new Map());
  let globalStrategy = $state<ConflictDialogResult['strategy'] | null>(null);

  let current = $derived(conflicts[currentIndex]);
  let isLast = $derived(currentIndex >= conflicts.length - 1);

  $effect(() => {
    if (open) {
      currentIndex = 0;
      resolved = new Map();
      applyToAll = false;
      globalStrategy = null;
    }
  });

  function handleStrategy(strategy: ConflictDialogResult['strategy']) {
    if (applyToAll && globalStrategy) {
      for (const c of conflicts) {
        resolved.set(c.uid, { strategy: globalStrategy, applyToAll: true });
      }
    } else {
      if (current) {
        resolved.set(current.uid, { strategy, applyToAll });
      }
    }

    if (applyToAll && globalStrategy) {
      finish();
    } else if (isLast) {
      finish();
    } else {
      currentIndex++;
    }
  }

  function handleApplyAllChange(checked: boolean) {
    applyToAll = checked;
    if (checked && !globalStrategy) {
      globalStrategy = 'overwrite';
    }
  }

  function finish() {
    open = false;
    onResolve(resolved);
  }

  function getFileExtension(name: string): string {
    const i = name.lastIndexOf('.');
    return i > 0 ? name.slice(i) : '';
  }

  function getFileNameWithoutExt(name: string): string {
    const i = name.lastIndexOf('.');
    return i > 0 ? name.slice(0, i) : name;
  }

  function windowsKeepBothName(name: string): string {
    const ext = getFileExtension(name);
    const base = getFileNameWithoutExt(name);
    const m = base.match(/^(.+?)\s*\((\d+)\)$/);
    const n = m ? parseInt(m[2], 10) + 1 : 2;
    const prefix = m ? m[1].trimEnd() : base;
    return `${prefix} (${n})${ext}`;
  }

  function handleClose(v: boolean) {
    if (!v) onCancel();
  }
</script>

<Dialog
  bind:open
  onOpenChangeComplete={handleClose}
  title={m.upload_conflict_title()}
  closable={true}
  footer={false}
  class="max-w-sm"
>
  {#if current}
    <div class="space-y-4">
      <p class="text-sm text-ink-3">
        {m.upload_conflict_message({ name: current.fileName })}
      </p>

      <!-- Action buttons -->
      <div class="space-y-2">
        <button
          type="button"
          onclick={() => handleStrategy('overwrite')}
          class="flex w-full items-center gap-3 rounded-lg border border-line px-4 py-2.5 text-left transition-colors hover:border-primary hover:bg-primary-soft"
        >
          <div>
            <div class="text-sm font-medium text-ink-2">{m.upload_conflict_overwrite()}</div>
            <div class="text-xs text-ink-3">{current.fileName}</div>
          </div>
        </button>
        <button
          type="button"
          onclick={() => handleStrategy('keep_both')}
          class="flex w-full items-center gap-3 rounded-lg border border-line px-4 py-2.5 text-left transition-colors hover:border-primary hover:bg-primary-soft"
        >
          <div>
            <div class="text-sm font-medium text-ink-2">{m.upload_conflict_keep_both()}</div>
            <div class="text-xs text-ink-3">{windowsKeepBothName(current.fileName)}</div>
          </div>
        </button>
        <button
          type="button"
          onclick={() => handleStrategy('skip')}
          class="flex w-full items-center gap-3 rounded-lg border border-line px-4 py-2.5 text-left transition-colors hover:border-danger hover:bg-danger-soft"
        >
          <div>
            <div class="text-sm font-medium text-ink-2">{m.upload_conflict_skip()}</div>
            <div class="text-xs text-ink-3">{current.fileName}</div>
          </div>
        </button>
      </div>

      <!-- Apply to all -->
      {#if conflicts.length > 1}
        <label class="flex cursor-pointer items-center gap-2">
          <input
            type="checkbox"
            checked={applyToAll}
            onchange={(e) => handleApplyAllChange(e.currentTarget.checked)}
            class="h-4 w-4 rounded border-line text-primary"
          />
          <span class="text-xs text-ink-3">{m.upload_conflict_apply_all()}</span>
        </label>
      {/if}

      <!-- Progress indicator -->
      {#if conflicts.length > 1 && !applyToAll}
        <div class="text-xs text-ink-4">
          {currentIndex + 1} / {conflicts.length}
        </div>
      {/if}
    </div>
  {/if}
</Dialog>
