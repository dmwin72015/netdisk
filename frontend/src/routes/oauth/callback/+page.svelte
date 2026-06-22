<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { api, setSession, type UserInfo } from '$lib/api/client';
	import { setUser } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { LoaderCircle } from '@lucide/svelte';

	let status = $state('Logging in...');

	onMount(async () => {
		if (!browser) return;

		const accessToken = page.url.searchParams.get('accessToken');
		const refreshToken = page.url.searchParams.get('refreshToken');
		const error = page.url.searchParams.get('error');
		const mode = page.url.searchParams.get('mode');
		const provider = page.url.searchParams.get('provider') ?? '';

		if (mode === 'bind') {
			if (window.opener) {
				const msg: Record<string, unknown> = { bound: true, provider };
				if (error) msg.error = error;
				const needReplaceConfirm = page.url.searchParams.get('needReplaceConfirm');
				if (needReplaceConfirm) msg.needReplaceConfirm = true;
				const alreadyBound = page.url.searchParams.get('alreadyBound');
				if (alreadyBound) msg.alreadyBound = true;
				const replaceToken = page.url.searchParams.get('replaceToken');
				if (replaceToken) msg.replaceToken = replaceToken;
				const replaceProvider = page.url.searchParams.get('replaceProvider');
				if (replaceProvider) msg.replaceProvider = replaceProvider;
				const oldProviderAccountId = page.url.searchParams.get('oldProviderAccountId');
				if (oldProviderAccountId) msg.oldProviderAccountId = oldProviderAccountId;
				const oldOauthEmail = page.url.searchParams.get('oldOauthEmail');
				if (oldOauthEmail) msg.oldOauthEmail = oldOauthEmail;
				window.opener.postMessage(msg, location.origin);
				return;
			}
			status = error || 'Account linked successfully';
			return;
		}

		if (mode === 'email-confirm') {
			if (window.opener) {
				window.opener.postMessage(
					{
						emailMatchConfirm: true,
						confirmToken: page.url.searchParams.get('confirmToken'),
						email: page.url.searchParams.get('email'),
						provider: provider,
					},
					location.origin,
				);
				return;
			}
			status = m.oauth_email_registered({ email: page.url.searchParams.get('email') ?? '' });
			return;
		}

		if (error) {
			status = error;
			return;
		}

		if (!accessToken || !refreshToken) {
			status = 'Invalid response from provider';
			return;
		}

		// If opened in a popup, send tokens to the opener and close
		if (window.opener) {
			window.opener.postMessage({ accessToken, refreshToken }, location.origin);
			window.close();
			return;
		}

		setSession(null, { accessToken, refreshToken, expiresIn: 3600 });

		try {
			const data = await api<UserInfo>('/api/v1/user/me', {
				method: 'GET',
			});
			setUser(data);
			await goto('/');
		} catch {
			status = 'Failed to complete login';
		}
	});
</script>

<div class="flex min-h-screen items-center justify-center">
	<div class="text-center">
		<LoaderCircle size={32} class="mx-auto mb-4 animate-spin text-primary" />
		<p class="text-sm text-ink-3">{status}</p>
	</div>
</div>
