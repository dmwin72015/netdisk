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

	let userItem = $state<AdminUser | null>(null);
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
			userItem = await adminGetUser(userId);
		} catch {
			toast.error(m.admin_load_failed());
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
{:else if userItem}
	<div class="space-y-6">
		<div class="flex items-center gap-2">
			<button
				onclick={() => goto('/admin/users')}
				class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
			>
				<ChevronLeft size={20} />
			</button>
			<h1 class="text-xl font-bold text-ink">{userItem.username}</h1>
			<span class="rounded-full bg-surface-sunken px-2.5 py-0.5 text-xs font-medium text-ink-3">{userItem.role}</span>
		</div>

		<div class="grid gap-6 md:grid-cols-2">
			<div class="rounded-xl border border-line bg-surface p-5">
				<h2 class="mb-4 text-xs font-semibold uppercase tracking-wide text-ink-4">{m.admin_basic_info()}</h2>
				<dl class="space-y-3 text-sm">
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.id()}</dt>
						<dd class="font-mono text-ink">{userItem.id}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.username()}</dt>
						<dd class="text-ink">{userItem.username}</dd>
					</div>
					<div class="flex items-center justify-between">
						<dt class="text-ink-3">{m.email()}</dt>
						<dd class="flex items-center gap-1 text-ink">
							<Mail size={13} class="text-ink-4" />
							{userItem.email}
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.col_role()}</dt>
						<dd>
							<span class="rounded bg-surface-sunken px-2 py-0.5 text-xs font-medium text-ink-3">{userItem.role}</span>
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.admin_register_method()}</dt>
						<dd class="flex items-center gap-1 text-ink">
							<Globe size={13} class="text-ink-4" />
							{userItem.registerMethod}
						</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.admin_joined()}</dt>
						<dd class="text-ink">{fmtDate(userItem.createdAt)}</dd>
					</div>
				</dl>
			</div>

			<div class="rounded-xl border border-line bg-surface p-5">
				<h2 class="mb-4 text-xs font-semibold uppercase tracking-wide text-ink-4">{m.admin_storage_usage()}</h2>
				<dl class="space-y-3 text-sm">
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.admin_used()}</dt>
						<dd class="text-ink">{fmtSize(userItem.usedBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">Base</dt>
						<dd class="text-ink">{fmtSize(userItem.baseBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.admin_member_bonus()}</dt>
						<dd class="text-ink">{fmtSize(userItem.memberBonusBytes)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-ink-3">{m.admin_pack()}</dt>
						<dd class="text-ink">{fmtSize(userItem.packBytes)}</dd>
					</div>
					<hr class="border-line-soft" />
					<div class="flex justify-between font-medium">
						<dt class="text-ink-2">{m.admin_total()}</dt>
						<dd class="text-ink">{fmtSize(userItem.totalBytes)}</dd>
					</div>
				</dl>
			</div>

			{#if userItem.profile}
				<div class="rounded-xl border border-line bg-surface p-5">
					<h2 class="mb-4 text-xs font-semibold uppercase tracking-wide text-ink-4">{m.admin_profile()}</h2>
					<dl class="space-y-3 text-sm">
						<div class="flex justify-between">
							<dt class="text-ink-3">{m.admin_display_name()}</dt>
							<dd class="text-ink">{userItem.profile.displayName || '-'}</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-ink-3">{m.admin_bio()}</dt>
							<dd class="text-right text-ink">{userItem.profile.bio || '-'}</dd>
						</div>
						{#if userItem.profile.avatarUrl}
							<div class="flex items-center justify-between">
								<dt class="text-ink-3">{m.admin_avatar()}</dt>
								<dd>
									<img src={userItem.profile.avatarUrl} alt="avatar" loading="lazy" class="h-10 w-10 rounded-full object-cover" />
								</dd>
							</div>
						{/if}
					</dl>
				</div>
			{/if}

			{#if userItem.oauthAccounts && userItem.oauthAccounts.length > 0}
				<div class="rounded-xl border border-line bg-surface p-5">
					<h2 class="mb-4 text-xs font-semibold uppercase tracking-wide text-ink-4">{m.admin_oauth_accounts()}</h2>
					<div class="space-y-3">
						{#each userItem.oauthAccounts as oa}
							<div class="flex items-center justify-between rounded-lg bg-surface-sunken px-3 py-2 text-sm">
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
