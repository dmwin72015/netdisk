<script lang="ts">
  import { LoaderCircle, CircleCheck, ExternalLink } from '@lucide/svelte';
  import { toast } from 'svelte-sonner';
  import { goto } from '$app/navigation';
  import { Dialog } from '$lib/ui/dialog';
  import { urlUpload } from '$lib/api/upload';
  import * as m from '$lib/paraglide/messages';

  let {
    open = $bindable(false),
    parentSlug = '',
  }: {
    open?: boolean;
    parentSlug?: string;
  } = $props();

  let url = $state('');
  let fileName = $state('');
  let submitting = $state(false);
  let done = $state(false);

  $effect(() => {
    if (open) {
      url = '';
      fileName = '';
      submitting = false;
      done = false;
    }
  });

  async function handleSubmit() {
    if (!url.trim()) return;
    submitting = true;
    try {
      await urlUpload(url.trim(), fileName.trim() || undefined, parentSlug || undefined);
      done = true;
      toast.success(m.remote_upload_queued());
    } catch (e) {
      toast.error(m.remote_upload_failed());
    } finally {
      submitting = false;
    }
  }
</script>

<Dialog
  bind:open
  title={done ? '' : m.remote_upload()}
  footer={false}
  class="max-w-sm"
>
  {#if done}
    <div class="flex flex-col items-center gap-3 py-6">
      <CircleCheck class="size-10 text-green-500" />
      <p class="text-sm text-gray-500">{m.remote_upload_queued()}</p>
      <div class="flex items-center gap-2 mt-1">
        <button
          type="button"
          onclick={() => goto('/tasks')}
          class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700"
        >
          {m.view_tasks()}
          <ExternalLink class="size-3.5" />
        </button>
      </div>
    </div>
    <div class="flex items-center justify-end gap-2 border-t border-gray-100 px-5 py-3 -mx-5 -mb-4">
      <button
        type="button"
        class="rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50"
        onclick={() => { open = false; }}
      >
        {m.close()}
      </button>
    </div>
  {:else}
    <div class="flex flex-col gap-1.5">
      <label for="url-input" class="text-sm font-medium text-gray-700">
        {m.remote_upload_url_label()}
      </label>
      <input
        id="url-input"
        type="url"
        bind:value={url}
        placeholder={m.remote_upload_url_placeholder()}
        class="rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none transition-colors focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
      />
    </div>

    <div class="flex flex-col gap-1.5 mt-2">
      <label for="filename-input" class="text-sm font-medium text-gray-700">
        {m.remote_upload_filename()}
      </label>
      <input
        id="filename-input"
        type="text"
        bind:value={fileName}
        class="rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none transition-colors focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
      />
    </div>

    <div class="flex items-center justify-end gap-2 border-t border-gray-100 px-5 py-3 -mx-5 -mb-4 mt-2">
      <button
        type="button"
        class="rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50"
        onclick={() => { open = false; }}
      >
        {m.cancel()}
      </button>
      <button
        type="button"
        onclick={handleSubmit}
        disabled={!url.trim() || submitting}
        class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {#if submitting}
          <LoaderCircle size={15} class="animate-spin" />
        {/if}
        {m.remote_upload_start()}
      </button>
    </div>
  {/if}
</Dialog>
