<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<div class="flex justify-end gap-3 pt-4 {className}">
	{@render children()}
</div>
