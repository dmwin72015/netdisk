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
	{level}
	class="text-base font-semibold leading-none tracking-tight {className}"
>
	{@render children()}
</Dialog.Title>
