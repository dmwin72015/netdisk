<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { listRecentFiles, type FileItem } from '$lib/api/files';
	import { listUploadTasks, type UploadTaskItem } from '$lib/api/upload-tasks';
	import { fmtSize, fmtTime } from '$lib/utils/format';
	import { Folder, Film, Star, Trash2, LoaderCircle, CircleAlert, RefreshCw, Upload, Clock } from '@lucide/svelte';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import * as m from '$lib/paraglide/messages';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';

	let recentFiles = $state<FileItem[]>([]);
	let loading = $state(true);
	let incompleteTasks = $state<UploadTaskItem[]>([]);

	const activeStatus = new Set(['created', 'uploading', 'merging', 'failed', 'queued', 'downloading']);

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

	// Status uses the semantic palette only (success / warning / danger / info / ink-3 for pending).
	const statusConfig: Record<string, { label: string; icon: any; class: string }> = {
		created:     { label: 'Pending',     icon: Upload,      class: 'text-primary' },
		uploading:   { label: 'Uploading',   icon: Upload,      class: 'text-primary' },
		merging:     { label: 'Merging',     icon: RefreshCw,   class: 'text-warning' },
		failed:      { label: 'Failed',      icon: CircleAlert, class: 'text-danger' },
		queued:      { label: 'Queued',      icon: Clock,       class: 'text-ink-4' },
		downloading: { label: 'Downloading', icon: Upload,      class: 'text-primary' },
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

	// Quick links — all use the cool-gray neutral surface. Icons carry the role tint; the card itself stays quiet.
	const quickLinks = [
		{ href: '/files/all',     label: m.home_all_files,    icon: Folder,  iconClass: 'text-primary' },
		{ href: '/media',         label: m.home_media_library, icon: Film,   iconClass: 'text-ink-2' },
		{ href: '/files/starred', label: m.home_starred,      icon: Star,    iconClass: 'text-star' },
		{ href: '/files/trash',   label: m.home_trash,        icon: Trash2,  iconClass: 'text-ink-3' },
	];
</script>

{#if $authReady && $user}
	<div class="relative px-6 pt-4 pb-6">
		<div
			class="pointer-events-none absolute inset-0 -z-10"
			style="background: radial-gradient(ellipse 55% 40% at 50% 8%, rgba(37,99,235,0.1) 0%, transparent 55%), radial-gradient(ellipse 75% 50% at 50% 30%, rgba(37,99,235,0.04) 0%, transparent 55%);"
		></div>
		<div class="relative space-y-8">
		<!-- Welcome -->
		<header>
			<h1 class="text-ink text-[22px] font-semibold tracking-tight">
				{m.home_welcome({ name: $user.profile.displayName || $user.username })}
			</h1>
		</header>

		<!-- Quick links — flat at rest; hover only changes background -->
		<section>
			<h2 class="text-ink-3 mb-3 text-xs font-medium tracking-wide">{m.home_quick_links()}</h2>
			<div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
				{#each quickLinks as link (link.href)}
					{@const Icon = link.icon}
					<a
						href={link.href}
						class="border-line bg-surface hover:bg-surface-muted group flex items-center gap-3 rounded-lg border px-3.5 py-3 transition-colors duration-150"
					>
						<span class="bg-surface-sunken group-hover:bg-surface flex h-9 w-9 items-center justify-center rounded-md transition-colors duration-150 {link.iconClass}">
							<Icon size={18} strokeWidth={1.75} />
						</span>
						<span class="text-ink-2 group-hover:text-ink text-sm font-medium transition-colors duration-150">
							{link.label()}
						</span>
					</a>
				{/each}
			</div>
		</section>

		<!-- Incomplete upload tasks -->
		{#if incompleteTasks.length > 0}
			<section>
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-ink-3 text-xs font-medium tracking-wide">{m.recent_uploads()}</h2>
					<a href="/tasks" class="text-primary hover:text-primary-hover text-xs transition-colors duration-150">
						{m.all_upload_tasks()} →
					</a>
				</div>
				<div class="space-y-2">
					{#each incompleteTasks as task (task.slug)}
						{@const cfg = statusConfig[task.status] || statusConfig.failed}
						{@const progress = taskProgress(task)}
						{@const showProgress = (task.status === 'uploading' || task.status === 'created' || task.status === 'downloading') && progress > 0}
						<div class="border-line bg-surface relative overflow-hidden rounded-lg border">
							<!-- progress fill (driven by transform; no width animation) -->
							{#if showProgress}
								<div
									aria-hidden="true"
									class="bg-primary-soft pointer-events-none absolute inset-y-0 left-0 origin-left transition-transform duration-200 ease-out"
									style="width:100%; transform:scaleX({progress / 100});"
								></div>
							{/if}
							<div class="relative flex items-center gap-3 px-4 py-3">
								<div class="bg-surface-sunken {cfg.class} flex h-8 w-8 shrink-0 items-center justify-center rounded-md">
									<cfg.icon size={14} strokeWidth={2} />
								</div>
								<div class="min-w-0 flex-1">
									<p class="text-ink truncate text-sm font-medium">{task.fileName || 'Unknown'}</p>
									<p class="text-ink-3 mt-0.5 text-xs tabular-nums">
										{task.status === 'downloading' || task.status === 'queued' ? fmtSize(task.receivedBytes) : fmtSize(task.fileSize)} · {cfg.label}
										{#if task.status === 'downloading' || task.status === 'queued'}
											<span class="text-primary"> · {fmtSize(task.receivedBytes)}</span>
										{:else if showProgress}
											<span class="text-primary"> · {progress}%</span>
										{/if}
									</p>
								</div>
								{#if task.status === 'failed'}
									<a
										href="/tasks"
										class="text-danger hover:bg-danger-soft shrink-0 rounded-md px-3 py-1.5 text-xs font-medium transition-colors duration-150"
									>{m.upload_retry()}</a>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</section>
		{/if}

		<!-- Recent files -->
		<section>
			<div class="mb-3 flex items-center justify-between">
				<h2 class="text-ink-3 text-xs font-medium tracking-wide">{m.home_recent_files()}</h2>
				<a href="/files/all" class="text-primary hover:text-primary-hover text-xs transition-colors duration-150">
					{m.all_files()} →
				</a>
			</div>

			{#if loading}
				<div class="flex items-center justify-center py-12">
					<LoaderCircle size={20} class="text-ink-4 animate-spin" />
				</div>
			{:else if recentFiles.length === 0}
				<div class="border-line text-ink-3 flex flex-col items-center justify-center rounded-lg border border-dashed py-14 text-center">
					<img src={noFilesSvg} class="mb-2 w-28 h-28" alt="" />
					<p class="text-sm">{m.home_no_files()}</p>
				</div>
			{:else}
				<div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
					{#each recentFiles as file (file.slug)}
						<a
							href={getFileUrl(file)}
							class="border-line bg-surface hover:bg-surface-muted group flex flex-col items-center rounded-lg border p-3.5 transition-colors duration-150"
							title={file.fileName}
						>
							<MimeIcon mimeType={file.mimeType} name={file.fileName} isDir={file.isDir} category={file.fileCategory} size={32} />
							<p class="text-ink mt-3 w-full truncate text-center text-sm font-medium">{file.fileName}</p>
							{#if file.parentName}
								<p class="text-ink-4 mt-0.5 w-full truncate text-center text-xs" title={file.parentName}>{file.parentName}</p>
							{/if}
							<div class="text-ink-4 mt-1 flex items-center gap-1.5 text-xs tabular-nums">
								<span>{fmtSize(file.fileSize)}</span>
								<span aria-hidden="true">·</span>
								<span>{fmtTime(file.createdAt)}</span>
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</section>
	</div>
	</div>
{/if}
