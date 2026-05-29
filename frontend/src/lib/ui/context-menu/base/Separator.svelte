<script lang="ts">
	import { ContextMenu } from 'bits-ui';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<ContextMenu.Separator class="mx-2 my-1 h-px bg-gray-100 {className}" />
