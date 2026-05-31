<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { user, authReady } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';

	let { children } = $props();

	let authorized = $state(false);

	onMount(() => {
		if (!$user) {
			void goto('/login');
			return;
		}
		if ($user.role !== 'admin') {
			void goto('/');
			return;
		}
		authorized = true;
	});
</script>

{#if $authReady && authorized}
	{@render children()}
{:else if $authReady && $user && $user.role !== 'admin'}
	<div class="flex flex-col items-center justify-center py-20 text-center">
		<p class="text-lg text-slate-500">{m.admin_only()}</p>
	</div>
{/if}
