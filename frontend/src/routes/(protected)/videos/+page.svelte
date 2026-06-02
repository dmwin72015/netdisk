<script lang="ts">
	import { onMount } from 'svelte';
	import { listVideos } from '$lib/api/videos';
	import type { Task } from '$lib/api/tasks';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import * as m from '$lib/paraglide/messages';

	let items = $state<Task[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function refresh() {
		loading = true;
		try {
			const data = await listVideos(20, 0);
			items = data.items;
		} catch (err) {
			error = err instanceof Error ? err.message : m.load_failed();
		} finally {
			loading = false;
		}
	}

	onMount(refresh);
</script>

<h1 class="text-xl font-semibold">{m.video_library()}</h1>
{#if loading}
	<p class="mt-3 text-sm text-slate-500">{m.loading()}</p>
{:else if error}
	<p class="mt-3 text-sm text-red-600">{error}</p>
{:else if items.length === 0}
	<p class="mt-3 text-sm text-slate-500">{m.no_videos()}</p>
{:else}
	<div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
		{#each items as video (video.id)}
			<VideoCard task={video} onChanged={refresh} />
		{/each}
	</div>
{/if}
