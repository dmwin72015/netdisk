<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user } from '$lib/stores/auth';

	onMount(() => {
		if (!browser) return;
		void goto('/files');
	});
</script>
