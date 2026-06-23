<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<div class={cn('flex flex-col space-y-1.5 p-4 pb-2', className)}>
	{@render children()}
</div>
