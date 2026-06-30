<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { createAlbum } from '$lib/api/albums';
	import { X, LoaderCircle, Plus } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';

	let {
		open = $bindable(false),
		onCreated,
	}: {
		open?: boolean;
		onCreated?: () => void;
	} = $props();

	let title = $state('');
	let description = $state('');
	let submitting = $state(false);

	async function submit() {
		if (!title.trim()) return;
		submitting = true;
		try {
			await createAlbum({ title: title.trim(), description: description.trim() || undefined });
			title = '';
			description = '';
			onCreated?.();
			open = false;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.create_failed());
		} finally {
			submitting = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Overlay
		class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			duration-200"
	/>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-xl bg-surface p-6 shadow-dialog
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			duration-200"
	>
		<div class="mb-4 flex items-center justify-between">
			<h2 class="text-lg font-semibold text-ink">{m.albums_create()}</h2>
			<Dialog.Close class="rounded-md p-1 text-ink-4 hover:text-ink-3" aria-label={m.close()}>
				<X size={20} />
			</Dialog.Close>
		</div>

		<div class="space-y-4">
			<div>
				<label for="album-title" class="mb-1 block text-sm font-medium text-ink-2">{m.albums_title()}</label>
				<input
					id="album-title"
					type="text"
					bind:value={title}
					class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
					placeholder={m.albums_title_placeholder()}
				/>
			</div>
			<div>
				<label for="album-desc" class="mb-1 block text-sm font-medium text-ink-2">{m.albums_description()}</label>
				<textarea
					id="album-desc"
					bind:value={description}
					rows={3}
					class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
					placeholder={m.albums_desc_placeholder()}
				></textarea>
			</div>
		</div>

		<div class="mt-6 flex justify-end gap-2">
			<Dialog.Close class="rounded-lg border border-line px-4 py-2 text-sm font-medium text-ink-2 transition-colors hover:bg-surface-sunken">
				{m.cancel()}
			</Dialog.Close>
			<button
				type="button"
				onclick={submit}
				disabled={!title.trim() || submitting}
				class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
			>
				{#if submitting}
					<LoaderCircle size={15} class="animate-spin" />
				{:else}
					<Plus size={15} />
				{/if}
				{m.albums_create()}
			</button>
		</div>
	</Dialog.Content>
</Dialog.Root>
