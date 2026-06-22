<script lang="ts">
	import { createAlbum } from '$lib/api/albums';
	import { X, LoaderCircle, Plus } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';

	let {
		show = $bindable(),
		onCreated,
	}: {
		show: boolean;
		onCreated: () => void;
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
			onCreated();
			show = false;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.create_failed());
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && !submitting) show = false;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if show}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4" role="presentation">
		<button
			type="button"
			class="absolute inset-0 bg-overlay"
			aria-label={m.close()}
			disabled={submitting}
			onclick={() => { if (!submitting) show = false; }}
		></button>
		<div
			role="dialog"
			aria-modal="true"
			aria-labelledby="album-create-title"
			class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-dialog"
		>
			<div class="mb-4 flex items-center justify-between">
				<h2 id="album-create-title" class="text-lg font-semibold text-ink">{m.albums_create()}</h2>
				<button type="button" onclick={() => (show = false)} class="rounded-md p-1 text-ink-4 hover:text-ink-3" aria-label={m.close()}>
					<X size={20} />
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label for="album-title" class="mb-1 block text-sm font-medium text-ink-2">{m.albums_title()}</label>
					<input
						id="album-title"
						type="text"
						bind:value={title}
						class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						placeholder={m.albums_title_placeholder()}
					/>
				</div>
				<div>
					<label for="album-desc" class="mb-1 block text-sm font-medium text-ink-2">{m.albums_description()}</label>
					<textarea
						id="album-desc"
						bind:value={description}
						rows={3}
						class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						placeholder={m.albums_desc_placeholder()}
					></textarea>
				</div>
			</div>

			<div class="mt-6 flex justify-end gap-2">
				<button
					type="button"
					onclick={() => (show = false)}
					class="rounded-lg border border-line px-4 py-2 text-sm font-medium text-ink-2 transition-colors hover:bg-surface-muted"
				>
					{m.cancel()}
				</button>
				<button
					type="button"
					onclick={submit}
					disabled={!title.trim() || submitting}
					class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if submitting}
						<LoaderCircle size={15} class="animate-spin" />
					{:else}
						<Plus size={15} />
					{/if}
					{m.albums_create()}
				</button>
			</div>
		</div>
	</div>
{/if}
