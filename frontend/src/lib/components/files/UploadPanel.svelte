<script lang="ts">
	import { CircleCheck, ChevronDown, ChevronUp, GripHorizontal, LoaderCircle, Pause, Play, RotateCcw, Upload, X } from '@lucide/svelte';
	import { fmtSize, fmtSpeed } from '$lib/utils/format';
	import type { UploadItem } from '$lib/types/upload';
	import * as m from '$lib/paraglide/messages';

	let {
		items,
		onPause,
		onResume,
		onDelete,
		onDismiss,
		onRetry
	}: {
		items: UploadItem[];
		onPause: (uid: string) => void;
		onResume: (uid: string) => void;
		onDelete: (uid: string) => void;
		onDismiss: () => void;
		onRetry?: (uid: string) => void;
	} = $props();

	let open = $state(true);

	let completedCount = $derived(items.filter((i) => i.phase === 'completed').length);
	let failedCount = $derived(items.filter((i) => i.phase === 'failed').length);
	let interruptedCount = $derived(items.filter((i) => i.phase === 'interrupted').length);
	let totalSpeed = $derived(items.reduce((s, i) => s + i.speed, 0));
	let hasActive = $derived(items.some((i) => i.phase !== 'completed' && i.phase !== 'failed'));

	const MAX_HEIGHT = 0.7;
	const MIN_HEIGHT = 160;
	let listEl = $state<HTMLElement>();
	let panelHeight = $state(280);

	function startResize(e: PointerEvent) {
		e.preventDefault();
		const startY = e.clientY;
		const startH = panelHeight;
		function onMove(ev: PointerEvent) {
			const h = startH - (ev.clientY - startY);
			const maxH = window.innerHeight * MAX_HEIGHT;
			panelHeight = Math.max(MIN_HEIGHT, Math.min(maxH, h));
		}
		function onUp() {
			window.removeEventListener('pointermove', onMove);
			window.removeEventListener('pointerup', onUp);
		}
		window.addEventListener('pointermove', onMove);
		window.addEventListener('pointerup', onUp);
	}
</script>

