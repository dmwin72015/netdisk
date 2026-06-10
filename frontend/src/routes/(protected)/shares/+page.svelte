<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { Ban, CalendarClock, ChevronDown, Copy, File, Globe, LoaderCircle, Lock, Share2, Trash2 } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { toast } from 'svelte-sonner';
	import { cancelShare, deleteShare, listShares, updateShare, type ShareItem } from '$lib/api/shares';
	import { copyToClipboard, fmtSize, fmtTime } from '$lib/utils/format';
	import AlertDialog from '$lib/ui/alert-dialog/AlertDialog.svelte';

	const expiryLabels: Record<ExpiryChoice, string> = {
		keep: '选择新的有效期',
		'1d': '从现在起 1 天',
		'7d': '从现在起 7 天',
		'30d': '从现在起 30 天',
		forever: '永久',
		custom: '自定义截止时间',
	};

	type ExpiryChoice = 'keep' | '1d' | '7d' | '30d' | 'forever' | 'custom';

	let shares = $state<ShareItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let savingSlug = $state<string | null>(null);
	let expiryChoices = $state<Record<string, ExpiryChoice>>({});
	let customExpiries = $state<Record<string, string>>({});
	let expandedFiles = $state<Record<string, boolean>>({});
	let cancelTarget = $state<ShareItem | null>(null);
	let showCancelDialog = $state(false);
	let cancelDescription = $state('');
	let deleteTarget = $state<ShareItem | null>(null);
	let showDeleteDialog = $state(false);
	let deleteDescription = $state('');

	onMount(() => {
		void loadShares();
	});

	async function loadShares(showLoading = true) {
		if (showLoading) loading = true;
		try {
			const data = await listShares(1, 100);
			shares = data.shares;
			total = data.total;
		} catch (error) {
			console.error(error);
			toast.error('加载分享列表失败');
		} finally {
			loading = false;
		}
	}

	function shareLink(slug: string) {
		return browser ? `${window.location.origin}/s/${slug}` : `/s/${slug}`;
	}

	function setExpiryChoice(slug: string, value: ExpiryChoice) {
		expiryChoices = { ...expiryChoices, [slug]: value };
	}

	function setCustomExpiry(slug: string, value: string) {
		customExpiries = { ...customExpiries, [slug]: value };
	}

	function resolveExpiresAt(slug: string) {
		const choice = expiryChoices[slug] ?? 'keep';
		if (choice === 'keep') return undefined;
		if (choice === 'forever') return null;
		if (choice === 'custom') {
			const customValue = customExpiries[slug];
			return customValue ? new Date(customValue).toISOString() : undefined;
		}
		const days = choice === '1d' ? 1 : choice === '7d' ? 7 : 30;
		return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
	}

	async function saveExpiry(share: ShareItem) {
		const expiresAt = resolveExpiresAt(share.slug);
		if (expiresAt === undefined) {
			toast.error('请选择新的有效期');
			return;
		}
		savingSlug = share.slug;
		try {
			const updated = await updateShare(share.slug, { expiresAt });
			shares = shares.map((item) => (item.slug === updated.slug ? updated : item));
			setExpiryChoice(share.slug, 'keep');
			toast.success('有效期已更新');
		} catch (error) {
			console.error(error);
			toast.error('更新有效期失败');
		} finally {
			savingSlug = null;
		}
	}

	async function handleDeleteShare(share: ShareItem) {
		savingSlug = share.slug;
		deleteTarget = null;
		showDeleteDialog = false;
		try {
			await deleteShare(share.slug);
			await loadShares(false);
			toast.success('分享已删除');
		} catch (error) {
			console.error(error);
			toast.error('删除分享失败');
		} finally {
			savingSlug = null;
		}
	}

	function confirmCancel(share: ShareItem) {
		cancelTarget = share;
		cancelDescription = `确定取消「${cancelNames(share)}」的分享吗？`;
		showCancelDialog = true;
	}

	async function disableShare(share: ShareItem) {
		savingSlug = share.slug;
		cancelTarget = null;
		showCancelDialog = false;
		try {
			await cancelShare(share.slug);
			await loadShares(false);
			toast.success('分享已取消');
		} catch (error) {
			console.error(error);
			toast.error('取消分享失败');
		} finally {
			savingSlug = null;
		}
	}

	function cancelNames(share: ShareItem) {
		return share.files.map((f) => f.fileName).join('、');
	}

	async function copyShareLink(share: ShareItem) {
		const ok = await copyToClipboard(shareLink(share.slug));
		if (ok) toast.success('分享链接已复制');
	}

	function navigateToFile(file: { parentSlug: string }) {
		if (file.parentSlug) {
			void goto(`/files/all/${file.parentSlug}`);
		}
	}

	function statusText(share: ShareItem) {
		if (share.disabledAt) return '已取消';
		if (share.isExpired) return '已过期';
		return '生效中';
	}

	function statusClass(share: ShareItem) {
		if (share.disabledAt) return 'bg-gray-100 text-gray-500';
		if (share.isExpired) return 'bg-amber-50 text-amber-700';
		return 'bg-emerald-50 text-emerald-700';
	}

	function totalSize(share: ShareItem) {
		return share.files.reduce((sum, f) => sum + f.fileSize, 0);
	}

	function fileCountText(share: ShareItem) {
		if (share.files.length === 0) return '无文件';
		return `${share.files.length} 个文件`;
	}

	function expiryClass(share: ShareItem) {
		if (!share.expiresAt) return 'bg-gray-50 text-gray-500';
		const remaining = new Date(share.expiresAt).getTime() - Date.now();
		const days = remaining / 86400000;
		if (days > 7) return 'bg-emerald-50 text-emerald-600';
		if (days > 1) return 'bg-amber-50 text-amber-600';
		return 'bg-red-50 text-red-600';
	}

	function subFileNames(share: ShareItem) {
		if (share.files.length <= 3) return share.files.map((f) => f.fileName).join('、');
		return share.files.slice(0, 3).map((f) => f.fileName).join('、') + ` 等${share.files.length}项`;
	}
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-semibold text-gray-900">我的分享</h1>
			<p class="mt-1 text-sm text-gray-500">管理已生成的文件分享链接、有效期和状态</p>
		</div>
		<div class="rounded-full bg-white px-3 py-1.5 text-sm text-gray-500 shadow-sm">共 {total} 个分享</div>
	</div>

	<section class="overflow-hidden rounded-3xl border border-gray-100 bg-white/80 shadow-sm shadow-gray-100/80">
		{#if loading}
			<div class="flex items-center justify-center py-24">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if shares.length === 0}
			<div class="flex flex-col items-center justify-center px-6 py-24 text-center">
				<Share2 size={48} class="mb-4 text-gray-300" />
				<p class="text-sm text-gray-500">暂无分享记录</p>
			</div>
		{:else}
			<div class="divide-y divide-gray-100">
				{#each shares as share (share.slug)}
					{@const choice = expiryChoices[share.slug] ?? 'keep'}
					{@const isPrivate = share.hasPassword}
					{@const isDisabled = Boolean(share.disabledAt) || share.isExpired}
					<div class="grid gap-4 px-5 py-4 lg:grid-cols-[minmax(0,1fr)_320px]">
						<div class="min-w-0 space-y-3">
							<div class="flex items-start gap-3">
								<div class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl {isPrivate ? 'bg-purple-50 text-purple-600' : 'bg-emerald-50 text-emerald-600'}">
									{#if isPrivate}<Lock size={18} />{:else}<Globe size={18} />{/if}
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-gray-900">{fileCountText(share)}</p>
									<p class="mt-1 text-xs text-gray-500">{fmtSize(totalSize(share))} · 创建于 {fmtTime(share.createdAt)}</p>
								</div>
								<span class="shrink-0 rounded-full px-2 py-1 text-xs font-medium {statusClass(share)}">{statusText(share)}</span>
							</div>

							<div class="flex flex-wrap items-center gap-2 text-xs">
								<span class="inline-flex items-center gap-1 rounded-full px-2 py-1 {expiryClass(share)}">
									<CalendarClock size={13} /> 有效期 {share.expiresAt ? fmtTime(share.expiresAt) : '永久'}
								</span>
								{#if isPrivate}
									<span class="inline-flex items-center gap-1 rounded-full bg-purple-50 px-2.5 py-1 font-medium text-purple-700 ring-1 ring-purple-200">
										<Lock size={13} /> 私密 {share.passwordCode ? `· 提取码 ${share.passwordCode}` : ''}
										{#if share.passwordCode}
											<button type="button" onclick={async () => { const ok = await copyToClipboard(share.passwordCode!); if (ok) toast.success('提取码已复制'); }} class="-mr-0.5 ml-0.5 inline-flex cursor-pointer items-center justify-center rounded-md p-0.5 text-purple-400 transition-colors hover:bg-purple-200 hover:text-purple-700">
												<Copy size={12} />
											</button>
										{/if}
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 font-medium text-emerald-700 ring-1 ring-emerald-200">
										<Globe size={13} /> 公开 · 无需提取码
									</span>
								{/if}
							</div>

								{#if share.files.length > 0}
								{@const expanded = expandedFiles[share.slug] ?? false}
								{@const displayFiles = expanded ? share.files : share.files.slice(0, 5)}
								<div class={expanded ? 'max-h-48 overflow-y-auto' : ''}>
									<div class="flex flex-wrap gap-1.5">
										{#each displayFiles as sf}
											<button type="button" onclick={() => navigateToFile(sf)} class="inline-flex cursor-pointer items-center gap-1 rounded-md bg-gray-50 px-2 py-1 text-xs text-gray-600 transition-colors hover:bg-blue-50 hover:text-blue-700">
												<File size={12} /> {sf.fileName}
											</button>
										{/each}
										{#if !expanded && share.files.length > 5}
											<button type="button" onclick={() => (expandedFiles = { ...expandedFiles, [share.slug]: true })} class="inline-flex cursor-pointer items-center rounded-md bg-gray-100 px-2 py-1 text-xs text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600">
												+{share.files.length - 5} 更多
											</button>
										{/if}
										{#if expanded && share.files.length > 5}
											<button type="button" onclick={() => (expandedFiles = { ...expandedFiles, [share.slug]: false })} class="inline-flex cursor-pointer items-center rounded-md bg-gray-100 px-2 py-1 text-xs text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600">
												收起
											</button>
										{/if}
									</div>
								</div>
							{/if}

							<div class="flex min-w-0 gap-2">
								<input readonly value={shareLink(share.slug)} class="h-9 min-w-0 flex-1 rounded-lg border {isPrivate ? 'border-purple-200 bg-purple-50/50' : 'border-emerald-200 bg-emerald-50/50'} px-3 text-xs {isPrivate ? 'text-purple-800' : 'text-emerald-800'} placeholder-gray-400 outline-none" />
								<button type="button" onclick={() => copyShareLink(share)} class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 transition-colors hover:bg-gray-50">
									<Copy size={14} /> 复制链接
								</button>
							</div>
						</div>

						<div class="space-y-3 rounded-2xl border border-gray-100 bg-gray-50 p-3">
							<p class="text-xs font-medium uppercase tracking-wide text-gray-400">修改有效期</p>
							<div class="flex gap-2">
								<Dropdown
									disabled={isDisabled}
									triggerClass="flex h-9 min-w-0 flex-1 items-center gap-1 rounded-lg border border-gray-200 bg-white px-2.5 text-sm text-gray-700 transition-colors {isDisabled ? 'cursor-not-allowed opacity-50' : 'hover:border-gray-300 hover:bg-gray-50'}"
									contentClass="min-w-[180px]"
								>
									{#snippet trigger()}
										<span class="truncate">{expiryLabels[choice]}</span>
										<ChevronDown size={14} class="ml-auto shrink-0 text-gray-400" />
									{/snippet}
									{#each Object.entries(expiryLabels) as [value, label]}
										<DropdownBase.Item onSelect={() => setExpiryChoice(share.slug, value as ExpiryChoice)}>
											<span class={choice === value ? 'font-medium text-gray-900' : ''}>{label}</span>
										</DropdownBase.Item>
									{/each}
								</Dropdown>
								<button type="button" onclick={() => saveExpiry(share)} disabled={isDisabled || savingSlug === share.slug} class="inline-flex h-9 items-center gap-1.5 rounded-lg bg-blue-600 px-3 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
									{#if savingSlug === share.slug}<LoaderCircle size={14} class="animate-spin" />{/if}
									保存
								</button>
							</div>
							{#if choice === 'custom'}
								<input type="datetime-local" value={customExpiries[share.slug] ?? ''} oninput={(event) => setCustomExpiry(share.slug, (event.currentTarget as HTMLInputElement).value)} disabled={isDisabled} class="h-9 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm outline-none focus:border-blue-500 disabled:cursor-not-allowed disabled:opacity-50" />
							{/if}

							<button type="button" onclick={() => confirmCancel(share)} disabled={isDisabled || savingSlug === share.slug} class="inline-flex h-9 w-full items-center justify-center gap-1.5 rounded-lg border border-red-100 bg-white px-3 text-sm text-red-600 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50">
								<Ban size={14} /> 取消分享
							</button>
							<button type="button" onclick={() => { deleteTarget = share; deleteDescription = `确定永久删除「${cancelNames(share)}」的分享吗？此操作不可撤销。`; showDeleteDialog = true; }} disabled={savingSlug === share.slug} class="inline-flex h-9 w-full items-center justify-center gap-1.5 rounded-lg bg-red-600 px-3 text-sm font-medium text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50">
								<Trash2 size={14} /> 删除
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</section>
</div>

<AlertDialog
	bind:open={showCancelDialog}
	title="取消分享"
	description={cancelDescription}
	confirmText="确认取消"
	cancelText="再想想"
	variant="destructive"
	onConfirm={() => { if (cancelTarget) void disableShare(cancelTarget); }}
	onCancel={() => { showCancelDialog = false; }}
/>

<AlertDialog
	bind:open={showDeleteDialog}
	title="删除分享"
	description={deleteDescription}
	confirmText="确认删除"
	cancelText="再想想"
	variant="destructive"
	onConfirm={() => { if (deleteTarget) void handleDeleteShare(deleteTarget); }}
	onCancel={() => { deleteTarget = null; showDeleteDialog = false; }}
/>
