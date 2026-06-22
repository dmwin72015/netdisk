<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		class: className = '',
		children,
		...restProps
	}: {
		class?: string;
		children?: Snippet;
	} = $props();
</script>

<Dialog.Overlay
	class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm
		data-[state=open]:animate-in data-[state=open]:fade-in-0
		data-[state=closed]:animate-out data-[state=closed]:fade-out-0
		{className}"
	{...restProps}
>
	{#if children}
		{@render children()}
	{/if}
</Dialog.Overlay>
