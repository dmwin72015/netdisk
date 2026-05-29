<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
	}: {
		children?: Snippet;
		class?: string;
	} = $props();
</script>

<Dialog.Close
	class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500
		hover:bg-gray-100 hover:text-gray-900 focus-visible:outline-none
		focus-visible:ring-2 focus-visible:ring-blue-500 {className}"
>
	{#if children}
		{@render children()}
	{:else}
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d="M18 6 6 18" />
			<path d="m6 6 12 12" />
		</svg>
	{/if}
</Dialog.Close>
