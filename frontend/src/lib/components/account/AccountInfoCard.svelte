<script lang="ts">
	import { User, Shield, Calendar, Crown } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let {
		username,
		email,
		levelName,
		levelExpiresAt,
		createdAt,
	}: {
		username: string;
		email: string;
		levelName: string | null;
		levelExpiresAt: string | null;
		createdAt: string;
	} = $props();

	function fmtDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString();
		} catch {
			return iso;
		}
	}
</script>

<div class="rounded-xl border border-line-soft bg-white p-6">
	<h2 class="mb-4 text-sm font-medium text-ink-3">{m.account_info()}</h2>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-surface-muted">
				<User size={16} class="text-ink-4" />
			</div>
			<div>
				<p class="text-xs text-ink-4">{m.username_label()}</p>
				<p class="text-sm font-medium text-ink-2">{username}</p>
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-surface-muted">
				<Shield size={16} class="text-ink-4" />
			</div>
			<div>
				<p class="text-xs text-ink-4">{m.email_label()}</p>
				<p class="text-sm font-medium text-ink-2">{email || '-'}</p>
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-surface-muted">
				<Crown size={16} class="text-ink-4" />
			</div>
			<div>
				<p class="text-xs text-ink-4">{m.level()}</p>
				<p class="text-sm font-medium text-ink-2">{levelName || '-'}</p>
				{#if levelExpiresAt}
					<p class="text-xs text-ink-4">{m.level_expires({ date: fmtDate(levelExpiresAt) })}</p>
				{/if}
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-surface-muted">
				<Calendar size={16} class="text-ink-4" />
			</div>
			<div>
				<p class="text-xs text-ink-4">{m.joined()}</p>
				<p class="text-sm font-medium text-ink-2">{fmtDate(createdAt)}</p>
			</div>
		</div>
	</div>
</div>
