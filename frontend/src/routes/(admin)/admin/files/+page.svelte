<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import {
		ChevronLeft, ChevronRight, LoaderCircle, Search,
		Trash2, RotateCcw, Star, ChevronDown, Check,
	} from '@lucide/svelte';
	import { Select } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import {
		adminListFiles,
		adminDeleteFile,
		adminRestoreFile,
		type AdminFile,
	} from '$lib/api/admin';
	import { fmtSize } from '$lib/utils/format';
	import { confirmDelete } from '$lib/dialog';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 20;

	let files = $state<AdminFile[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let loading = $state(true);
	let searchQuery = $state('');
	let categoryFilter = $state('');
	let trashedFilter = $state('');
	let sortBy = $state('-created_at');

	let currentPage = $derived(Math.floor(offset / PAGE_SIZE) + 1);
	let totalPages = $derived(Math.ceil(total / PAGE_SIZE));

	const categoryOptions = [
		{ value: '', label: m.admin_all_categories() },
		{ value: 'document', label: 'Document' },
		{ value: 'image', label: 'Image' },
		{ value: 'video', label: 'Video' },
		{ value: 'audio', label: 'Audio' },
		{ value: 'archive', label: 'Archive' },
		{ value: 'other', label: 'Other' },
	];

	const trashedOptions = [
		{ value: '', label: m.admin_all_status() },
		{ value: 'no', label: m.admin_active_only() },
		{ value: 'only', label: m.admin_trashed_only() },
	];

	const sortOptions = [
		{ value: '-created_at', label: m.admin_newest_first() },
		{ value: 'created_at', label: m.admin_oldest_first() },
		{ value: 'name', label: m.admin_name_az() },
		{ value: '-name', label: m.admin_name_za() },
		{ value: 'size', label: m.admin_largest_first() },
	];

	onMount(() => {
		if (!browser) return;
		loadFiles();
	});

	async function loadFiles() {
		loading = true;
		try {
			const res = await adminListFiles(PAGE_SIZE, offset, searchQuery || undefined, categoryFilter || undefined, trashedFilter || undefined, sortBy);
			files = res.items;
			total = res.total;
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}

	function goPage(page: number) {
		offset = (page - 1) * PAGE_SIZE;
		loadFiles();
	}

	function handleSearch() {
		offset = 0;
		loadFiles();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleSearch();
	}

	function handleFilterChange() {
		offset = 0;
		loadFiles();
	}

	async function handleDeleteFile(f: AdminFile) {
		const ok = await confirmDelete(m.admin_confirm_delete_file({ name: f.fileName }));
		if (!ok) return;
		try {
			await adminDeleteFile(f.id);
			files = files.filter((x) => x.id !== f.id);
			total--;
			toast.success(m.admin_file_deleted());
		} catch {
			toast.error(m.admin_delete_failed());
		}
	}

	async function handleRestoreFile(f: AdminFile) {
		try {
			await adminRestoreFile(f.id);
			f.isTrashed = false;
			toast.success(m.admin_file_restored());
		} catch {
			toast.error(m.admin_restore_failed());
		}
	}

	function fmtDate(ts: number): string {
		return new Date(ts * 1000).toLocaleString();
	}
</script>

<div class="space-y-5">
	<div>
		<h1 class="text-xl font-bold text-ink">{m.admin_file_management()}</h1>
		<p class="mt-0.5 text-sm text-ink-4">{m.admin_file_management_desc()}</p>
	</div>

	<div class="flex flex-wrap items-center gap-3">
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<Search size={16} class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-4" />
			<input
				bind:value={searchQuery}
				onkeydown={handleKeydown}
				placeholder={m.admin_search_files()}
				class="w-full rounded-lg border border-line bg-surface py-2 pl-9 pr-3 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
			/>
		</div>
		<Select.Root type="single" bind:value={categoryFilter}>
			<Select.Trigger class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[150px] data-[placeholder]:text-ink-4">
				<Select.Value placeholder={m.admin_all_categories()} />
				<ChevronDown size={14} class="text-ink-4" />
			</Select.Trigger>
			<Select.Content class="z-50 w-[var(--bits-select-trigger-width)] overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
				{#each categoryOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === categoryFilter}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
		<Select.Root type="single" bind:value={trashedFilter}>
			<Select.Trigger class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[135px] data-[placeholder]:text-ink-4">
				<Select.Value placeholder={m.admin_all_status()} />
				<ChevronDown size={14} class="text-ink-4" />
			</Select.Trigger>
			<Select.Content class="z-50 w-[var(--bits-select-trigger-width)] overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
				{#each trashedOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === trashedFilter}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
		<Select.Root type="single" bind:value={sortBy}>
			<Select.Trigger class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[155px] data-[placeholder]:text-ink-4">
				<Select.Value />
				<ChevronDown size={14} class="text-ink-4" />
			</Select.Trigger>
			<Select.Content class="z-50 w-[var(--bits-select-trigger-width)] overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
				{#each sortOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === sortBy}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
		<button
			onclick={handleSearch}
			class="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover"
		>
			<Search size={14} />
			{m.admin_search()}
		</button>
		<span class="ml-auto text-sm text-ink-4">{m.total_items({ total: String(total) })}</span>
	</div>

	<div class="overflow-hidden rounded-xl border border-line">
		<table class="w-full table-fixed text-left text-sm">
			<thead class="border-b border-line bg-surface-sunken text-xs text-ink-4">
				<tr>
					<th class="px-4 py-3 font-medium">{m.col_filename()}</th>
					<th class="w-[220px] px-4 py-3 font-medium">{m.admin_owner()}</th>
					<th class="w-[220px] px-4 py-3 font-medium">{m.col_type()}</th>
					<th class="w-[100px] px-4 py-3 font-medium">{m.col_size()}</th>
					<th class="w-[170px] px-4 py-3 font-medium">{m.admin_uploaded()}</th>
					<th class="w-[120px] px-4 py-3 font-medium">{m.status()}</th>
					<th class="w-[110px] px-4 py-3 font-medium">{m.col_actions()}</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-line-soft">
				{#if loading}
					<tr>
						<td colspan="7" class="px-4 py-12 text-center text-ink-4">
							<LoaderCircle size={24} class="mx-auto animate-spin" />
						</td>
					</tr>
				{:else if files.length === 0}
					<tr>
						<td colspan="7" class="px-4 py-12 text-center text-ink-4">{m.no_files()}</td>
					</tr>
				{:else}
					{#each files as f (f.id)}
						<tr class="transition-colors hover:bg-surface-muted" class:opacity-60={f.isTrashed}>
							<td class="w-[320px] px-4 py-3">
								<div class="flex items-center gap-2">
									<MimeIcon
										mimeType={f.mimeType}
										name={f.fileName}
										isDir={f.isDir}
										category={f.fileCategory}
										size={24}
										class="shrink-0"
									/>
									<span class="min-w-0 truncate font-medium text-ink" title={f.fileName}>{f.fileName}</span>
								</div>
							</td>
							<td class="truncate px-4 py-3">
								<button
									class="truncate text-primary hover:underline"
									onclick={() => goto(`/admin/users/${f.userId}`)}
									title={f.username}
								>
									{f.username}
								</button>
							</td>
							<td class="truncate px-4 py-3 text-ink-3" title={f.isDir ? '' : (f.mimeType || f.fileCategory || '')}>
								{f.isDir ? m.admin_directory() : (f.mimeType || f.fileCategory || '-')}
							</td>
							<td class="px-4 py-3 font-mono text-xs text-ink-3">
								{f.isDir ? '-' : fmtSize(f.fileSize)}
							</td>
							<td class="px-4 py-3 text-xs text-ink-4">{fmtDate(f.createdAt)}</td>
							<td class="px-4 py-3">
								<div class="flex items-center gap-2">
									{#if f.isTrashed}
										<span class="rounded-full bg-danger-soft px-2 py-0.5 text-xs text-danger">Trashed</span>
									{/if}
									{#if f.isStarred}
										<Star size={13} class="text-star" />
									{/if}
								</div>
							</td>
							<td class="px-4 py-3">
								<div class="flex items-center gap-1">
									{#if f.isTrashed}
										<button
											class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-success-soft hover:text-success"
											onclick={() => handleRestoreFile(f)}
											title={m.admin_restore()}
										>
											<RotateCcw size={15} />
										</button>
									{/if}
									<button
										class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-danger-soft hover:text-danger"
										onclick={() => handleDeleteFile(f)}
										title={m.admin_delete_permanent()}
									>
										<Trash2 size={15} />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2">
			<button
				class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-4 transition-colors hover:bg-surface-sunken disabled:opacity-40"
				disabled={currentPage <= 1}
				onclick={() => goPage(currentPage - 1)}
			>
				<ChevronLeft size={14} />
			</button>
			<span class="text-sm text-ink-4">{currentPage} / {totalPages}</span>
			<button
				class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-4 transition-colors hover:bg-surface-sunken disabled:opacity-40"
				disabled={currentPage >= totalPages}
				onclick={() => goPage(currentPage + 1)}
			>
				<ChevronRight size={14} />
			</button>
		</div>
	{/if}
</div>
