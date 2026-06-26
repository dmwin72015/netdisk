<script lang="ts">
	import { Check, Copy, File, LoaderCircle, QrCode, RefreshCw } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { browser } from '$app/environment';
	import { Dialog } from '$lib/ui/dialog';
	import { createShare, type ShareItem } from '$lib/api/shares';
	import type { NormalizedFile } from '$lib/types/file';
	import { copyToClipboard, clipboardUnavailableReason, fmtSize } from '$lib/utils/format';
	import { qrCodeDataUrl } from '$lib/utils/qrcode';
	import * as m from '$lib/paraglide/messages';

	type ExpiryOption = '1d' | '7d' | '30d' | 'forever' | 'custom';

	let {
		open = $bindable(false),
		files = [],
	}: {
		open?: boolean;
		files: NormalizedFile[];
	} = $props();

	let totalSize = $derived(files.reduce((sum, f) => sum + f.size, 0));

	let expiryOption = $state<ExpiryOption>('7d');
	let customExpiresAt = $state('');
	let privateShare = $state(true);
	let passwordCode = $state(generatePasswordCode());
	let creating = $state(false);
	let share = $state<ShareItem | null>(null);
	let copied = $state(false);

	let shareLink = $derived(share && browser ? `${window.location.origin}/s/${share.slug}` : '');
	let qrCodeUrl = $derived(shareLink ? qrCodeDataUrl(shareLink) : null);

	$effect(() => {
		if (open) resetForm();
	});

	function resetForm() {
		expiryOption = '7d';
		customExpiresAt = '';
		privateShare = true;
		passwordCode = generatePasswordCode();
		creating = false;
		share = null;
		copied = false;
	}

	function generatePasswordCode() {
		return Math.random().toString(36).slice(2, 6).toUpperCase();
	}

	function refreshPasswordCode() {
		passwordCode = generatePasswordCode();
	}

	function resolveExpiresAt() {
		if (expiryOption === 'forever') return null;
		if (expiryOption === 'custom') {
			if (!customExpiresAt) return undefined;
			return new Date(customExpiresAt).toISOString();
		}
		const days = expiryOption === '1d' ? 1 : expiryOption === '7d' ? 7 : 30;
		return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
	}

	async function submitShare() {
		if (files.length === 0 || creating) return;
		const expiresAt = resolveExpiresAt();
		if (expiresAt === undefined) {
			toast.error(m.share_select_custom_expiry());
			return;
		}
		if (privateShare && passwordCode.trim() === '') {
			toast.error(m.share_enter_password());
			return;
		}

		creating = true;
		try {
			share = await createShare({
				fileSlugs: files.map((f) => f.id),
				expiresAt,
				passwordCode: privateShare ? passwordCode.trim() : null,
			});
			if (share.passwordCode) passwordCode = share.passwordCode;
			toast.success(m.share_create_success());
		} catch (error) {
			console.error(error);
			toast.error(m.share_create_failed());
		} finally {
			creating = false;
		}
	}

	async function copyLink() {
		if (!shareLink) return;
		const ok = await copyToClipboard(shareLink);
		if (ok) {
			copied = true;
			toast.success(m.share_link_copied());
			setTimeout(() => (copied = false), 1200);
		} else {
			toast.error(clipboardUnavailableReason());
		}
	}

	async function copyPasswordCode() {
		if (!share?.hasPassword) return;
		const ok = await copyToClipboard(passwordCode);
		if (ok) toast.success(m.share_password_copied());
		else toast.error(clipboardUnavailableReason());
	}
</script>

