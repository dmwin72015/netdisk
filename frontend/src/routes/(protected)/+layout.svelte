<script lang="ts">
	import { onMount } from 'svelte';
	import { setContext } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { fetchConfig } from '$lib/stores/config';
	import { loadPreferencesFromServer } from '$lib/stores/file-preferences.svelte';
	import { rememberFilesUrl } from '$lib/stores/last-section-url';
	import AppShell from '$lib/components/AppShell.svelte';
	import UploadPanel from '$lib/components/files/UploadPanel.svelte';
	import { createUploadManager } from '$lib/upload-manager.svelte';

	let { children } = $props();

	const upload = createUploadManager({
		storageKey: 'nd.global.uploads',
	});
	setContext('upload', upload);

	let retryInput: HTMLInputElement | undefined = $state();
	let retryUid = $state<string | null>(null);

	function handleRetry(uid: string) {
		retryUid = uid;
		retryInput?.click();
	}

	function onRetryPick(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const file = el?.files?.[0];
		el.value = '';
		if (!file || !retryUid) return;
		upload.retryItem(retryUid, file);
		retryUid = null;
	}

	onMount(() => {
		if (!browser) return;
		if (!$user) {
			void goto('/login');
			return;
		}
		void fetchConfig();
		void loadPreferencesFromServer();
		upload.restore();
	});

	// 仅缓存「文件」tab 下的最后一次访问路径，从照片 / 媒体库切回时恢复。
	afterNavigate(({ to }) => {
		if (!to) return;
		rememberFilesUrl(to.url.pathname, to.url.search);
	});
</script>

{#if $authReady && $user}
	<AppShell>
		{@render children()}
	</AppShell>

	<input bind:this={retryInput} type="file" class="hidden" onchange={onRetryPick} />

	<UploadPanel
		items={upload.items}
		onPause={upload.pauseUpload}
		onResume={upload.resumeUpload}
		onDelete={upload.deleteUpload}
		onDismiss={upload.dismissAll}
		onRetry={handleRetry}
	/>
{/if}
