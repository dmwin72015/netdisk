<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		ref = $bindable(null),
		children,
		class: className = '',
		level = 2 as 1 | 2 | 3 | 4 | 5 | 6,
		...restProps
	}: Dialog.TitleProps & { children: Snippet; class?: string } = $props();
</script>

<Dialog.Title
	bind:ref
	class={cn('text-sm font-medium leading-5 text-ink-2', className)}
	{level}
	{...restProps}
>
	{@render children()}
</Dialog.Title>
