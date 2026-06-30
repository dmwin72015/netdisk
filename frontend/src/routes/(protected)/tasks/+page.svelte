<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, authReady } from '$lib/stores/auth';
	import { listUploadTasks, retryUploadTask, deleteUploadTask, deleteUploadTasks, type UploadTaskItem } from '$lib/api/upload-tasks';
	import { ListRestart, CircleCheck, CircleX, Clock, Upload, Download, LoaderCircle, ArrowLeft, ArrowRight, Trash2, Check, ChevronDown, Copy } from '@lucide/svelte';
	import { confirmDelete } from '$lib/dialog';
	import { DatePicker } from '$lib/ui/date-picker';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { toast } from 'svelte-sonner';
	import Popover from '$lib/ui/popover/Popover.svelte';
	import Tooltip from '$lib/ui/tooltip/Tooltip.svelte';
	import * as m from '$lib/paraglide/messages';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';
	import { fmtSize, fmtTime, copyToClipboard } from '$lib/utils/format';

	let tasks = $state<UploadTaskItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let limit = $state(20);
	let offset = $state(0);
	let retrying = $state<Record<string, boolean>>({});
	let startDate = $state<Date | null>(null);
	let endDate = $state<Date | null>(null);
	let selected = $state<Set<string>>(new Set());
	let statusFilter = $state('');
	let statusFilterOpen = $state(false);

	let totalPages = $derived(Math.ceil(total / limit));
	let currentPage = $derived(Math.floor(offset / limit) + 1);
	let allSelected = $derived(tasks.length > 0 && tasks.every(t => selected.has(t.slug)));

	async function refresh() {
		if (!$user) return;
		loading = true;
		try {
			const sd = startDate ? startDate.toISOString() : undefined;
			const ed = endDate ? new Date(endDate.getFullYear(), endDate.getMonth(), endDate.getDate(), 23, 59, 59).toISOString() : undefined;
			const data = await listUploadTasks(limit, offset, sd, ed, statusFilter || undefined);
			tasks = data.items;
			total = data.total;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loading = false;
		}
	}

	async function handleRetry(task: UploadTaskItem) {
		retrying[task.slug] = true;
		try {
			const result = await retryUploadTask(task.slug);
			toast.success(m.upload_retry() + ' ' + m.upload_started());
			// Navigate to upload page with the new task slug
			await goto(`/files/all?resume=${result.uploadSlug}`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.upload_failed());
		} finally {
			retrying[task.slug] = false;
		}
	}

	function statusIcon(status: string) {
		switch (status) {
			case 'done': return CircleCheck;
			case 'failed': return CircleX;
			case 'uploading':
			case 'merging': return Upload;
			case 'downloading': return Download;
			default: return Clock;
		}
	}

	function statusClass(status: string) {
		switch (status) {
			case 'done': return 'text-success';
			case 'failed': return 'text-danger';
			case 'uploading':
			case 'merging':
			case 'downloading': return 'text-primary';
			default: return 'text-ink-4';
		}
	}

	function statusLabel(status: string) {
		switch (status) {
			case 'done': return m.upload_done();
			case 'failed': return m.failed();
			case 'uploading': return m.uploading_status();
			case 'merging': return m.converting_status();
			case 'created': return m.waiting();
			case 'queued': return m.queued_status();
			case 'downloading': return m.downloading_status();
			default: return status;
		}
	}

	function statusFilterLabel(status: string) {
		return status ? statusLabel(status) : m.all_status();
	}

	function statusFilterOptions() {
		return [
			{ value: '', label: m.all_status() },
			{ value: 'done', label: m.upload_done() },
			{ value: 'failed', label: m.failed() },
			{ value: 'uploading', label: m.uploading_status() },
			{ value: 'downloading', label: m.downloading_status() },
			{ value: 'queued', label: m.queued_status() },
			{ value: 'created', label: m.waiting() },
			{ value: 'merging', label: m.converting_status() },
		];
	}

	function selectStatusFilter(status: string) {
		statusFilter = status;
		statusFilterOpen = false;
	}

	function taskProgress(task: UploadTaskItem): number {
		if (task.fileSize <= 0) return 0;
		return Math.min(100, Math.round((task.receivedBytes / task.fileSize) * 100));
	}

	function goToPage(page: number) {
		offset = (page - 1) * limit;
		refresh();
	}

	function applyDateFilter() {
		offset = 0;
		refresh();
	}

	function clearDateFilter() {
		startDate = null;
		endDate = null;
		statusFilter = '';
		offset = 0;
		refresh();
	}

	function toggleSelect(slug: string) {
		if (selected.has(slug)) selected.delete(slug);
		else selected.add(slug);
		selected = new Set(selected); // trigger reactivity
	}

	function toggleSelectAll() {
		if (allSelected) {
			selected = new Set();
		} else {
			selected = new Set(tasks.map(t => t.slug));
		}
	}

	async function handleDeleteSelected() {
		if (selected.size === 0) return;
		const count = selected.size;
		if (!(await confirmDelete(m.confirm_delete_tasks({ count: String(count) })))) return;
		try {
			await deleteUploadTasks([...selected]);
			selected = new Set();
			toast.success(m.tasks_deleted());
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		}
	}

	async function handleDeleteSingle(slug: string) {
		if (!(await confirmDelete(m.confirm_delete_task()))) return;
		try {
			await deleteUploadTask(slug);
			selected.delete(slug);
			selected = new Set(selected);
			toast.success(m.task_deleted());
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		}
	}

	onMount(() => {
		refresh();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4 px-6 pt-4 pb-6">
		<!-- Header -->
		<div class="flex items-center gap-2">
			<ListRestart size={20} class="text-ink-3" />
			<h1 class="text-lg font-semibold text-ink">{m.upload_title()}</h1>
			<span class="text-sm text-ink-4">{m.total_items({ total: String(total) })}</span>
		</div>

		<!-- Filters -->
		<div class="flex items-center gap-2">
			<DatePicker bind:value={startDate} placeholderText={m.start_date()} />
			<span class="text-xs text-ink-4">—</span>
			<DatePicker bind:value={endDate} placeholderText={m.end_date()} />
			<Dropdown
				bind:open={statusFilterOpen}
				triggerClass="flex h-8 w-[6.75rem] items-center justify-between gap-2 rounded-lg border border-line bg-surface px-2.5 text-sm text-ink-2 transition-colors hover:border-line hover:bg-surface-sunken data-[state=open]:border-primary data-[state=open]:bg-surface"
				contentClass="min-w-36"
				sideOffset={6}
				align="end"
			>
				{#snippet trigger()}
					<span class="truncate">{statusFilterLabel(statusFilter)}</span>
					<ChevronDown size={14} class="shrink-0 text-ink-4" />
				{/snippet}
				{#each statusFilterOptions() as option (option.value)}
					<DropdownBase.Item onSelect={() => selectStatusFilter(option.value)}>
						{#snippet children()}
							<span class="flex flex-1 items-center justify-between gap-3">
								<span>{option.label}</span>
								{#if statusFilter === option.value}
									<Check size={14} class="text-primary" />
								{/if}
							</span>
						{/snippet}
					</DropdownBase.Item>
				{/each}
			</Dropdown>
			<button type="button" onclick={applyDateFilter}
				class="flex h-8 items-center rounded-lg bg-primary px-3 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover">
				{m.filter()}
			</button>
			{#if startDate || endDate || statusFilter}
				<button type="button" onclick={clearDateFilter}
					class="flex h-8 items-center rounded-lg border border-line px-3 text-sm text-ink-3 transition-colors hover:bg-surface-sunken">
					{m.reset()}
				</button>
			{/if}
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if tasks.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center">
				<img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
				<p class="text-sm text-ink-4">{m.no_tasks()}</p>
			</div>
		{:else}
			{#if selected.size > 0}
				<div class="flex items-center gap-3 rounded-lg border border-line-soft bg-surface-muted px-4 py-2">
					<span class="text-sm text-ink-3">{m.selected_count({ count: String(selected.size) })}</span>
					<button type="button" onclick={handleDeleteSelected} class="inline-flex items-center gap-1.5 rounded-lg bg-danger px-3 py-1.5 text-xs font-medium text-primary-on transition-colors hover:bg-danger-hover">
						<Trash2 size={12} />
						{m.delete_btn()}
					</button>
				</div>
			{/if}

			<div class="overflow-hidden rounded-xl border border-line-soft bg-surface ">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-line-soft text-left text-xs font-medium text-ink-4">
							<th class="w-10 px-4 py-2.5">
								<button type="button" onclick={toggleSelectAll} class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {allSelected ? 'border-primary bg-primary text-primary-on' : 'border-line'}">
									{#if allSelected}
										<Check size={12} />
									{/if}
								</button>
							</th>
							<th class="px-4 py-2.5 font-medium">{m.col_filename()}</th>
							<th class="w-[100px] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
							<th class="w-[120px] px-4 py-2.5 font-medium">{m.status()}</th>
							<th class="w-32 px-4 py-2.5 font-medium">{m.col_upload_type()}</th>
							<th class="w-[160px] px-4 py-2.5 font-medium">{m.col_directory()}</th>
							<th class="w-[140px] px-4 py-2.5 text-right font-medium">{m.col_upload_time()}</th>
							<th class="w-[120px] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody>
						{#each tasks as task (task.slug)}
							{@const StatusIcon = statusIcon(task.status)}
							{@const progress = taskProgress(task)}
							<tr class="border-b border-line-soft transition-colors last:border-0 hover:bg-surface-sunken/80">
								<td class="px-4 py-2.5">
									<button type="button" onclick={() => toggleSelect(task.slug)} class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {selected.has(task.slug) ? 'border-primary bg-primary text-primary-on' : 'border-line'}">
										{#if selected.has(task.slug)}
											<Check size={12} />
										{/if}
									</button>
								</td>
								<td class="max-w-[200px] truncate px-4 py-2.5 text-ink" title={task.fileName || task.slug}>
									{task.fileName || task.slug}
								</td>
								<td class="px-4 py-2.5 text-right tabular-nums text-ink-3">{fmtSize(task.fileSize)}</td>
								<td class="px-4 py-2.5">
									{#if task.errorMsg}
										<Tooltip content={task.errorMsg} side="bottom" sideOffset={2} delayDuration={0}>
											{#snippet children()}
												<span class="inline-flex items-center gap-1.5 cursor-help">
													<StatusIcon size={14} class={statusClass(task.status)} />
													<span class={statusClass(task.status)}>{statusLabel(task.status)}</span>
													{#if task.status === 'downloading' || task.status === 'queued'}
														<span class="text-xs text-primary">{fmtSize(task.receivedBytes)}</span>
													{:else if progress > 0 && (task.status === 'uploading' || task.status === 'created')}
														<span class="text-xs text-primary">{progress}%</span>
													{/if}
												</span>
											{/snippet}
										</Tooltip>
									{:else}
										<span class="inline-flex items-center gap-1.5">
											<StatusIcon size={14} class={statusClass(task.status)} />
											<span class={statusClass(task.status)}>{statusLabel(task.status)}</span>
											{#if task.status === 'downloading' || task.status === 'queued'}
												<span class="text-xs text-primary">{fmtSize(task.receivedBytes)}</span>
											{:else if progress > 0 && (task.status === 'uploading' || task.status === 'created')}
												<span class="text-xs text-primary">{progress}%</span>
											{/if}
										</span>
									{/if}
								</td>
								<td class="w-32 px-4 py-2.5">
									{#if task.taskType === 'url'}
										<Tooltip content={task.sourceUrl ?? ''} side="bottom" sideOffset={2} delayDuration={0}>
											{#snippet children()}
												<span class="flex items-center gap-1.5 text-primary cursor-help">
													<Download class="size-3.5" />
													{m.remote_upload()}
													<button
														class="text-ink-4 hover:text-ink-3"
														onclick={async (e: MouseEvent) => {
															e.stopPropagation();
															const ok = await copyToClipboard(task.sourceUrl!);
															if (ok) toast.success(m.copied()); else toast.error(m.copy_failed());
														}}
													>
														<Copy class="size-3" />
													</button>
												</span>
											{/snippet}
										</Tooltip>
									{:else}
										<span class="text-ink-3"><Upload class="inline size-3.5 mr-1" />{m.normal_upload()}</span>
									{/if}
								</td>
								<td class="px-4 py-2.5 tabular-nums text-ink-4">
									{#if task.parentSlug}
										<a href="/files/all/{task.parentSlug}" class="text-primary hover:underline" title={task.parentName ? `${task.parentName} (${task.parentSlug})` : task.parentSlug}>{task.parentName || task.parentSlug}</a>
									{:else}
										<span class="text-ink-4">/</span>
									{/if}
								</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right tabular-nums text-ink-4">{fmtTime(task.createdAt)}</td>
								<td class="px-4 py-2.5 text-right">
									<div class="flex items-center justify-end gap-1">
										{#if task.status === 'failed'}
											<button
												type="button"
												onclick={() => handleRetry(task)}
												disabled={retrying[task.slug]}
												class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:opacity-50"
											>
												{#if retrying[task.slug]}
													<LoaderCircle size={12} class="animate-spin" />
												{/if}
												{m.upload_retry()}
											</button>
										{/if}
										<button type="button" onclick={() => handleDeleteSingle(task.slug)} class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-danger" title={m.delete_btn()}>
											<Trash2 size={14} />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="flex items-center justify-center gap-1 text-sm">
					<button
						type="button"
						onclick={() => goToPage(currentPage - 1)}
						disabled={currentPage <= 1}
						class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-ink-3 transition-colors hover:bg-surface-sunken disabled:opacity-30"
					>
						<ArrowLeft size={14} />
					</button>
					{#each Array(totalPages) as _, i}
						<button
							type="button"
							onclick={() => goToPage(i + 1)}
							class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-colors {i + 1 === currentPage ? 'bg-primary font-medium text-primary-on' : 'text-ink-3 hover:bg-surface-sunken'}"
						>
							{i + 1}
						</button>
					{/each}
					<button
						type="button"
						onclick={() => goToPage(currentPage + 1)}
						disabled={currentPage >= totalPages}
						class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-ink-3 transition-colors hover:bg-surface-sunken disabled:opacity-30"
					>
						<ArrowRight size={14} />
					</button>
				</div>
			{/if}
		{/if}
	</div>
{/if}
