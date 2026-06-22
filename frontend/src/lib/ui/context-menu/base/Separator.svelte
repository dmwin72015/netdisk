<script lang="ts">
	import { ContextMenu } from 'bits-ui';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<ContextMenu.Separator class="bg-line-soft mx-1 my-1 h-px {className}" />
