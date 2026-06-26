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
	}: Dialog.TriggerProps & { children: Snippet; class?: string; disabled?: boolean; type?: 'button' | 'submit' | 'reset' } = $props();
</script>

<Dialog.Trigger
	bind:ref
	class={cn('inline-flex items-center justify-center outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2', className)}
	{disabled}
	{type}
	{...restProps}
>
	{@render children()}
</Dialog.Trigger>
