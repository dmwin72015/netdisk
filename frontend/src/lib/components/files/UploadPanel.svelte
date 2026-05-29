<script lang="ts">
	import { CheckCircle2, ChevronDown, ChevronUp, Loader2, Pause, Play, X, XCircle } from '@lucide/svelte';
	import { fmtSize, fmtSpeed } from '$lib/utils/format';
	import type { UploadItem } from '$lib/types/upload';
	import * as m from '$lib/paraglide/messages';

	let {
		items,
		onPause,
		onResume,
		onDelete,
		onClear
	}: {
		items: UploadItem[];
		onPause: (uid: string) => void;
		onResume: (uid: string) => void;
		onDelete: (uid: string) => void;
		onClear: () => void;
	} = $props();

	let open = $state(true);

	let completedCount = $derived(items.filter((i) => i.phase === 'completed').length);
	let failedCount = $derived(items.filter((i) => i.phase === 'failed').length);
	let totalSpeed = $derived(items.reduce((s, i) => s + i.speed, 0));
	let hasActive = $derived(items.some((i) => i.phase !== 'completed' && i.phase !== 'failed'));
</script>

{#if items.length > 0}
	<div class="fixed bottom-4 right-4 z-40 w-80 sm:w-96">
		{#if open}
			<div class="rounded-xl border border-gray-200 bg-white shadow-lg">
				<div class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5">
					<div class="flex items-center gap-2">
						<span class="text-sm font-medium text-gray-800">
							{hasActive ? m.upload_panel_uploading() : m.upload_panel_done()}
						</span>
						<span class="text-xs text-gray-400">{completedCount}/{items.length}</span>
						{#if totalSpeed > 0}
							<span class="text-xs text-blue-500">{fmtSpeed(totalSpeed)}</span>
						{/if}
					</div>
					<div class="flex items-center gap-1">
						<button type="button" onclick={onClear} class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600">
							<X size={14} />
						</button>
						<button type="button" onclick={() => (open = false)} class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600">
							<ChevronDown size={14} />
						</button>
					</div>
				</div>
				<div class="max-h-72 overflow-y-auto">
					{#each items as item (item.uid)}
						<div class="border-b border-gray-50 px-4 py-2 last:border-0">
							<div class="flex items-center gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm text-gray-700" title={item.fileName}>{item.fileName}</p>
									<div class="flex items-center gap-2 text-xs text-gray-400">
										<span>{fmtSize(item.fileSize)}</span>
										{#if item.phase === 'uploading' && item.speed > 0}
											<span class="text-blue-500">{fmtSpeed(item.speed)}</span>
										{/if}
									</div>
								</div>
								<div class="flex shrink-0 items-center gap-0.5">
									{#if item.phase === 'hashing'}
										<span class="text-xs text-gray-400">{m.hashing()} {item.hashProgress > 0 ? `${item.hashProgress}%` : ''}</span>
										<Loader2 size={14} class="animate-spin text-gray-300" />
									{:else if item.phase === 'verifying'}
										<span class="text-xs text-gray-400">{m.checking()}</span>
										<Loader2 size={14} class="animate-spin text-gray-300" />
									{:else if item.phase === 'importing'}
										<span class="text-xs text-blue-500">{m.importing()}</span>
										<Loader2 size={14} class="animate-spin text-blue-400" />
									{:else if item.phase === 'uploading'}
										<span class="mr-1 text-xs font-medium text-blue-600">{item.progress}%</span>
										<button type="button" onclick={() => onPause(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-amber-500">
											<Pause size={13} />
										</button>
									{:else if item.phase === 'paused'}
										<span class="mr-1 text-xs font-medium text-amber-500">{m.paused()}</span>
										<button type="button" onclick={() => onResume(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500">
											<Play size={13} />
										</button>
									{:else if item.phase === 'completed'}
										<CheckCircle2 size={14} class="text-green-500" />
									{:else if item.phase === 'failed'}
										<XCircle size={14} class="text-red-500" />
										<button type="button" onclick={() => onResume(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500">
											<Play size={13} />
										</button>
									{/if}
									{#if item.phase !== 'completed' && item.phase !== 'importing'}
										<button type="button" onclick={() => onDelete(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-red-500">
											<X size={13} />
										</button>
									{/if}
								</div>
							</div>
							{#if item.phase === 'uploading'}
								<div class="mt-1.5 h-1 overflow-hidden rounded-full bg-gray-100">
									<div class="h-full rounded-full bg-blue-500 transition-all" style="width:{item.progress}%"></div>
								</div>
							{:else if item.phase === 'paused'}
								<div class="mt-1.5 h-1 overflow-hidden rounded-full bg-gray-100">
									<div class="h-full rounded-full bg-amber-400 transition-all" style="width:{item.progress}%"></div>
								</div>
							{:else if item.phase === 'failed' && item.errorMsg}
								<p class="mt-1 text-xs text-red-500">{item.errorMsg}</p>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<button type="button" onclick={() => (open = true)} class="flex w-full items-center gap-3 rounded-xl border border-gray-200 bg-white px-4 py-2.5 shadow-lg transition-colors hover:bg-gray-50">
				{#if hasActive}
					<Loader2 size={16} class="animate-spin text-blue-500" />
				{:else}
					<CheckCircle2 size={16} class="text-green-500" />
				{/if}
				<span class="flex-1 text-sm text-gray-700">{completedCount}/{items.length}</span>
				{#if failedCount > 0}
					<span class="text-xs text-red-500">{m.upload_panel_failed({ count: failedCount })}</span>
				{/if}
				<ChevronUp size={14} class="text-gray-400" />
			</button>
		{/if}
	</div>
{/if}
