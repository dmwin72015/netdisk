<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		level = 2,
	}: {
		children: Snippet;
		class?: string;
		level?: 1 | 2 | 3 | 4 | 5 | 6;
	} = $props();
</script>

<Dialog.Title
	class="text-sm font-medium leading-5 text-ink-2 {className}"
	{level}
>
	{@render children()}
</Dialog.Title>
