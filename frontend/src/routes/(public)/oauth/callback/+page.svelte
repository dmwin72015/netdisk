<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { api, setSession, type UserInfo } from '$lib/api/client';
	import { setUser } from '$lib/stores/auth';
	import { LoaderCircle } from '@lucide/svelte';

	let status = $state('Logging in...');

	onMount(async () => {
		if (!browser) return;

		const accessToken = page.url.searchParams.get('accessToken');
		const refreshToken = page.url.searchParams.get('refreshToken');
		const error = page.url.searchParams.get('error');
		const mode = page.url.searchParams.get('mode');

	if (mode === 'bind') {
		const provider = page.url.searchParams.get('provider') ?? '';
		if (window.opener) {
			window.opener.postMessage({ bound: true, provider, error: error || null }, location.origin);
			return;
		}
		status = error || 'Account linked successfully';
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
		<LoaderCircle size={32} class="mx-auto mb-4 animate-spin text-blue-600" />
		<p class="text-sm text-gray-500">{status}</p>
	</div>
</div>
