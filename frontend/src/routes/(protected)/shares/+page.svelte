<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { Ban, CalendarClock, ChevronDown, Copy, File, Globe, LoaderCircle, Lock, Trash2 } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { toast } from 'svelte-sonner';
	import { cancelShare, deleteShare, listShares, updateShare, type ShareItem } from '$lib/api/shares';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';
	import { copyToClipboard, clipboardUnavailableReason, fmtSize, fmtTime } from '$lib/utils/format';
	import AlertDialog from '$lib/ui/alert-dialog/AlertDialog.svelte';
	import * as m from '$lib/paraglide/messages';

	const expiryLabels: Record<ExpiryChoice, string> = {
		keep: m.share_select_expiry(),
		'1d': m.share_1d_desc(),
		'7d': m.share_7d_desc(),
		'30d': m.share_30d_desc(),
		forever: m.share_forever(),
		custom: m.share_custom(),
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
			toast.error(m.share_load_failed());
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
			toast.error(m.share_select_new_expiry());
			return;
		}
		savingSlug = share.slug;
		try {
			const updated = await updateShare(share.slug, { expiresAt });
			shares = shares.map((item) => (item.slug === updated.slug ? updated : item));
			setExpiryChoice(share.slug, 'keep');
			toast.success(m.share_expiry_updated());
		} catch (error) {
			console.error(error);
			toast.error(m.share_expiry_update_failed());
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
			toast.success(m.share_deleted());
		} catch (error) {
			console.error(error);
			toast.error(m.share_delete_failed());
		} finally {
			savingSlug = null;
		}
	}

	function confirmCancel(share: ShareItem) {
		cancelTarget = share;
		cancelDescription = m.share_confirm_cancel_desc({ name: cancelNames(share) });
		showCancelDialog = true;
	}

	async function disableShare(share: ShareItem) {
		savingSlug = share.slug;
		cancelTarget = null;
		showCancelDialog = false;
		try {
			await cancelShare(share.slug);
			await loadShares(false);
			toast.success(m.share_cancelled());
		} catch (error) {
			console.error(error);
			toast.error(m.share_cancel_failed());
		} finally {
			savingSlug = null;
		}
	}

	function cancelNames(share: ShareItem) {
		return share.files.map((f) => f.fileName).join('、');
	}

	async function copyShareLink(share: ShareItem) {
		const ok = await copyToClipboard(shareLink(share.slug));
		if (ok) toast.success(m.share_link_copied());
		else toast.error(clipboardUnavailableReason());
	}

	function navigateToFile(file: { parentSlug: string }) {
		if (file.parentSlug) {
			void goto(`/files/all/${file.parentSlug}`);
		}
	}

	function statusText(share: ShareItem) {
		if (share.disabledAt) return m.share_status_cancelled();
		if (share.isExpired) return m.share_status_expired();
		return m.share_status_active();
	}

	function statusClass(share: ShareItem) {
		if (share.disabledAt) return 'bg-surface-sunken text-ink-3';
		if (share.isExpired) return 'bg-warning-soft text-warning';
		return 'bg-success-soft text-success';
	}

	function totalSize(share: ShareItem) {
		return share.files.reduce((sum, f) => sum + f.fileSize, 0);
	}

	function fileCountText(share: ShareItem) {
		if (share.files.length === 0) return m.share_no_files();
		return m.share_files_count_detail({ count: String(share.files.length) });
	}

	function expiryClass(share: ShareItem) {
		if (!share.expiresAt) return 'bg-surface-muted text-ink-3';
		const remaining = new Date(share.expiresAt).getTime() - Date.now();
		const days = remaining / 86400000;
		if (days > 7) return 'bg-success-soft text-success';
		if (days > 1) return 'bg-warning-soft text-warning';
		return 'bg-danger-soft text-danger';
	}

	function subFileNames(share: ShareItem) {
		if (share.files.length <= 3) return share.files.map((f) => f.fileName).join('、');
		return share.files.slice(0, 3).map((f) => f.fileName).join('、') + m.share_files_more({ count: String(share.files.length) });
	}
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-semibold text-ink">{m.shares_title()}</h1>
			<p class="mt-1 text-sm text-ink-3">{m.shares_subtitle()}</p>
		</div>
		<div class="rounded-full bg-white px-3 py-1.5 text-sm text-ink-3 ">{m.shares_total({ total: String(total) })}</div>
	</div>

	<section class="overflow-hidden rounded-xl border border-line bg-white">
		{#if loading}
			<div class="flex items-center justify-center py-24">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if shares.length === 0}
			<div class="flex flex-col items-center justify-center px-6 py-24 text-center">
				<img src={noFilesSvg} class="mb-3 w-32 h-32" alt="" />
				<p class="text-sm text-ink-3">{m.shares_empty()}</p>
			</div>
		{:else}
			<div class="divide-y divide-line-soft">
				{#each shares as share (share.slug)}
					{@const choice = expiryChoices[share.slug] ?? 'keep'}
					{@const isPrivate = share.hasPassword}
					{@const isDisabled = Boolean(share.disabledAt) || share.isExpired}
					<div class="grid gap-4 px-5 py-4 lg:grid-cols-[minmax(0,1fr)_320px]">
						<div class="min-w-0 space-y-3">
							<div class="flex items-start gap-3">
								<div class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl {isPrivate ? 'bg-surface-sunken text-ink-2' : 'bg-success-soft text-success'}">
									{#if isPrivate}<Lock size={18} />{:else}<Globe size={18} />{/if}
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-ink">{fileCountText(share)}</p>
									<p class="mt-1 text-xs text-ink-3">{fmtSize(totalSize(share))} · {m.share_created_at()} {fmtTime(share.createdAt)}</p>
								</div>
								<span class="shrink-0 rounded-full px-2 py-1 text-xs font-medium {statusClass(share)}">{statusText(share)}</span>
							</div>

							<div class="flex flex-wrap items-center gap-2 text-xs">
								<span class="inline-flex items-center gap-1 rounded-full px-2 py-1 {expiryClass(share)}">
									<CalendarClock size={13} /> {m.share_expiry_label()} {share.expiresAt ? fmtTime(share.expiresAt) : m.share_forever()}
								</span>
								{#if isPrivate}
									<span class="inline-flex items-center gap-1 rounded-full bg-surface-sunken px-2.5 py-1 font-medium text-ink-2 ring-1 ring-line">
										<Lock size={13} /> {m.share_private()} {share.passwordCode ? `· ${m.share_password_code()} ${share.passwordCode}` : ''}
										{#if share.passwordCode}
											<button type="button" onclick={async () => { const ok = await copyToClipboard(share.passwordCode!); if (ok) toast.success(m.share_password_copied()); else toast.error(clipboardUnavailableReason()); }} class="-mr-0.5 ml-0.5 inline-flex cursor-pointer items-center justify-center rounded-md p-0.5 text-ink-2 transition-colors hover:bg-line hover:text-ink-2">
												<Copy size={12} />
											</button>
										{/if}
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 rounded-full bg-success-soft px-2.5 py-1 font-medium text-success ring-1 ring-success">
										<Globe size={13} /> {m.share_public_no_password()}
									</span>
								{/if}
							</div>

								{#if share.files.length > 0}
								{@const expanded = expandedFiles[share.slug] ?? false}
								{@const displayFiles = expanded ? share.files : share.files.slice(0, 5)}
								<div class={expanded ? 'max-h-48 overflow-y-auto' : ''}>
									<div class="flex flex-wrap gap-1.5">
										{#each displayFiles as sf}
											<button type="button" onclick={() => navigateToFile(sf)} class="inline-flex cursor-pointer items-center gap-1 rounded-md bg-surface-muted px-2 py-1 text-xs text-ink-3 transition-colors hover:bg-primary-soft hover:text-primary">
												<File size={12} /> {sf.fileName}
											</button>
										{/each}
										{#if !expanded && share.files.length > 5}
											<button type="button" onclick={() => (expandedFiles = { ...expandedFiles, [share.slug]: true })} class="inline-flex cursor-pointer items-center rounded-md bg-surface-sunken px-2 py-1 text-xs text-ink-4 transition-colors hover:bg-line hover:text-ink-3">
												+{share.files.length - 5} {m.share_files_more({ count: String(share.files.length - 5) })}
											</button>
										{/if}
										{#if expanded && share.files.length > 5}
											<button type="button" onclick={() => (expandedFiles = { ...expandedFiles, [share.slug]: false })} class="inline-flex cursor-pointer items-center rounded-md bg-surface-sunken px-2 py-1 text-xs text-ink-4 transition-colors hover:bg-line hover:text-ink-3">
												{m.share_collapse()}
											</button>
										{/if}
									</div>
								</div>
							{/if}

							<div class="flex min-w-0 gap-2">
								<input readonly value={shareLink(share.slug)} class="h-9 min-w-0 flex-1 rounded-lg border {isPrivate ? 'border-line bg-surface-sunken/50' : 'border-success bg-success-soft/50'} px-3 text-xs {isPrivate ? 'text-ink-2' : 'text-success'} placeholder-ink-4 outline-none" />
								<button type="button" onclick={() => copyShareLink(share)} class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-line px-3 text-sm text-ink-3 transition-colors hover:bg-surface-muted">
									<Copy size={14} /> {m.share_copy_link()}
								</button>
							</div>
						</div>

						<div class="space-y-3 rounded-xl border border-line-soft bg-surface-muted p-3">
							<p class="text-xs font-medium uppercase tracking-wide text-ink-4">{m.share_modify_expiry()}</p>
							<div class="flex gap-2">
								<Dropdown
									disabled={isDisabled}
									triggerClass="flex h-9 min-w-0 flex-1 items-center gap-1 rounded-lg border border-line bg-white px-2.5 text-sm text-ink-2 transition-colors {isDisabled ? 'cursor-not-allowed opacity-50' : 'hover:border-line hover:bg-surface-sunken'}"
									contentClass="min-w-[180px]"
								>
									{#snippet trigger()}
										<span class="truncate">{expiryLabels[choice]}</span>
										<ChevronDown size={14} class="ml-auto shrink-0 text-ink-4" />
									{/snippet}
									{#each Object.entries(expiryLabels) as [value, label]}
										<DropdownBase.Item onSelect={() => setExpiryChoice(share.slug, value as ExpiryChoice)}>
											<span class={choice === value ? 'font-medium text-ink' : ''}>{label}</span>
										</DropdownBase.Item>
									{/each}
								</Dropdown>
								<button type="button" onclick={() => saveExpiry(share)} disabled={isDisabled || savingSlug === share.slug} class="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-60">
									{#if savingSlug === share.slug}<LoaderCircle size={14} class="animate-spin" />{/if}
									{m.share_save()}
								</button>
							</div>
							{#if choice === 'custom'}
								<input type="datetime-local" value={customExpiries[share.slug] ?? ''} oninput={(event) => setCustomExpiry(share.slug, (event.currentTarget as HTMLInputElement).value)} disabled={isDisabled} class="h-9 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-primary disabled:cursor-not-allowed disabled:opacity-50" />
							{/if}

							<button type="button" onclick={() => confirmCancel(share)} disabled={isDisabled || savingSlug === share.slug} class="inline-flex h-9 w-full items-center justify-center gap-1.5 rounded-lg border border-danger bg-white px-3 text-sm text-danger hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-50">
								<Ban size={14} /> {m.share_cancel_share()}
							</button>
							<button type="button" onclick={() => { deleteTarget = share; deleteDescription = m.share_confirm_delete_desc({ name: cancelNames(share) }); showDeleteDialog = true; }} disabled={savingSlug === share.slug} class="inline-flex h-9 w-full items-center justify-center gap-1.5 rounded-lg bg-danger px-3 text-sm font-medium text-white hover:bg-danger-hover disabled:cursor-not-allowed disabled:opacity-50">
								<Trash2 size={14} /> {m.share_delete_share()}
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
	title={m.share_confirm_cancel_title()}
	description={cancelDescription}
	confirmText={m.share_confirm_cancel_btn()}
	cancelText={m.share_confirm_cancel_cancel()}
	variant="destructive"
	onConfirm={() => { if (cancelTarget) void disableShare(cancelTarget); }}
	onCancel={() => { showCancelDialog = false; }}
/>

<AlertDialog
	bind:open={showDeleteDialog}
	title={m.share_confirm_delete_title()}
	description={deleteDescription}
	confirmText={m.share_confirm_delete_btn()}
	cancelText={m.share_confirm_delete_cancel()}
	variant="destructive"
	onConfirm={() => { if (deleteTarget) void handleDeleteShare(deleteTarget); }}
	onCancel={() => { deleteTarget = null; showDeleteDialog = false; }}
/>
