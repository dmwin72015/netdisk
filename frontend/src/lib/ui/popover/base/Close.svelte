<script lang="ts">
	import { Popover } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		...restProps
	}: {
		children?: Snippet;
		class?: string;
	} = $props();
</script>

<Popover.Close
	class="inline-flex items-center justify-center rounded-md p-1 text-gray-400 outline-none transition-colors hover:text-gray-600 focus-visible:ring-2 focus-visible:ring-blue-500 {className}"
	{...restProps}
>
	{#if children}
		{@render children()}
	{:else}
		<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
			<path d="M18 6 6 18" /><path d="m6 6 12 12" />
		</svg>
	{/if}
</Popover.Close>
