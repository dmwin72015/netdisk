<script lang="ts">
	import { DropdownMenu } from 'bits-ui';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<DropdownMenu.Separator class="bg-line-soft mx-1 my-1 h-px {className}" />
