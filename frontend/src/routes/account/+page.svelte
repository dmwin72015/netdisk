<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, setUser, authReady } from '$lib/stores/auth';
	import { getProfile, type ProfileData } from '$lib/api/profile';
	import ProfileCard from '$lib/components/account/ProfileCard.svelte';
	import AccountInfoCard from '$lib/components/account/AccountInfoCard.svelte';
	import StorageCard from '$lib/components/account/StorageCard.svelte';
	import * as m from '$lib/paraglide/messages';

	let profile = $state<ProfileData | null>(null);
	let loading = $state(true);

	onMount(() => {
		if (!browser) return;
		if (!$user) {
			void goto('/login');
			return;
		}
		getProfile()
			.then((p) => {
				profile = p;
				setUser({
					...$user!,
					username: p.username,
					email: p.email,
					storage: p.storage,
					level: p.level,
					createdAt: p.createdAt,
				});
			})
			.catch(() => {})
			.finally(() => (loading = false));
	});

	async function refreshProfile() {
		try {
			const fresh = await getProfile();
			profile = fresh;
			if ($user) {
				setUser({
					...$user,
					username: fresh.username,
					email: fresh.email,
					storage: fresh.storage,
					level: fresh.level,
					createdAt: fresh.createdAt,
				});
			}
		} catch { /* ignore */ }
	}

	let usedBytes = $derived(profile?.storage?.storageUsed ?? 0);
	let quotaBytes = $derived(profile?.storage?.storageQuota ?? 0);
</script>

{#if !$authReady}
{:else if $user}
	<div class="space-y-6">
		<h1 class="text-xl font-semibold">{m.account_center()}</h1>

		<ProfileCard
			displayName={profile?.profile?.displayName ?? ''}
			avatarUrl={profile?.profile?.avatarUrl ?? ''}
			bio={profile?.profile?.bio ?? ''}
			username={$user.username}
			onSaved={refreshProfile}
		/>

		<AccountInfoCard
			username={profile?.username ?? $user.username}
			email={profile?.email ?? $user.email}
			levelName={profile?.level?.levelName ?? $user.level?.levelName ?? null}
			levelExpiresAt={profile?.level?.expiresAt ?? $user.level?.expiresAt ?? null}
			createdAt={profile?.createdAt ?? $user.createdAt}
		/>

		<StorageCard
			{usedBytes}
			{quotaBytes}
			{loading}
		/>
	</div>
{:else}
	<p class="text-gray-600">{@html m.please_login({ link: '<a href="/login" class="underline">' + m.login_link_text() + '</a>' })}</p>
{/if}