<Dialog bind:open title={m.share_dialog_title()} description={files.length === 1 ? files[0]?.name : m.share_files_count({ count: String(files.length) })} footer={false} size="lg">
	<div class="grid gap-6 md:grid-cols-[minmax(0,1fr)_220px]">
		<div class="space-y-5">
			{#if files.length > 0}
				<div class="rounded-xl border border-line-soft bg-surface-muted px-4 py-3">
					{#if files.length === 1}
						<p class="truncate text-sm font-medium text-ink">{files[0].name}</p>
						<p class="mt-1 text-xs text-ink-3">{fmtSize(totalSize)} · {files[0].mimeType || m.share_unknown_type()}</p>
					{:else}
						<p class="text-sm font-medium text-ink">{m.share_files_count({ count: String(files.length) })}</p>
						<p class="mt-1 text-xs text-ink-3">{m.share_total_size({ size: fmtSize(totalSize) })}</p>
						<div class="mt-2 max-h-24 space-y-1 overflow-y-auto">
							{#each files as f}
								<p class="truncate text-xs text-ink-3"><File size={12} class="inline" /> {f.name}</p>
							{/each}
						</div>
					{/if}
				</div>
			{/if}

			{#if !share}
				<section class="space-y-2">
					<p class="text-sm font-medium text-ink-2">{m.share_expiry_label()}</p>
					<div class="grid grid-cols-3 gap-2 sm:grid-cols-5">
						{#each [
							['1d', m.share_1d()],
							['7d', m.share_7d()],
							['30d', m.share_30d()],
							['forever', m.share_forever()],
							['custom', m.share_custom()],
						] as [value, label]}
							<button
								type="button"
								onclick={() => (expiryOption = value as ExpiryOption)}
								class="rounded-lg border px-3 py-2 text-sm transition-colors {expiryOption === value ? 'border-primary bg-primary-soft text-primary' : 'border-line bg-white text-ink-3 hover:border-line'}"
							>
								{label}
							</button>
						{/each}
					</div>
					{#if expiryOption === 'custom'}
						<input
							type="datetime-local"
							bind:value={customExpiresAt}
							class="h-10 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-primary"
						/>
					{/if}
				</section>

				<section class="space-y-3 rounded-xl border border-line-soft p-4">
					<label class="flex items-center justify-between gap-4">
						<span>
							<span class="block text-sm font-medium text-ink-2">{m.share_private_share()}</span>
							<span class="mt-1 block text-xs text-ink-3">{m.share_private_desc()}</span>
						</span>
						<input type="checkbox" bind:checked={privateShare} class="h-4 w-4 rounded border-line text-primary" />
					</label>
					{#if privateShare}
						<div class="flex gap-2">
							<input
								bind:value={passwordCode}
								maxlength="16"
								class="h-10 min-w-0 flex-1 rounded-lg border border-line px-3 text-sm uppercase outline-none focus:border-primary"
							/>
							<button type="button" onclick={refreshPasswordCode} class="inline-flex h-10 items-center gap-1.5 rounded-lg border border-line px-3 text-sm text-ink-3 hover:bg-surface-muted">
								<RefreshCw size={14} /> {m.share_refresh_code()}
							</button>
						</div>
					{/if}
				</section>

				<div class="flex justify-end gap-2 border-t border-line-soft pt-4">
					<button type="button" onclick={() => (open = false)} class="h-9 rounded-lg px-4 text-sm text-ink-3 hover:bg-surface-sunken">{m.cancel()}</button>
					<button type="button" onclick={submitShare} disabled={creating} class="inline-flex h-9 items-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-60">
						{#if creating}<LoaderCircle size={15} class="animate-spin" />{/if}
{m.share_create_link()}
					</button>
				</div>
			{:else}
				<section class="space-y-4">
					<div>
						<p class="mb-2 text-sm font-medium text-ink-2">{m.share_link_label()}</p>
						<div class="flex gap-2">
							<input readonly value={shareLink} class="h-10 min-w-0 flex-1 rounded-lg border border-line bg-surface-muted px-3 text-sm text-ink-2" />
							<button type="button" onclick={copyLink} class="inline-flex h-10 items-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-medium text-white hover:bg-primary-hover">
								{#if copied}<Check size={14} />{:else}<Copy size={14} />{/if} {m.share_copy()}
							</button>
						</div>
					</div>
					{#if share.hasPassword}
						<div>
							<p class="mb-2 text-sm font-medium text-ink-2">{m.share_password_label()}</p>
							<div class="flex gap-2">
								<input readonly value={passwordCode} class="h-10 w-32 rounded-lg border border-line bg-surface-muted px-3 text-sm tracking-widest text-ink-2" />
								<button type="button" onclick={copyPasswordCode} class="inline-flex h-10 items-center gap-1.5 rounded-lg border border-line px-3 text-sm text-ink-3 hover:bg-surface-muted">
									<Copy size={14} /> {m.share_copy_password()}
								</button>
							</div>
						</div>
					{/if}
					<p class="text-xs text-ink-3">{m.share_expiry_info({ expiry: share.expiresAt ? new Date(share.expiresAt).toLocaleString() : m.share_forever() })}</p>
				</section>
			{/if}
		</div>

		<aside class="flex min-h-52 items-center justify-center rounded-xl border border-dashed border-line bg-surface-muted p-4">
			{#if qrCodeUrl}
				<img src={qrCodeUrl} alt={m.share_qr_code()} class="h-44 w-44 rounded-lg bg-white p-2 " />
			{:else if shareLink}
				<div class="text-center text-ink-4">
					<QrCode size={44} class="mx-auto mb-3" />
					<p class="text-sm">{m.share_qr_too_long()}</p>
				</div>
			{:else}
				<div class="text-center text-ink-4">
					<QrCode size={44} class="mx-auto mb-3" />
					<p class="text-sm">{m.share_qr_not_generated()}</p>
				</div>
			{/if}
		</aside>
	</div>
</Dialog>
