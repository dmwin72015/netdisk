<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<Dialog.Description class="text-xs text-ink-3 {className}">
	{@render children()}
</Dialog.Description>
