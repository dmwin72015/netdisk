<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';

	let { children } = $props();

	onMount(() => {
		if (!browser) return;
		if ($user) {
			void goto('/');
		}
	});
</script>

{#if $authReady && !$user}
	{@render children()}
{/if}
