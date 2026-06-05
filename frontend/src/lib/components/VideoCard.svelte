<script lang="ts">
	import { goto } from '$app/navigation';
	import { CircleCheck, CircleX, LoaderCircle, Trash2, Copy, FileVideo, Play } from '@lucide/svelte';
	import type { Task } from '$lib/api/tasks';
	import { deleteTask } from '$lib/api/tasks';
	import { confirmDelete } from '$lib/dialog';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize, fmtDurationHMS, timeAgo, authedUrl, copyToClipboard } from '$lib/utils/format';

	let {
		task,
		onChanged
	}: { task: Task; onChanged?: () => void } = $props();

	let copied = $state(false);
	let removing = $state(false);
	let thumbFailed = $state(false);

	async function copyUrl(url: string) {
		const fullUrl = new URL(url, window.location.origin).toString();
		if (await copyToClipboard(fullUrl)) {
			copied = true;
			setTimeout(() => (copied = false), 1500);
		} else {
			toast.error(m.copy_failed());
		}
	}

	async function remove(e: Event) {
		e.stopPropagation();
		if (!(await confirmDelete(m.confirm_delete_task()))) return;
		removing = true;
		try {
			await deleteTask(task.id);
			onChanged?.();
		} finally {
			removing = false;
		}
	}

	function open() {
		if (task.status === 'completed') void goto(`/videos/${task.id}`);
	}

	const duration = $derived(fmtDurationHMS(task.durationSec));
	const thumbnailSrc = $derived(task.thumbnailUrl ? authedUrl(task.thumbnailUrl) : null);
</script>

<article
	class="group overflow-hidden rounded-xl bg-white transition hover:shadow-lg {task.status ===
		'completed'
		? 'cursor-pointer'
		: ''}"
>
	<div
		role="button"
		tabindex="0"
		onclick={open}
		onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && open()}
		class="relative aspect-video w-full overflow-hidden bg-slate-900"
	>
		{#if task.status === 'completed' && thumbnailSrc && !thumbFailed}
			<img
				src={thumbnailSrc}
				alt={task.originalName}
				loading="lazy"
				onerror={() => (thumbFailed = true)}
				class="h-full w-full object-cover transition group-hover:scale-105"
			/>
			<div
				class="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/0 transition group-hover:bg-black/30"
			>
				<Play size={48} class="text-white opacity-0 transition group-hover:opacity-100" />
			</div>
			{#if duration}
				<span class="absolute bottom-2 right-2 rounded bg-black/80 px-1.5 py-0.5 text-xs font-medium text-white">
					{duration}
				</span>
			{/if}
		{:else if task.status === 'failed'}
			<div class="flex h-full w-full flex-col items-center justify-center gap-2 bg-red-50 text-red-700">
				<CircleX size={32} />
				<span class="text-xs">{m.conversion_failed()}</span>
			</div>
		{:else if task.status === 'processing' || task.status === 'pending'}
			<div class="flex h-full w-full flex-col items-center justify-center gap-3 bg-slate-100 text-slate-600">
				<LoaderCircle size={32} class="animate-spin" />
				<span class="text-xs">{task.status === 'pending' ? m.queued() : m.converting_progress({ progress: task.progress })}</span>
			</div>
		{:else}
			<div class="flex h-full w-full items-center justify-center bg-slate-200 text-slate-400">
				<FileVideo size={36} />
			</div>
		{/if}
	</div>

	<div class="p-3">
		<header class="flex items-start gap-2">
			<h3 class="min-w-0 flex-1 truncate text-sm font-medium text-slate-900" title={task.originalName}>
				{task.originalName}
			</h3>
			<button
				type="button"
				onclick={remove}
				disabled={removing}
				class="shrink-0 text-slate-400 transition hover:text-red-600"
				aria-label={m.delete_label()}
				title={m.delete_label()}
			>
				<Trash2 size={14} />
			</button>
		</header>

		<div class="mt-1 flex items-center justify-between gap-2 text-xs text-slate-500">
			<span>{fmtSize(task.fileSize)}</span>
			<span>{timeAgo(task.createdAt)}</span>
		</div>

		{#if task.status === 'processing' || task.status === 'pending'}
			<div class="mt-2 h-1 w-full overflow-hidden rounded bg-slate-200">
				<div class="h-full bg-slate-900 transition-all" style="width:{task.progress}%"></div>
			</div>
		{:else if task.status === 'failed' && task.errorMessage}
			<p class="mt-2 break-all text-xs text-red-600">{task.errorMessage}</p>
		{:else if task.status === 'completed' && task.m3u8Url}
			<div class="mt-2 flex items-center gap-3 text-xs">
				<a href="/videos/{task.id}" class="flex items-center gap-1 text-slate-700 hover:text-slate-900">
					<CircleCheck size={12} /> {m.play_label()}
				</a>
				<button
					type="button"
					onclick={(e) => {
						e.stopPropagation();
						copyUrl(task.m3u8Url!);
					}}
					class="flex items-center gap-1 text-slate-500 hover:text-slate-900"
				>
					<Copy size={12} /> {copied ? m.copied() : m.copy_url()}
				</button>
			</div>
		{/if}
	</div>
</article>
