<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
		disabled = false,
		...restProps
	}: {
		children?: Snippet;
		class?: string;
		disabled?: boolean;
		[key: string]: unknown;
	} = $props();
</script>

<Dialog.Close
	class={cn('inline-flex items-center justify-center rounded-md p-1 text-ink-3 outline-none transition-colors hover:bg-surface-sunken hover:text-ink', className)}
	{disabled}
	{...restProps}
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
