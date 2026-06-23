<script lang="ts">
	import { AlertDialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
		...restProps
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<AlertDialog.Description
	class={cn('mt-1 text-xs text-ink-3', className)}
	{...restProps}
>
	{@render children()}
</AlertDialog.Description>
