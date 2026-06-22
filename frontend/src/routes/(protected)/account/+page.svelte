<script lang="ts">
	import { onMount } from 'svelte';
	import { user, setUser, authReady } from '$lib/stores/auth';
	import { getProfile, type ProfileData } from '$lib/api/profile';
	import ProfileCard from '$lib/components/account/ProfileCard.svelte';
	import AccountInfoCard from '$lib/components/account/AccountInfoCard.svelte';
	import StorageCard from '$lib/components/account/StorageCard.svelte';
	import * as m from '$lib/paraglide/messages';

	let profile = $state<ProfileData | null>(null);
	let loading = $state(true);

	onMount(() => {
		getProfile()
			.then((p) => {
				profile = p;
				setUser({
					...$user!,
					username: p.username,
					email: p.email,
					profile: p.profile,
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
					profile: fresh.profile,
					storage: fresh.storage,
					level: fresh.level,
					createdAt: fresh.createdAt,
				});
			}
		} catch { /* ignore */ }
	}

	let usedBytes = $derived(profile?.storage?.storageUsed ?? 0);
	let quotaBytes = $derived(profile?.storage?.storageQuota ?? 0);
	let oauthAccounts = $derived(profile?.oauthAccounts ?? []);
	const appVersion = __APP_VERSION__;
	const appBuildTime = __APP_BUILD_TIME__;
</script>

{#if $authReady && $user}
	<div class="space-y-6">
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
			{oauthAccounts}
			onRefresh={refreshProfile}
		/>

		<StorageCard
			{usedBytes}
			{quotaBytes}
			{loading}
		/>

		<p class="pb-2 pt-4 text-center text-xs text-ink-4" title={appBuildTime}>
			{m.app_version({ version: appVersion })}
		</p>
	</div>
{/if}
