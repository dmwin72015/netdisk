<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import * as m from '$lib/paraglide/messages';
	import AppDialog from '$lib/components/AppDialog.svelte';
	import PinInputDialog from '$lib/components/PinInputDialog.svelte';
	import { Toast } from '$lib/ui/toast';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { getStoredUser } from '$lib/api/client';
	import { user, authReady } from '$lib/stores/auth';
	import { themeManager } from '$lib/services/themeManager.svelte';

	let { children, data } = $props();
	let isAuthPage = $derived(data.isAuthPage);

	if (browser) {
		onMount(() => {
			themeManager.init();
			user.set(getStoredUser());
			authReady.set(true);
		});
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>{m.app_name()}</title>
</svelte:head>

<div class="min-h-screen">
	<main class="min-h-screen w-full">
		{@render children()}
	</main>
</div>

<AppDialog />
<PinInputDialog />
<Toast theme={themeManager.resolvedTheme} />
