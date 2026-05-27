<script lang="ts">
	import { onDestroy } from 'svelte';
	import { CheckCircle2, XCircle, Loader2, Copy } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let {
		taskId,
		onDone
	}: { taskId: string; onDone?: (m3u8?: string) => void } = $props();

	type Frame = {
		task_id: string;
		status: 'pending' | 'processing' | 'completed' | 'failed';
		progress: number;
		m3u8_url?: string;
		error?: string;
	};

	let frame = $state<Frame | null>(null);
	let copied = $state(false);
	let es: EventSource | null = null;

	function connect() {
		// EventSource does not support custom headers, so the access token is
		// passed as a query parameter the backend accepts. For now we just open
		// the stream and rely on cookies/proxy auth in production. In the dev
		// environment we tunnel through Vite which forwards Authorization-bearing
		// cookies it does not, so we send the token via query param fallback:
		const token = localStorage.getItem('vf.access') ?? '';
		const url = `/api/v1/tasks/${taskId}/progress${token ? `?access_token=${encodeURIComponent(token)}` : ''}`;
		es = new EventSource(url);
		const handler = (e: MessageEvent) => {
			try {
				frame = JSON.parse(e.data);
			} catch {
				/* ignore */
			}
		};
		es.addEventListener('progress', handler);
		es.addEventListener('done', (e) => {
			handler(e as MessageEvent);
			es?.close();
			onDone?.(frame?.m3u8_url);
		});
		es.onerror = () => {
			// EventSource auto-reconnects when readyState is CONNECTING. Only
			// close if the browser has already given up (readyState=CLOSED).
			if (es && es.readyState === EventSource.CLOSED) {
				es = null;
			}
		};
	}

	$effect(() => {
		// re-subscribe when taskId changes
		es?.close();
		frame = null;
		copied = false;
		if (taskId) connect();
		return () => es?.close();
	});

	onDestroy(() => es?.close());

	async function copyUrl() {
		if (!frame?.m3u8_url) return;
		await navigator.clipboard.writeText(frame.m3u8_url);
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}
</script>

<div class="rounded border bg-white p-4">
	{#if !frame}
		<div class="flex items-center gap-2 text-sm text-slate-600">
			<Loader2 class="animate-spin" size={16} />
			<span>{m.connecting()}</span>
		</div>
	{:else if frame.status === 'completed'}
		<div class="flex items-center gap-2 text-sm text-emerald-700">
			<CheckCircle2 size={16} /> <span>{m.conversion_done()}</span>
		</div>
		{#if frame.m3u8_url}
			<div class="mt-3 flex items-center gap-2">
				<code class="flex-1 truncate rounded bg-slate-100 px-2 py-1 text-xs">{frame.m3u8_url}</code>
				<button
					type="button"
					onclick={copyUrl}
					class="flex items-center gap-1 rounded border px-2 py-1 text-xs hover:bg-slate-50"
				>
					<Copy size={12} /> {copied ? m.copied() : m.copy()}
				</button>
			</div>
		{/if}
	{:else if frame.status === 'failed'}
		<div class="flex items-center gap-2 text-sm text-red-700">
			<XCircle size={16} />
			<span>{m.conversion_failed()}{frame.error ? ': ' + frame.error : ''}</span>
		</div>
	{:else}
		<div class="flex items-center justify-between text-sm">
			<div class="flex items-center gap-2 text-slate-700">
				<Loader2 class="animate-spin" size={16} />
				<span>{frame.status === 'pending' ? m.queued() : m.converting()}</span>
			</div>
			<span class="text-slate-500">{frame.progress}%</span>
		</div>
		<div class="mt-2 h-2 w-full overflow-hidden rounded bg-slate-200">
			<div class="h-full bg-slate-900 transition-all" style="width:{frame.progress}%"></div>
		</div>
	{/if}
</div>
