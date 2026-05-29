<script lang="ts">
	import { FileVideo, X } from '@lucide/svelte';
	import type { UploadSession } from '$lib/api/uploads';
	import * as m from '$lib/paraglide/messages';

	let {
		session,
		busy,
		onResume,
		onCancel
	}: {
		session: UploadSession;
		busy?: boolean;
		onResume: (session: UploadSession) => void;
		onCancel: (session: UploadSession) => void;
	} = $props();

	function fmtSize(size: number): string {
		if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
		if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
		return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`;
	}

	function timeAgo(unix: number): string {
		const diff = Math.max(0, Date.now() / 1000 - unix);
		if (diff < 60) return m.just_now();
		if (diff < 3600) return m.minutes_ago({ n: Math.floor(diff / 60) });
		if (diff < 86400) return m.hours_ago({ n: Math.floor(diff / 3600) });
		return m.days_ago({ n: Math.floor(diff / 86400) });
	}

	const pct = $derived(Math.round((session.receivedBytes / session.totalSize) * 100));
</script>

<article class="rounded-lg border border-amber-200 bg-amber-50 p-3">
	<header class="flex items-center gap-2 text-sm">
		<FileVideo size={16} class="text-amber-700" />
		<span class="min-w-0 flex-1 truncate font-medium" title={session.filename}>
			{session.filename}
		</span>
		<button
			type="button"
			onclick={() => onCancel(session)}
			disabled={busy}
			class="text-slate-400 hover:text-red-600 disabled:opacity-40"
			aria-label={m.cancel()}
			title={m.cancel_upload_title()}
		>
			<X size={14} />
		</button>
	</header>
	<div class="mt-2 flex items-center justify-between text-xs text-slate-600">
		<span>{fmtSize(session.receivedBytes)} / {fmtSize(session.totalSize)}</span>
		<span>{timeAgo(session.updatedAt)}</span>
	</div>
	<div class="mt-2 h-1.5 w-full overflow-hidden rounded bg-amber-200">
		<div class="h-full bg-amber-600 transition-all" style="width:{pct}%"></div>
	</div>
	<div class="mt-2 flex items-center justify-between text-xs">
		<span class="text-slate-600">{pct}%</span>
		<button
			type="button"
			onclick={() => onResume(session)}
			disabled={busy}
			class="rounded bg-amber-700 px-2 py-1 text-white hover:bg-amber-800 disabled:opacity-50"
		>
			{m.select_resume()}
		</button>
	</div>
</article>
