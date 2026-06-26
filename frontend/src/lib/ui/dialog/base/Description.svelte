<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		ref = $bindable(null),
		children,
		class: className = '',
		...restProps
	}: Dialog.DescriptionProps & { children: Snippet; class?: string } = $props();
</script>

<Dialog.Description
	bind:ref
	class={cn('text-xs text-ink-3', className)}
	{...restProps}
>
	{@render children()}
</Dialog.Description>
