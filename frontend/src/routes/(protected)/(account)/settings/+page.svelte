<script lang="ts">
  import {
    getShowSystemDirs,
    setShowSystemDirs,
    getUploadConcurrency,
    setUploadConcurrency,
    getDuplicateStrategy,
    setDuplicateStrategy,
    getDirectoryUnlockTtlHours,
    setDirectoryUnlockTtlHours,
    getThemePreference,
    setThemePreference,
    exportPreferences,
    importPreferences,
  } from "$lib/stores/file-preferences.svelte";
  import { UPLOAD_FILE_CONCURRENCY } from "$lib/upload-concurrency";
  import { Eye, Upload, FileWarning, Lock, FileJson, Globe } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import * as m from "$lib/paraglide/messages";
  import { lockManager } from "$lib/services/lockManager.svelte";
  import LanguageDropdown from "$lib/components/LanguageDropdown.svelte";

  let importInput: HTMLInputElement | undefined = $state();
  let importing = $state(false);

  $effect(() => {
    const ttl = getDirectoryUnlockTtlHours();
    void ttl;
    lockManager.recheckExpiration();
  });

  const ttlOptions = [
    { value: 1, label: m.settings_ttl_1h() },
    { value: 2, label: m.settings_ttl_2h() },
    { value: 6, label: m.settings_ttl_6h() },
    { value: 24, label: m.settings_ttl_24h() },
    { value: -1, label: m.settings_ttl_forever() },
  ];

  function downloadSettingsJson() {
    const blob = new Blob([JSON.stringify(exportPreferences(), null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "netdisk-settings.json";
    link.click();
    URL.revokeObjectURL(url);
  }

  async function onImportSettings(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    importing = true;
    try {
      const raw = await file.text();
      await importPreferences(JSON.parse(raw));
      toast.success(m.settings_restored());
    } catch {
      toast.error(m.settings_restore_failed());
    } finally {
      importing = false;
    }
  }
</script>

<div class="space-y-8">
  <!-- Language -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <Globe size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.language()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div class="flex items-center justify-between gap-4 px-5 py-4">
        <div class="min-w-0">
          <p class="text-sm font-medium text-ink-2">{m.language()}</p>
          <p class="mt-1 text-xs leading-5 text-ink-3">
            {m.language_desc()}
          </p>
        </div>
        <LanguageDropdown
          triggerClass="flex items-center gap-1.5 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-sunken"
        />
      </div>
    </div>
  </section>

  <!-- Display -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <Eye size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.display_settings()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div class="flex items-center justify-between gap-4 border-b border-line-soft px-5 py-4">
        <div class="min-w-0">
          <p class="text-sm font-medium text-ink-2">{m.theme_settings()}</p>
          <p class="mt-1 text-xs leading-5 text-ink-3">
            {m.theme_settings_desc()}
          </p>
        </div>
        <select
          value={getThemePreference()}
          onchange={(e) =>
            setThemePreference(
              (e.currentTarget as HTMLSelectElement).value as
                | "system"
                | "light"
                | "dark",
            )}
          class="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-2 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option value="system">{m.theme_system()}</option>
          <option value="light">{m.theme_light()}</option>
          <option value="dark">{m.theme_dark()}</option>
        </select>
      </div>
      <div class="flex items-center justify-between gap-4 px-5 py-4">
        <div class="min-w-0">
          <p class="text-sm font-medium text-ink-2">{m.show_system_dirs()}</p>
          <p class="mt-1 text-xs leading-5 text-ink-3">
            {m.show_system_dirs_desc()}
          </p>
        </div>
        <input
          type="checkbox"
          checked={getShowSystemDirs()}
          onchange={(e) =>
            setShowSystemDirs((e.currentTarget as HTMLInputElement).checked)}
          class="h-4 w-4 shrink-0 rounded border-line text-primary focus:ring-primary"
        />
      </div>
    </div>
  </section>

  <!-- Directory Lock -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <Lock size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.settings_dir_lock()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div class="flex items-center justify-between gap-4 px-5 py-4">
        <div class="min-w-0">
<p class="text-sm font-medium text-ink-2">{m.settings_password_ttl()}</p>
           <p class="mt-1 text-xs leading-5 text-ink-3">
             {m.settings_password_ttl_desc()}
           </p>
        </div>
        <select
          value={getDirectoryUnlockTtlHours()}
          onchange={(e) =>
            setDirectoryUnlockTtlHours(
              parseInt((e.currentTarget as HTMLSelectElement).value, 10),
            )}
          class="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-2 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          {#each ttlOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
    </div>
  </section>

  <!-- Backup -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <FileJson size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.settings_backup()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div
        class="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <div class="min-w-0">
<p class="text-sm font-medium text-ink-2">{m.settings_backup_desc()}</p>
           <p class="mt-1 text-xs leading-5 text-ink-3">
             {m.settings_backup_desc()}
           </p>
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onclick={downloadSettingsJson}
            class="rounded-lg border border-line px-3 py-2 text-sm font-medium text-ink-2 transition-colors hover:bg-surface-sunken"
          >
            {m.settings_export_json()}
          </button>
          <button
            type="button"
            disabled={importing}
            onclick={() => importInput?.click()}
            class="rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:opacity-50"
          >
            {importing ? m.settings_importing() : m.settings_import_from_json()}
          </button>
          <input
            bind:this={importInput}
            type="file"
            accept="application/json,.json"
            class="hidden"
            onchange={onImportSettings}
          />
        </div>
      </div>
    </div>
  </section>

  <!-- Upload -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <Upload size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.upload_title()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div class="flex items-center justify-between gap-4 px-5 py-4">
        <div class="min-w-0">
          <p class="text-sm font-medium text-ink-2">{m.upload_concurrency()}</p>
          <p class="mt-1 text-xs leading-5 text-ink-3">
            {m.upload_concurrency_desc()}
          </p>
        </div>
        <span class="flex items-center gap-3">
          <input
            type="range"
            min="1"
            max={UPLOAD_FILE_CONCURRENCY}
            step="1"
            value={getUploadConcurrency()}
            oninput={(e) =>
              setUploadConcurrency(
                parseInt((e.currentTarget as HTMLInputElement).value, 10),
              )}
            class="h-1.5 w-24 appearance-none rounded-full bg-line accent-primary"
          />
          <span class="w-6 text-center text-sm font-medium text-ink-2"
            >{getUploadConcurrency()}</span
          >
        </span>
      </div>
    </div>
  </section>

  <!-- Duplicates -->
  <section>
    <div class="mb-3 flex items-center gap-2">
      <FileWarning size={16} class="text-ink-4" />
      <h2 class="text-sm font-semibold uppercase tracking-wide text-ink-3">
        {m.duplicate_strategy()}
      </h2>
    </div>
    <div class="rounded-xl border border-line-soft bg-surface">
      <div class="border-b border-line-soft px-5 py-4">
        <p class="text-sm leading-5 text-ink-3">
          {m.duplicate_strategy_desc()}
        </p>
      </div>
      <div class="space-y-1 px-2 py-2">
        {#each [["prompt", m.duplicate_strategy_prompt()], ["overwrite", m.duplicate_strategy_overwrite()], ["keep_both", m.duplicate_strategy_keep_both()], ["skip", m.duplicate_strategy_skip()]] as [value, label]}
          <button
            type="button"
            onclick={() => setDuplicateStrategy(value)}
            class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors {getDuplicateStrategy() ===
            value
              ? 'bg-primary-soft text-primary'
              : 'text-ink-2 hover:bg-surface-sunken'}"
          >
            <span
              class="flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2 {getDuplicateStrategy() ===
              value
                ? 'border-primary'
                : 'border-line'}"
            >
              {#if getDuplicateStrategy() === value}
                <span class="h-2 w-2 rounded-full bg-primary"></span>
              {/if}
            </span>
            <span class="text-sm">{label}</span>
          </button>
        {/each}
      </div>
    </div>
  </section>
</div>
