<script lang="ts">
	import { onMount } from 'svelte';
	import { User, Shield, Calendar, Crown, Link as LinkIcon, Plus, X } from '@lucide/svelte';
	import * as TooltipBase from '$lib/ui/tooltip/base';
	import AlertDialog from '$lib/ui/alert-dialog/AlertDialog.svelte';
	import type { OAuthAccountInfo } from '$lib/api/profile';
	import { unlinkOAuth } from '$lib/api/profile';
	import { ApiError } from '$lib/api/client';
	import { getAccessToken } from '$lib/api/client';
	import * as m from '$lib/paraglide/messages';

	let {
		username,
		email,
		levelName,
		levelExpiresAt,
		createdAt,
		oauthAccounts = [],
		onRefresh
	}: {
		username: string;
		email: string;
		levelName: string | null;
		levelExpiresAt: string | null;
		createdAt: string;
		oauthAccounts?: OAuthAccountInfo[];
		onRefresh?: () => void;
	} = $props();

	function fmtDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString();
		} catch {
			return iso;
		}
	}

	const knownProviders: Record<string, { name: string; color: string }> = {
		github: { name: 'GitHub', color: '#24292e' },
		'2libra': { name: '2libra', color: '#6366f1' },
	};

	const allProviders = ['github', '2libra'] as const;
	let linkedProviders = $derived(new Set(oauthAccounts.map((a) => a.provider)));
	let unlinkedProviders = $derived(allProviders.filter((p) => !linkedProviders.has(p)));

	let bindError = $state<string | null>(null);
	let bindPopup: Window | null = null;
	let unlinkError = $state<string | null>(null);
	let unlinkTarget = $state<string | null>(null);
	let unlinkDialogOpen = $state(false);
	let unlinkBusy = $state(false);

	async function confirmUnlink() {
		if (!unlinkTarget) return;
		unlinkBusy = true;
		unlinkError = null;
		try {
			await unlinkOAuth(unlinkTarget);
			unlinkTarget = null;
			onRefresh?.();
		} catch (err) {
			unlinkError = err instanceof ApiError ? err.message : 'Failed to unlink';
		} finally {
			unlinkBusy = false;
		}
	}

	function oauthBind(provider: string) {
		bindError = null;
		const width = 520;
		const height = 600;
		const left = (screen.width - width) / 2;
		const top = (screen.height - height) / 2;
		const token = getAccessToken();
		bindPopup = window.open(
			`/api/v1/auth/oauth/${provider}/bind?access_token=${token}`,
			'oauth-bind',
			`width=${width},height=${height},left=${left},top=${top},popup=1`,
		);
	}

	function onMessage(event: MessageEvent) {
		if (event.origin !== location.origin) return;
		const { bound, error } = event.data ?? {};
		if (bound === undefined) return;
		if (bindPopup && !bindPopup.closed) {
			bindPopup.close();
		}
		bindPopup = null;
		if (error) {
			bindError = error;
			return;
		}
		bindError = null;
		onRefresh?.();
	}

	onMount(() => {
		window.addEventListener('message', onMessage);
		return () => window.removeEventListener('message', onMessage);
	});
</script>

<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
	<h2 class="mb-4 text-sm font-medium text-gray-500">{m.account_info()}</h2>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
				<User size={16} class="text-gray-400" />
			</div>
			<div>
				<p class="text-xs text-gray-400">{m.username_label()}</p>
				<p class="text-sm font-medium text-gray-800">{username}</p>
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
				<Shield size={16} class="text-gray-400" />
			</div>
			<div>
				<p class="text-xs text-gray-400">{m.email_label()}</p>
				<p class="text-sm font-medium text-gray-800">{email || '-'}</p>
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
				<Crown size={16} class="text-gray-400" />
			</div>
			<div>
				<p class="text-xs text-gray-400">{m.level()}</p>
				<p class="text-sm font-medium text-gray-800">{levelName || '-'}</p>
				{#if levelExpiresAt}
					<p class="text-xs text-gray-400">{m.level_expires({ date: fmtDate(levelExpiresAt) })}</p>
				{/if}
			</div>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
				<Calendar size={16} class="text-gray-400" />
			</div>
			<div>
				<p class="text-xs text-gray-400">{m.joined()}</p>
				<p class="text-sm font-medium text-gray-800">{fmtDate(createdAt)}</p>
			</div>
		</div>
	</div>

	<hr class="my-4 border-gray-100" />
	<h3 class="mb-3 text-sm font-medium text-gray-500">{m.linked_accounts()}</h3>
	<div class="flex flex-wrap items-center gap-2">
		{#if oauthAccounts.length > 0}
			<TooltipBase.Provider>
				{#each oauthAccounts as acc}
					{@const p = knownProviders[acc.provider] ?? { name: acc.provider, color: '#6b7280' }}
					<TooltipBase.Root delayDuration={0}>
						<TooltipBase.Trigger>
							<span class="group inline-flex cursor-default items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium text-white"
								style="background-color: {p.color}"
							>
								<LinkIcon size={12} />
								{p.name}
							<button
								onclick={(e) => { e.stopPropagation(); unlinkTarget = acc.provider; unlinkDialogOpen = true; }}
								class="ml-0.5 rounded-full p-0.5 opacity-0 transition hover:bg-white/20 group-hover:opacity-100"
								>
									<X size={10} />
								</button>
							</span>
						</TooltipBase.Trigger>
						<TooltipBase.Content class="w-56 px-3 py-2 text-left">
							<p class="font-medium text-white">{p.name}</p>
							<p class="mt-1 break-all text-gray-300">ID: {acc.providerAccountId}</p>
							{#if acc.oauthEmail}
								<p class="break-all text-gray-300">{acc.oauthEmail}</p>
							{/if}
							<p class="mt-1 text-gray-400">{m.joined()}: {fmtDate(acc.createdAt)}</p>
						</TooltipBase.Content>
					</TooltipBase.Root>
				{/each}
			</TooltipBase.Provider>
		{/if}

		{#each unlinkedProviders as provider}
			{@const p = knownProviders[provider]}
			<button
				onclick={() => oauthBind(provider)}
				class="inline-flex cursor-pointer items-center gap-1.5 rounded-full border border-dashed border-gray-300 px-3 py-1 text-xs font-medium text-gray-500 transition hover:border-gray-400 hover:text-gray-700"
			>
				<Plus size={12} />
				{p.name}
			</button>
		{/each}
	</div>

	{#if bindError}
		<p class="mt-2 text-xs text-red-500">{bindError}</p>
	{/if}
	{#if unlinkError}
		<p class="mt-2 text-xs text-red-500">{unlinkError}</p>
	{/if}
</div>

<AlertDialog
	bind:open={unlinkDialogOpen}
	title="Unlink {knownProviders[unlinkTarget ?? '']?.name ?? unlinkTarget ?? ''}?"
	description={!email ? 'You have no email bound. After unlinking, you will only be able to log in via other linked OAuth accounts.' : undefined}
	confirmText={unlinkBusy ? 'Unlinking...' : 'Unlink'}
	variant="destructive"
	onConfirm={confirmUnlink}
	onCancel={() => { unlinkTarget = null; unlinkDialogOpen = false; }}
/>
