<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		ref = $bindable(null),
		children,
		class: className = '',
		disabled = false,
		type = 'button' as const,
		...restProps
	}: Dialog.CloseProps & { children?: Snippet; class?: string; disabled?: boolean; type?: 'button' | 'submit' | 'reset' } = $props();
</script>

<Dialog.Close
	bind:ref
	class={cn('inline-flex items-center justify-center rounded-md p-1 text-ink-3 outline-none transition-colors hover:bg-surface-sunken hover:text-ink', className)}
	{disabled}
	{type}
	{...restProps}
>
	{#if children}
		{@render children()}
	{/if}
</Dialog.Close>
