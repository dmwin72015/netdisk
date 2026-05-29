<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, authReady } from '$lib/stores/auth';
	import { listUploadTasks, retryUploadTask, type UploadTaskItem } from '$lib/api/upload-tasks';
	import { ListRestart, CheckCircle2, XCircle, Clock, Upload, LoaderCircle, ArrowLeft, ArrowRight } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize, fmtTime } from '$lib/utils/format';

	let tasks = $state<UploadTaskItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let limit = $state(20);
	let offset = $state(0);
	let retrying = $state<Record<string, boolean>>({});

	let totalPages = $derived(Math.ceil(total / limit));
	let currentPage = $derived(Math.floor(offset / limit) + 1);

	async function refresh() {
		if (!$user) return;
		loading = true;
		try {
			const data = await listUploadTasks(limit, offset);
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
			case 'done': return CheckCircle2;
			case 'failed': return XCircle;
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

	function goToPage(page: number) {
		offset = (page - 1) * limit;
		refresh();
	}

	onMount(refresh);
</script>

{#if !$authReady}
{:else if $user}
	<div class="space-y-4">
		<h1 class="flex items-center gap-2 text-xl font-semibold">
			<ListRestart size={20} /> {m.upload_title()}
		</h1>

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
			<div class="overflow-hidden rounded-xl border border-gray-100">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-50 bg-gray-50/50 text-left text-xs font-medium text-gray-500">
							<th class="px-4 py-3">{m.col_filename()}</th>
							<th class="px-4 py-3">{m.col_size()}</th>
							<th class="px-4 py-3">{m.status()}</th>
							<th class="px-4 py-3">{m.col_upload_time()}</th>
							<th class="px-4 py-3">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each tasks as task (task.slug)}
							{@const StatusIcon = statusIcon(task.status)}
							<tr class="transition-colors hover:bg-gray-50/50">
								<td class="max-w-[200px] truncate px-4 py-3 text-gray-900" title={task.fileName || task.slug}>
									{task.fileName || task.slug}
								</td>
								<td class="px-4 py-3 tabular-nums text-gray-500">{fmtSize(task.fileSize)}</td>
								<td class="px-4 py-3">
									<span class="inline-flex items-center gap-1.5">
										<StatusIcon size={14} class={statusClass(task.status)} />
										<span class={statusClass(task.status)}>{statusLabel(task.status)}</span>
									</span>
									{#if task.errorMsg}
										<p class="mt-0.5 text-xs text-red-400">{task.errorMsg}</p>
									{/if}
								</td>
								<td class="px-4 py-3 tabular-nums text-gray-400">{fmtTime(task.createdAt)}</td>
								<td class="px-4 py-3">
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
{:else}
	<p class="text-gray-600">{@html m.please_login({ link: '<a href="/login" class="underline">' + m.login_link_text() + '</a>' })}</p>
{/if}