{#if items.length > 0}
	<div class="fixed bottom-4 right-4 z-40 w-80 sm:w-96">
		{#if open}
			<div class="flex flex-col rounded-xl border border-line bg-surface shadow-pop" style="max-height: {MAX_HEIGHT * 100}vh;">
				<div
					onpointerdown={startResize}
					role="separator"
					aria-orientation="horizontal"
					class="flex shrink-0 cursor-s-resize items-center justify-center border-b border-line-soft py-0.5 text-ink-4 hover:text-ink-3 select-none"
				>
					<GripHorizontal size={14} />
				</div>
				<div class="flex items-center justify-between border-b border-line-soft px-4 py-2.5 shrink-0">
					<div class="flex items-center gap-2">
						<span class="text-sm font-medium text-ink-2">
							{hasActive ? m.upload_panel_uploading() : m.upload_panel_done()}
						</span>
						<span class="text-xs text-ink-4">{completedCount}/{items.length}</span>
						{#if totalSpeed > 0}
							<span class="text-xs text-primary">{fmtSpeed(totalSpeed)}</span>
						{/if}
					</div>
					<div class="flex items-center gap-1">
						<button type="button" onclick={onDismiss} class="rounded-md p-1 text-ink-4 transition-colors hover:text-ink-3">
							<X size={14} />
						</button>
						<button type="button" onclick={() => (open = false)} class="rounded-md p-1 text-ink-4 transition-colors hover:text-ink-3">
							<ChevronDown size={14} />
						</button>
					</div>
				</div>
				<div bind:this={listEl} class="overflow-y-auto" style="min-height: {MIN_HEIGHT}px; height: {panelHeight}px;">
					{#each items as item (item.uid)}
						{@const showProgress = item.phase === 'uploading' || item.phase === 'paused'}
						{@const isFailed = item.phase === 'failed'}
						{@const bgColor = item.phase === 'paused' ? 'bg-warning-soft' : 'bg-primary-soft'}
						<div class="relative border-b border-line-soft last:border-0 {showProgress ? bgColor : ''} {isFailed ? 'bg-danger-soft/60' : ''}" style={showProgress ? `background:linear-gradient(to right, ${item.phase === 'paused' ? '#fef3c7' : '#dbeafe'} ${item.progress}%, transparent ${item.progress}%)` : ''}>
							<div class="relative flex items-center gap-2 px-4 py-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm text-ink-2" title={item.fileName}>{item.fileName}</p>
									<div class="flex items-center gap-2 text-xs text-ink-4">
										<span>{fmtSize(item.fileSize)}</span>
										{#if item.phase === 'uploading' && item.speed > 0}
											<span class="text-primary">{fmtSpeed(item.speed)}</span>
										{/if}
									</div>
								</div>
								<div class="flex shrink-0 items-center gap-1.5">
									{#if item.phase === 'queued'}
									<span class="text-xs text-ink-4">{m.upload_waiting()}</span>
								{:else if item.phase === 'hashing'}
										<span class="text-xs text-ink-4">{m.hashing()} {item.hashProgress > 0 ? `${item.hashProgress}%` : ''}</span>
										<LoaderCircle size={14} class="animate-spin text-ink-4" />
									{:else if item.phase === 'verifying'}
										<span class="text-xs text-ink-4">{m.checking()}</span>
										<LoaderCircle size={14} class="animate-spin text-ink-4" />
									{:else if item.phase === 'importing'}
										<span class="text-xs text-primary">{m.importing()}</span>
										<LoaderCircle size={14} class="animate-spin text-primary" />
									{:else if item.phase === 'uploading'}
										<span class="text-xs font-medium text-primary">{item.progress}%</span>
										<button type="button" onclick={() => onPause(item.uid)} class="rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-warning">
											<Pause size={14} />
										</button>
									{:else if item.phase === 'paused'}
										<span class="text-xs font-medium text-warning">{m.paused()}</span>
										<button type="button" onclick={() => onResume(item.uid)} class="rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-primary">
											<Play size={14} />
										</button>
									{:else if item.phase === 'completed'}
										<CircleCheck size={14} class="text-success" />
									{:else if item.phase === 'failed'}
										<button type="button" onclick={() => onResume(item.uid)} class="rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-primary" title={m.upload_retry()}>
											<RotateCcw size={14} />
										</button>
									{:else if item.phase === 'interrupted'}
										<span class="text-xs font-medium text-warning">{m.interrupted()}</span>
										{#if onRetry}
											<button type="button" onclick={() => onRetry(item.uid)} class="rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-primary" title={m.upload_retry()}>
												<Upload size={14} />
											</button>
										{/if}
									{/if}
									{#if item.phase !== 'completed' && item.phase !== 'importing'}
										<button type="button" onclick={() => onDelete(item.uid)} class="rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-danger">
											<X size={14} />
										</button>
									{/if}
								</div>
							</div>
							{#if item.phase === 'failed' && item.errorMsg}
								<p class="relative px-4 pb-2 text-xs text-danger">{item.errorMsg}</p>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<button type="button" onclick={() => (open = true)} class="flex w-full items-center gap-3 rounded-xl border border-line bg-surface px-4 py-2.5 shadow-pop transition-colors hover:bg-surface-sunken">
				{#if hasActive}
					<LoaderCircle size={16} class="animate-spin text-primary" />
				{:else}
					<CircleCheck size={16} class="text-success" />
				{/if}
				<span class="flex-1 text-sm text-ink-2">{completedCount}/{items.length}</span>
				{#if interruptedCount > 0}
					<span class="text-xs text-warning">{m.interrupted_count({ count: interruptedCount })}</span>
				{/if}
				{#if failedCount > 0}
					<span class="text-xs text-danger">{m.upload_panel_failed({ count: failedCount })}</span>
				{/if}
				<ChevronUp size={14} class="text-ink-4" />
			</button>
		{/if}
	</div>
{/if}
