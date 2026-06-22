<script lang="ts">
	import { onMount } from 'svelte';
	import { User, Shield, Calendar, Crown, Link as LinkIcon, Plus, X } from '@lucide/svelte';
	import * as TooltipBase from '$lib/ui/tooltip/base';
	import AlertDialog from '$lib/ui/alert-dialog/AlertDialog.svelte';
	import type { OAuthAccountInfo } from '$lib/api/profile';
	import { unlinkOAuth } from '$lib/api/profile';
	import { ApiError } from '$lib/api/client';
	import { getAccessToken } from '$lib/api/client';
	import { toast } from 'svelte-sonner';
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
	let replaceTarget = $state<{
		provider: string;
		token: string;
		oldProviderAccountId?: string;
		oldOauthEmail?: string;
	} | null>(null);
	let replaceDialogOpen = $state(false);
	let replaceBusy = $state(false);

	async function confirmUnlink() {
		if (!unlinkTarget) return;
		unlinkBusy = true;
		unlinkError = null;
		try {
			const target = unlinkTarget;
			await unlinkOAuth(target);
			unlinkTarget = null;
			onRefresh?.();
			toast.success(`已成功解绑 ${knownProviders[target]?.name ?? target}`);
		} catch (err) {
			unlinkError = err instanceof ApiError ? err.message : 'Failed to unlink';
			toast.error(unlinkError);
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

	function closeBindPopup() {
		if (bindPopup && !bindPopup.closed) {
			bindPopup.close();
		}
		bindPopup = null;
	}

	function onMessage(event: MessageEvent) {
		if (event.origin !== location.origin) return;
		const data = event.data ?? {};
		const {
			needReplaceConfirm,
			replaceToken,
			replaceProvider,
			oldProviderAccountId,
			oldOauthEmail,
			bound,
			provider,
			error,
			alreadyBound
		} = data as {
			needReplaceConfirm?: boolean;
			replaceToken?: string;
			replaceProvider?: string;
			oldProviderAccountId?: string;
			oldOauthEmail?: string;
			bound?: boolean;
			provider?: string;
			error?: string | null;
			alreadyBound?: boolean;
		};

		if (needReplaceConfirm) {
			closeBindPopup();
			bindError = null;
			replaceTarget = {
				provider: replaceProvider ?? provider ?? '',
				token: replaceToken ?? '',
				oldProviderAccountId,
				oldOauthEmail
			};
			replaceDialogOpen = true;
			return;
		}

		if (bound === undefined) return;

		// Any terminal bind message (success, already-bound, or error) must close
		// the popup so the user is not stuck looking at the OAuth provider page.
		closeBindPopup();

		const providerName = knownProviders[provider ?? '']?.name ?? provider ?? '';

		if (error) {
			bindError = error;
			toast.error(error);
			return;
		}

		bindError = null;
		if (alreadyBound) {
			toast.info(`该 ${providerName} 账号已绑定到当前账号`);
		} else {
			toast.success(`已成功绑定 ${providerName}`);
		}
		onRefresh?.();
	}

	async function confirmReplace() {
		if (!replaceTarget) return;
		replaceBusy = true;
		bindError = null;
		try {
			const response = await fetch(
				`/api/v1/auth/oauth/bind/confirm-replace?token=${encodeURIComponent(replaceTarget.token)}`,
				{
					method: 'POST',
					headers: {
						'Authorization': `Bearer ${getAccessToken()}`,
					},
				}
			);
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw new ApiError(data.message || 'Failed to replace binding', response.status, 0);
			}
			const providerName =
				knownProviders[replaceTarget?.provider ?? '']?.name ?? replaceTarget?.provider ?? '';
			replaceTarget = null;
			replaceDialogOpen = false;
			onRefresh?.();
			toast.success(`已成功绑定 ${providerName}`);
		} catch (err) {
			bindError = err instanceof ApiError ? err.message : 'Failed to replace binding';
			toast.error(bindError);
		} finally {
			replaceBusy = false;
		}
	}

	onMount(() => {
		window.addEventListener('message', onMessage);
		return () => window.removeEventListener('message', onMessage);
	});
</script>

<div class="rounded-xl border border-line-soft bg-white p-6 ">
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

	<hr class="my-4 border-line-soft" />
	<h3 class="mb-3 text-sm font-medium text-ink-3">{m.linked_accounts()}</h3>
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
							<p class="mt-1 break-all text-ink-4">ID: {acc.providerAccountId}</p>
							{#if acc.oauthEmail}
								<p class="break-all text-ink-4">{acc.oauthEmail}</p>
							{/if}
							<p class="mt-1 text-ink-4">{m.joined()}: {fmtDate(acc.createdAt)}</p>
						</TooltipBase.Content>
					</TooltipBase.Root>
				{/each}
			</TooltipBase.Provider>
		{/if}

		{#each unlinkedProviders as provider}
			{@const p = knownProviders[provider]}
			<button
				onclick={() => oauthBind(provider)}
				class="inline-flex cursor-pointer items-center gap-1.5 rounded-full border border-dashed border-line px-3 py-1 text-xs font-medium text-ink-3 transition hover:border-ink-4 hover:text-ink-2"
			>
				<Plus size={12} />
				{p.name}
			</button>
		{/each}
	</div>

	{#if unlinkError}
		<p class="mt-2 text-xs text-danger">{unlinkError}</p>
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

<AlertDialog
	bind:open={replaceDialogOpen}
	title="Replace {knownProviders[replaceTarget?.provider ?? '']?.name ?? replaceTarget?.provider ?? ''} binding"
	description={(() => {
		const ident = replaceTarget?.oldOauthEmail || replaceTarget?.oldProviderAccountId;
		return ident
			? `This account is already linked to another ${knownProviders[replaceTarget?.provider ?? '']?.name ?? replaceTarget?.provider ?? ''} account (${ident}). Replacing it will unbind that account and link the new one to your current account.`
			: `This account is already linked to another ${knownProviders[replaceTarget?.provider ?? '']?.name ?? replaceTarget?.provider ?? ''} account. Replacing it will unbind that account and link the new one to your current account.`;
	})()}
	confirmText={replaceBusy ? 'Replacing...' : 'Replace'}
	variant="destructive"
	onConfirm={confirmReplace}
	onCancel={() => { replaceTarget = null; replaceDialogOpen = false; }}
/>
