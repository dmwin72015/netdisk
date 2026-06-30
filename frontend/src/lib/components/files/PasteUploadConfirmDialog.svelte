<script lang="ts">
	import { fmtSize } from '$lib/utils/format';
	import { AlertTriangle, Check, Clipboard, X } from '@lucide/svelte';
	import { Dialog } from '$lib/ui/dialog';
	import * as m from '$lib/paraglide/messages';

	let {
		acceptedFiles,
		rejectedFiles = [],
		targetLabel,
		open = $bindable(false),
		onConfirm,
		onCancel,
	}: {
		acceptedFiles: File[];
		rejectedFiles?: File[];
		targetLabel: string;
		open?: boolean;
		onConfirm: (files: File[]) => void | Promise<void>;
		onCancel: () => void;
	} = $props();

	const MAX_VISIBLE_FILES = 8;
	let visibleAccepted = $derived(acceptedFiles.slice(0, MAX_VISIBLE_FILES));
	let hiddenAcceptedCount = $derived(Math.max(0, acceptedFiles.length - visibleAccepted.length));
	let totalSize = $derived(acceptedFiles.reduce((sum, file) => sum + file.size, 0));

	let fileName = $state('');
	let filenameModified = $state(false);

	$effect(() => {
		if (open && acceptedFiles.length > 0) {
			fileName = acceptedFiles[0].name;
			filenameModified = false;
		}
	});

	function handleOpenChangeComplete(value: boolean) {
		if (!value) {
			onCancel();
		}
	}

	function handleFileNameInput() {
		filenameModified = true;
	}

	function buildRenamedFiles(): File[] {
		if (acceptedFiles.length === 0) return acceptedFiles;

		const baseName = fileName.trim() || acceptedFiles[0].name;

		if (acceptedFiles.length === 1) {
			const [file] = acceptedFiles;
			return [new File([file], baseName, { type: file.type, lastModified: file.lastModified })];
		}

		const dotIndex = baseName.lastIndexOf('.');
		let namePart: string;
		let extPart: string;
		if (dotIndex > 0) {
			namePart = baseName.slice(0, dotIndex);
			extPart = baseName.slice(dotIndex);
		} else {
			namePart = baseName;
			extPart = '';
		}

		return acceptedFiles.map((file, i) => {
			const suffix = i === 0 ? '' : ` (${i})`;
			const newName = `${namePart}${suffix}${extPart}`;
			return new File([file], newName, { type: file.type, lastModified: file.lastModified });
		});
	}
</script>

<Dialog
	bind:open
	onOpenChangeComplete={handleOpenChangeComplete}
	onCancel={onCancel}
	title={m.paste_confirm_title()}
	footer={false}
	size="md"
	bodyClass="p-0"
	closable={false}
>
	<div class="border-b border-line-soft px-5 py-4">
		<div class="flex items-start gap-3">
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
				<Clipboard size={20} />
			</div>
			<div class="min-w-0">
				<p class="text-sm text-ink-2">
					{m.paste_uploading_files({ count: String(acceptedFiles.length), target: targetLabel })}
				</p>
				<p class="mt-1 text-xs text-ink-4">{m.paste_total_size({ size: fmtSize(totalSize) })}</p>
			</div>
		</div>
	</div>

	{#if acceptedFiles.length > 0}
		<!-- Filename input -->
		<div class="border-b border-line-soft px-5 py-3">
			<label for="paste-filename" class="text-sm font-medium text-ink-3">{m.paste_filename()}</label>
			<input
				id="paste-filename"
				type="text"
				bind:value={fileName}
				oninput={handleFileNameInput}
				placeholder={m.paste_filename_placeholder()}
				class="mt-1.5 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
			/>
		</div>

		<div class="max-h-[45vh] overflow-y-auto px-2 py-3">
			{#each visibleAccepted as file (file.name + file.size + file.lastModified)}
				<div class="flex items-center gap-3 rounded-lg px-3 py-2">
					<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-surface-sunken text-ink-3">
						<Check size={15} />
					</div>
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm text-ink-2" title={file.name}>{file.name}</p>
					</div>
					<span class="shrink-0 text-xs text-ink-4">{fmtSize(file.size)}</span>
				</div>
			{/each}
			{#if hiddenAcceptedCount > 0}
				<p class="px-3 py-2 text-xs text-ink-4">{m.paste_more_files({ count: String(hiddenAcceptedCount) })}</p>
			{/if}
		</div>
	{:else}
		<div class="px-5 py-8 text-center text-sm text-ink-3">
			{m.paste_no_files()}
		</div>
	{/if}

	{#if rejectedFiles.length > 0}
		<div class="mx-5 mt-4 rounded-xl border border-warning bg-warning-soft px-3 py-2 text-sm text-warning">
			<div class="flex items-start gap-2">
				<AlertTriangle size={16} class="mt-0.5 shrink-0" />
				<p>{m.paste_skipped_files({ count: String(rejectedFiles.length) })}</p>
			</div>
		</div>
	{/if}

	<div class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3">
		<button
			type="button"
			onclick={() => { open = false; }}
			class="inline-flex items-center gap-1.5 rounded-lg border border-line bg-surface px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-sunken"
		>
			<X size={14} /> {m.cancel()}
		</button>
		<button
			type="button"
			onclick={() => onConfirm(buildRenamedFiles())}
			disabled={acceptedFiles.length === 0}
			class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> {m.paste_confirm_upload()}
		</button>
	</div>
</Dialog>
