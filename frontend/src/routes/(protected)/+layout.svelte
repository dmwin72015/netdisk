<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { fetchConfig } from '$lib/stores/config';

	let { children } = $props();

	onMount(() => {
		if (!browser) return;
		if (!$user) {
			void goto('/login');
			return;
		}
		void fetchConfig();
	});
</script>

{#if $authReady && $user}
	{@render children()}
{/if}
