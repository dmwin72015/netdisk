<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { ChevronLeft, LoaderCircle, Globe, Mail } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { adminGetUser, type AdminUser } from '$lib/api/admin';
	import { fmtSize, timeAgo } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let user = $state<AdminUser | null>(null);
	let loading = $state(true);

	let userId = $derived($page.params.id);

	onMount(() => {
		if (!browser) return;
		loadUser();
	});

	async function loadUser() {
		if (!userId) return;
		loading = true;
		try {
			user = await adminGetUser(userId);
		} catch {
			toast.error(m.load_failed());
		} finally {
			loading = false;
		}
	}

	function fmtDate(ts: number): string {
		return new Date(ts * 1000).toLocaleString();
	}
</script>

{#if loading}
	<div class="flex justify-center py-16">
		<LoaderCircle size={24} class="animate-spin text-ink-4" />
	</div>
{:else if user}
	<div class="space-y-6">
		<div class="flex items-center gap-2">
			<button
				onclick={() => goto('/admin')}
				class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
			>
				<ChevronLeft size={20} />
			</button>
			<h1 class="text-xl font-semibold text-ink-2">{user.username}</h1>
			<span class="rounded-full bg-surface-sunken px-2.5 py-0.5 text-xs font-medium text-ink-3">{user.role}</span>
		</div>

		<div class="grid gap-6 md:grid-cols-2">
			<!-- Basic Info -->
			<div class="rounded-lg border bg-white p-5">
				<h2 class="mb-4 text-sm font-semibold text-ink-3 uppercase tracking-wide">Basic Info</h2>
				<dl class="space-y-3 text-sm">
					<div class="flex justify-between">
						<dt class="text-ink-3">ID</dt>
						<dd class="font-mono text-ink">{user.id}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.username()}</dt>
						<dd class="text-ink">{user.username}</dd>
					</div>
					<div class="flex justify-between items-center">
						<dt class="text-ink-3">{m.email()}</dt>
						<dd class="flex items-center gap-1 text-ink">
							<Mail size={13} class="text-ink-4" />
							{user.email}
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.col_role()}</dt>
						<dd>
							<span class="rounded bg-surface-sunken px-2 py-0.5 text-xs font-medium">{user.role}</span>
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">Register Method</dt>
						<dd class="flex items-center gap-1">
							{#if user.registerMethod === 'email'}
								<Mail size={13} class="text-ink-4" />
							{:else}
								<Globe size={13} class="text-ink-4" />
							{/if}
							{user.registerMethod}
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.joined()}</dt>
						<dd class="text-ink">{fmtDate(user.createdAt)}</dd>
					</div>
				</dl>
			</div>

			<!-- Storage -->
			<div class="rounded-lg border bg-white p-5">
				<h2 class="mb-4 text-sm font-semibold text-ink-3 uppercase tracking-wide">{m.drive_storage()}</h2>
				<dl class="space-y-3 text-sm">
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.used()}</dt>
						<dd class="text-ink">{fmtSize(user.usedBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.storage_base()}</dt>
						<dd class="text-ink">{fmtSize(user.baseBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.storage_bonus()}</dt>
						<dd class="text-ink">{fmtSize(user.memberBonusBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.storage_pack()}</dt>
						<dd class="text-ink">{fmtSize(user.packBytes)}</dd>
					</div>
					<hr class="border-line-soft" />
					<div class="flex justify-between font-medium">
						<dt class="text-ink-2">{m.col_storage_limit()}</dt>
						<dd class="text-ink">{fmtSize(user.totalBytes)}</dd>
					</div>
				</dl>
			</div>

			<!-- Profile -->
			{#if user.profile}
				<div class="rounded-lg border bg-white p-5">
					<h2 class="mb-4 text-sm font-semibold text-ink-3 uppercase tracking-wide">{m.profile_info()}</h2>
					<dl class="space-y-3 text-sm">
						<div class="flex justify-between">
							<dt class="text-ink-3">{m.nickname_label()}</dt>
							<dd class="text-ink">{user.profile.displayName || '-'}</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-ink-3">{m.bio_label()}</dt>
							<dd class="text-right text-ink">{user.profile.bio || '-'}</dd>
						</div>
						{#if user.profile.avatarUrl}
							<div class="flex justify-between items-center">
								<dt class="text-ink-3">Avatar</dt>
								<dd>
										<img src={user.profile.avatarUrl} alt="avatar" loading="lazy" class="h-10 w-10 rounded-full object-cover" />
								</dd>
							</div>
						{/if}
					</dl>
				</div>
			{/if}

			<!-- OAuth Accounts -->
			{#if user.oauthAccounts && user.oauthAccounts.length > 0}
				<div class="rounded-lg border bg-white p-5">
					<h2 class="mb-4 text-sm font-semibold text-ink-3 uppercase tracking-wide">OAuth Accounts</h2>
					<div class="space-y-3">
						{#each user.oauthAccounts as oa}
							<div class="flex items-center justify-between rounded-md bg-surface-muted px-3 py-2 text-sm">
								<div class="flex items-center gap-2">
									<Globe size={14} class="text-ink-4" />
									<span class="font-medium text-ink-2">{oa.provider}</span>
								</div>
								<div class="text-xs text-ink-4">{timeAgo(oa.createdAt)}</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
