<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		children,
		...restProps
	}: Dialog.RootProps & { children: Snippet } = $props();
</script>

<Dialog.Root bind:open {...restProps}>
	{@render children()}
</Dialog.Root>
