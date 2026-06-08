<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, authReady } from '$lib/stores/auth';
	import { listUploadTasks, retryUploadTask, deleteUploadTask, deleteUploadTasks, type UploadTaskItem } from '$lib/api/upload-tasks';
	import { ListRestart, CircleCheck, CircleX, Clock, Upload, LoaderCircle, ArrowLeft, ArrowRight, Trash2, Check, ChevronDown } from '@lucide/svelte';
	import { confirmDelete } from '$lib/dialog';
	import { DatePicker } from '$lib/ui/date-picker';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize, fmtTime } from '$lib/utils/format';

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
			default: return Clock;
		}
	}

	function statusClass(status: string) {
		switch (status) {
			case 'done': return 'text-green-600';
			case 'failed': return 'text-red-600';
			case 'uploading':
			case 'merging': return 'text-blue-600';
			default: return 'text-gray-400';
		}
	}

	function statusLabel(status: string) {
		switch (status) {
			case 'done': return m.upload_done();
			case 'failed': return m.failed();
			case 'uploading': return m.uploading_status();
			case 'merging': return m.converting_status();
			case 'created': return m.waiting();
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
	<div class="space-y-4">
		<div class="flex items-center justify-between">
			<h1 class="flex items-center gap-2 text-xl font-semibold">
				<ListRestart size={20} /> {m.upload_title()}
			</h1>
			<div class="flex items-center gap-2">
				<DatePicker bind:value={startDate} placeholderText={m.start_date()} />
				<span class="text-xs text-gray-400">—</span>
				<DatePicker bind:value={endDate} placeholderText={m.end_date()} />
					<Dropdown
						bind:open={statusFilterOpen}
						triggerClass="flex h-8 w-[6.75rem] items-center justify-between gap-2 rounded-lg border border-gray-200 bg-white px-2.5 text-sm text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50 data-[state=open]:border-blue-400 data-[state=open]:bg-white"
						contentClass="min-w-36"
						sideOffset={6}
					align="end"
				>
					{#snippet trigger()}
						<span class="truncate">{statusFilterLabel(statusFilter)}</span>
						<ChevronDown size={14} class="shrink-0 text-gray-400" />
					{/snippet}
					{#each statusFilterOptions() as option (option.value)}
						<DropdownBase.Item onSelect={() => selectStatusFilter(option.value)}>
							{#snippet children()}
								<span class="flex flex-1 items-center justify-between gap-3">
									<span>{option.label}</span>
									{#if statusFilter === option.value}
										<Check size={14} class="text-blue-500" />
									{/if}
								</span>
							{/snippet}
						</DropdownBase.Item>
					{/each}
				</Dropdown>
				<button type="button" onclick={applyDateFilter}
					class="h-8 rounded-lg bg-blue-600 px-3 text-sm font-medium text-white transition-colors hover:bg-blue-700">
					{m.filter()}
				</button>
				{#if startDate || endDate || statusFilter}
					<button type="button" onclick={clearDateFilter}
						class="h-8 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 transition-colors hover:bg-gray-50">
						{m.reset()}
					</button>
				{/if}
			</div>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if tasks.length === 0}
			<div class="flex flex-col items-center justify-center py-16 text-center">
				<ListRestart size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.no_tasks()}</p>
			</div>
		{:else}
			{#if selected.size > 0}
				<div class="flex items-center gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-2">
					<span class="text-sm text-gray-600">{m.selected_count({ count: String(selected.size) })}</span>
					<button type="button" onclick={handleDeleteSelected} class="inline-flex items-center gap-1.5 rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-red-700">
						<Trash2 size={12} />
						{m.delete_btn()}
					</button>
				</div>
			{/if}

			<div class="overflow-hidden rounded-xl border border-gray-100">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-50 bg-gray-50/50 text-left text-xs font-medium text-gray-500">
							<th class="w-10 px-4 py-3">
								<button type="button" onclick={toggleSelectAll} class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {allSelected ? 'border-blue-500 bg-blue-500 text-white' : 'border-gray-300'}">
									{#if allSelected}
										<Check size={12} />
									{/if}
								</button>
							</th>
							<th class="px-4 py-3">{m.col_filename()}</th>
							<th class="px-4 py-3">{m.col_size()}</th>
							<th class="px-4 py-3">{m.status()}</th>
							<th class="px-4 py-3">{m.col_directory()}</th>
							<th class="px-4 py-3">{m.col_upload_time()}</th>
							<th class="px-4 py-3">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each tasks as task (task.slug)}
							{@const StatusIcon = statusIcon(task.status)}
							{@const progress = taskProgress(task)}
							{@const isInterrupted = (task.status === 'uploading' || task.status === 'created') && progress > 0}
							<tr class="relative transition-colors hover:bg-gray-50/50" style={isInterrupted ? `background:linear-gradient(to right, #dbeafe ${progress}%, transparent ${progress}%)` : ''}>
								<td class="relative w-10 px-4 py-3">
									<button type="button" onclick={() => toggleSelect(task.slug)} class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {selected.has(task.slug) ? 'border-blue-500 bg-blue-500 text-white' : 'border-gray-300'}">
										{#if selected.has(task.slug)}
											<Check size={12} />
										{/if}
									</button>
								</td>
								<td class="relative max-w-[200px] truncate px-4 py-3 text-gray-900" title={task.fileName || task.slug}>
									{task.fileName || task.slug}
								</td>
								<td class="relative px-4 py-3 tabular-nums text-gray-500">{fmtSize(task.fileSize)}</td>
								<td class="relative px-4 py-3">
									<span class="inline-flex items-center gap-1.5">
										<StatusIcon size={14} class={statusClass(task.status)} />
										<span class={statusClass(task.status)}>{statusLabel(task.status)}</span>
										{#if isInterrupted}
											<span class="text-xs text-blue-500">{progress}%</span>
										{/if}
									</span>
									{#if task.errorMsg}
										<p class="mt-0.5 text-xs text-red-400">{task.errorMsg}</p>
									{/if}
								</td>
								<td class="relative px-4 py-3 tabular-nums text-gray-400">
									{#if task.parentSlug}
											<a href="/files/all/{task.parentSlug}" class="text-blue-600 hover:underline" title={task.parentName ? `${task.parentName} (${task.parentSlug})` : task.parentSlug}>{task.parentName || task.parentSlug}</a>
									{:else}
										<span class="text-gray-300">/</span>
									{/if}
								</td>
								<td class="relative px-4 py-3 tabular-nums text-gray-400">{fmtTime(task.createdAt)}</td>
								<td class="relative px-4 py-3">
									<div class="flex items-center gap-1">
										{#if task.status === 'failed'}
											<button
												type="button"
												onclick={() => handleRetry(task)}
												disabled={retrying[task.slug]}
												class="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white shadow-sm transition-colors hover:bg-blue-700 disabled:opacity-50"
											>
												{#if retrying[task.slug]}
													<LoaderCircle size={12} class="animate-spin" />
												{/if}
												{m.upload_retry()}
											</button>
										{/if}
										<button type="button" onclick={() => handleDeleteSingle(task.slug)} class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-red-500" title={m.delete_btn()}>
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
				<div class="flex items-center justify-center gap-2 text-sm">
					<button
						type="button"
						onclick={() => goToPage(currentPage - 1)}
						disabled={currentPage <= 1}
						class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 disabled:opacity-30"
					>
						<ArrowLeft size={14} /> {m.prev()}
					</button>
					{#each Array(totalPages) as _, i}
						<button
							type="button"
							onclick={() => goToPage(i + 1)}
							class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-colors {i + 1 === currentPage ? 'bg-blue-600 font-medium text-white' : 'text-gray-600 hover:bg-gray-100'}"
						>
							{i + 1}
						</button>
					{/each}
					<button
						type="button"
						onclick={() => goToPage(currentPage + 1)}
						disabled={currentPage >= totalPages}
						class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 disabled:opacity-30"
					>
						{m.next()} <ArrowRight size={14} />
					</button>
				</div>
			{/if}
		{/if}
	</div>
{/if}
