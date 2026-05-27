<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Navbar from '$lib/components/Navbar.svelte';
	import * as m from '$lib/paraglide/messages';
	import AppDialog from '$lib/components/AppDialog.svelte';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { getStoredUser } from '$lib/api/client';
	import { user, authReady } from '$lib/stores/auth';

	let { children } = $props();

	if (browser) {
		onMount(() => {
			user.set(getStoredUser());
			authReady.set(true);
		});
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>{m.app_name()}</title>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />
	<main class="mx-auto w-full max-w-6xl flex-1 px-4 py-5">
		{@render children()}
	</main>
</div>

<AppDialog />
