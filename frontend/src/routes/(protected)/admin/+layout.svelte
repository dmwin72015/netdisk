<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { user, authReady } from '$lib/stores/auth';
	import { Users, FileText, ChevronLeft } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let { children } = $props();

	let authorized = $state(false);

	const currentPath = $derived($page.url.pathname);

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

	const tabs = [
		{ label: 'Users', path: '/admin', icon: Users },
		{ label: 'Files', path: '/admin/files', icon: FileText },
	];
</script>

{#if $authReady && authorized}
	<div class="mx-auto max-w-6xl px-4 py-6">
		<div class="mb-6 flex items-center gap-3">
			<button
				onclick={() => goto('/')}
				class="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
			>
				<ChevronLeft size={20} />
			</button>
			<h1 class="text-2xl font-bold text-slate-800">Admin</h1>
		</div>

		<div class="mb-6 flex gap-1 rounded-lg bg-slate-100 p-1">
			{#each tabs as tab}
				<button
					onclick={() => goto(tab.path)}
					class="flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors"
					class:bg-white={currentPath === tab.path || (tab.path === '/admin' && !['/admin/files'].includes(currentPath))}
					class:shadow-sm={currentPath === tab.path || (tab.path === '/admin' && !['/admin/files'].includes(currentPath))}
				>
					<tab.icon size={16} />
					{tab.label}
				</button>
			{/each}
		</div>

		{@render children()}
	</div>
{:else if $authReady && $user && $user.role !== 'admin'}
	<div class="flex flex-col items-center justify-center py-20 text-center">
		<p class="text-lg text-slate-500">{m.admin_only()}</p>
	</div>
{/if}
