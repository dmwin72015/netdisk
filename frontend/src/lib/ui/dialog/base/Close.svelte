<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		disabled = false,
		...restProps
	}: {
		children: Snippet;
		class?: string;
		disabled?: boolean;
		[key: string]: unknown;
	} = $props();
</script>

<Dialog.Close
	class="inline-flex items-center justify-center rounded-md p-1 text-ink-3 outline-none transition-colors hover:bg-surface-sunken hover:text-ink {className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</Dialog.Close>
