<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { listRecentFiles, type FileItem } from '$lib/api/files';
	import { listUploadTasks, type UploadTaskItem } from '$lib/api/upload-tasks';
	import { fmtSize, fmtTime } from '$lib/utils/format';
	import { Folder, Film, Star, Trash2, LoaderCircle, File, CircleAlert, RefreshCw, Upload, CircleCheck } from '@lucide/svelte';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import * as m from '$lib/paraglide/messages';

	let recentFiles = $state<FileItem[]>([]);
	let loading = $state(true);
	let incompleteTasks = $state<UploadTaskItem[]>([]);

	const activeStatus = new Set(['created', 'uploading', 'merging', 'failed']);

	onMount(() => {
		loadRecentFiles();
		loadIncompleteTasks();
	});

	async function loadRecentFiles() {
		loading = true;
		try {
			const data = await listRecentFiles(10);
			recentFiles = data.files;
		} catch {
			recentFiles = [];
		} finally {
			loading = false;
		}
	}

	async function loadIncompleteTasks() {
		try {
			const data = await listUploadTasks(20, 0);
			incompleteTasks = data.items.filter(t => t.status !== 'done' && activeStatus.has(t.status));
		} catch {
			incompleteTasks = [];
		}
	}

	const statusConfig: Record<string, { label: string; icon: any; class: string }> = {
		created:   { label: 'Pending',     icon: Upload,      class: 'text-blue-600' },
		uploading: { label: 'Uploading',   icon: Upload,      class: 'text-blue-600' },
		merging:   { label: 'Merging',     icon: RefreshCw,   class: 'text-amber-600' },
		failed:    { label: 'Failed',      icon: CircleAlert, class: 'text-red-600' },
	};

	function taskProgress(task: UploadTaskItem): number {
		if (task.fileSize <= 0) return 0;
		return Math.min(100, Math.round((task.receivedBytes / task.fileSize) * 100));
	}

	function getFileUrl(file: FileItem): string {
		if (file.parentSlug) {
			return `/files/all/${file.parentSlug}`;
		}
		return '/files/all';
	}
</script>

{#if $authReady && $user}
	<div class="space-y-6">
		<!-- Welcome -->
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">{m.home_welcome({ name: $user.username })}</h1>
		</div>

		<!-- Quick links -->
		<div>
			<h2 class="mb-3 text-sm font-medium text-gray-500">{m.home_quick_links()}</h2>
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
				<a href="/files/all" class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-blue-200 hover:shadow-md">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
						<Folder size={20} />
					</div>
					<span class="text-sm font-medium text-gray-700">{m.home_all_files()}</span>
				</a>
				<a href="/media" class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-purple-200 hover:shadow-md">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-50 text-purple-600">
						<Film size={20} />
					</div>
					<span class="text-sm font-medium text-gray-700">{m.home_media_library()}</span>
				</a>
				<a href="/files/starred" class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-amber-200 hover:shadow-md">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
						<Star size={20} />
					</div>
					<span class="text-sm font-medium text-gray-700">{m.home_starred()}</span>
				</a>
				<a href="/files/trash" class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-300 hover:shadow-md">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100 text-gray-600">
						<Trash2 size={20} />
					</div>
					<span class="text-sm font-medium text-gray-700">{m.home_trash()}</span>
				</a>
			</div>
		</div>

		<!-- Incomplete upload tasks -->
		{#if incompleteTasks.length > 0}
			<div>
				<div class="flex items-center justify-between mb-3">
					<h2 class="text-sm font-medium text-gray-500">{m.recent_uploads()}</h2>
					<a href="/tasks" class="text-xs text-blue-600 hover:text-blue-700">{m.all_upload_tasks()} →</a>
				</div>
				<div class="space-y-2">
					{#each incompleteTasks as task (task.slug)}
						{@const cfg = statusConfig[task.status] || statusConfig.failed}
						{@const progress = taskProgress(task)}
						{@const showProgress = (task.status === 'uploading' || task.status === 'created') && progress > 0}
						<div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm" style={showProgress ? `background:linear-gradient(to right, #dbeafe ${progress}%, white ${progress}%)` : ''}>
							<div class="relative flex items-center gap-3 px-4 py-3">
								<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white/80 {cfg.class}">
									<cfg.icon size={14} />
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-gray-900">{task.fileName || 'Unknown'}</p>
									<p class="text-xs text-gray-500">
										{fmtSize(task.fileSize)} &middot; {cfg.label}
										{#if showProgress}
											<span class="text-blue-500"> &middot; {progress}%</span>
										{/if}
									</p>
								</div>
								{#if task.status === 'failed'}
									<a
										href="/tasks"
										class="shrink-0 rounded-lg bg-white/80 px-3 py-1.5 text-xs font-medium text-red-600 transition-colors hover:bg-white"
									>{m.upload_retry()}</a>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Recent files -->
		<div>
			<div class="flex items-center justify-between mb-3">
				<h2 class="text-sm font-medium text-gray-500">{m.home_recent_files()}</h2>
				<a href="/files/all" class="text-xs text-blue-600 hover:text-blue-700">{m.all_files()} →</a>
			</div>

			{#if loading}
				<div class="flex items-center justify-center py-12">
					<LoaderCircle size={24} class="animate-spin text-gray-300" />
				</div>
			{:else if recentFiles.length === 0}
				<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-12 text-center">
					<File size={40} class="mb-3 text-gray-300" />
					<p class="text-sm text-gray-400">{m.home_no_files()}</p>
				</div>
			{:else}
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
					{#each recentFiles as file (file.slug)}
						<a
							href={getFileUrl(file)}
							class="group flex flex-col items-center rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-200 hover:shadow-md"
						>
								<MimeIcon mimeType={file.mimeType} name={file.fileName} isDir={file.isDir} category={file.fileCategory} size={36} />
							<p class="mt-3 w-full truncate text-center text-sm font-medium text-gray-700" title={file.fileName}>{file.fileName}</p>
							{#if file.parentName}
								<p class="mt-0.5 w-full truncate text-center text-xs text-gray-400" title={file.parentName}>{file.parentName}</p>
							{/if}
							<div class="mt-0.5 flex items-center gap-1.5 text-xs text-gray-400">
								<span>{fmtSize(file.fileSize)}</span>
								<span>·</span>
								<span>{fmtTime(file.createdAt)}</span>
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/if}
