<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getPublicShare, publicShareFileUrl, verifyPublicShare, type PublicShareInfo } from '$lib/api/shares';
	import { fmtSize, fmtTime } from '$lib/utils/format';
	import { Download, File, KeyRound, LoaderCircle, ShieldCheck } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';

	let info = $state<PublicShareInfo | null>(null);
	let loading = $state(true);
	let verifying = $state(false);
	let errorMessage = $state('');
	let passwordCode = $state('');
	let verified = $state(false);

	let slug = $derived(page.params.slug ?? '');
	let canAccess = $derived(Boolean(info && (!info.hasPassword || verified)));
	let isMultiFile = $derived(info ? info.files.length > 1 : false);

	onMount(() => {
		void loadShare();
	});

	async function loadShare() {
		loading = true;
		errorMessage = '';
		try {
			info = await getPublicShare(slug);
			verified = !info.hasPassword;
		} catch (error) {
			if (error instanceof ApiError && error.errCode === 1001) {
				errorMessage = m.share_not_found();
			} else {
				errorMessage = m.share_load_failed();
			}
		} finally {
			loading = false;
		}
	}

	async function verifyPassword() {
		if (!info || verifying) return;
		if (!passwordCode.trim()) {
			toast.error(m.share_enter_password());
			return;
		}
		verifying = true;
		try {
			info = await verifyPublicShare(info.slug, passwordCode.trim());
			verified = true;
			toast.success(m.share_password_verified());
		} catch {
			verified = false;
			toast.error(m.share_password_wrong());
		} finally {
			verifying = false;
		}
	}

	function fileViewUrl(fileSlug: string) {
		if (!info) return '';
		return publicShareFileUrl(info.slug, fileSlug, info.hasPassword ? passwordCode.trim() : undefined, false);
	}

	function fileDownloadUrl(fileSlug: string) {
		if (!info) return '';
		return publicShareFileUrl(info.slug, fileSlug, info.hasPassword ? passwordCode.trim() : undefined, true);
	}

	function isPreviewable(mimeType: string | null) {
		if (!mimeType) return false;
		return mimeType.startsWith('image/') || mimeType.startsWith('video/') || mimeType.startsWith('audio/') || mimeType === 'application/pdf';
	}
</script>

<svelte:head>
	<title>{info ? (isMultiFile ? m.share_files_count({ count: String(info.files.length) }) : info.files[0]?.fileName) : m.share_page_title()}</title>
</svelte:head>

<div class="min-h-screen bg-surface-muted px-4 py-10 text-ink">
	<div class="mx-auto max-w-4xl space-y-5">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-primary-on shadow-pop">
				<ShieldCheck size={20} />
			</div>
			<div>
				<h1 class="text-lg font-semibold">{m.share_page_title()}</h1>
				<p class="text-sm text-ink-3">{m.share_page_subtitle()}</p>
			</div>
		</div>

		<section class="overflow-hidden rounded-xl border border-line bg-surface">
			{#if loading}
				<div class="flex items-center justify-center py-24">
					<LoaderCircle size={26} class="animate-spin text-ink-4" />
				</div>
			{:else if errorMessage}
				<div class="flex flex-col items-center justify-center px-6 py-24 text-center">
					<File size={48} class="mb-4 text-ink-4" />
					<p class="text-base text-ink-3">{errorMessage}</p>
				</div>
			{:else if info}
				{#if info.hasPassword && !verified}
					<form class="mx-auto max-w-sm space-y-4 px-6 py-16" onsubmit={(event) => { event.preventDefault(); void verifyPassword(); }}>
						<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary-soft text-primary">
							<KeyRound size={22} />
						</div>
						<div class="text-center">
							<p class="text-sm font-medium text-ink-2">{m.share_requires_password()}</p>
							<p class="mt-1 text-xs text-ink-3">{m.share_requires_password_desc()}</p>
						</div>
						<input
							bind:value={passwordCode}
							maxlength="16"
							placeholder={m.share_enter_password()}
							class="h-11 w-full rounded-xl border border-line px-4 text-center text-sm tracking-widest outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
						/>
						<button type="submit" disabled={verifying} class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:opacity-60">
							{#if verifying}<LoaderCircle size={16} class="animate-spin" />{/if}
							{m.share_verify_btn()}
						</button>
					</form>
				{:else}
					<div class="px-6 py-5">
						<p class="text-lg font-semibold text-ink">{isMultiFile ? m.share_files_count({ count: String(info.files.length) }) : info.files[0]?.fileName}</p>
						<p class="mt-1 text-sm text-ink-3">{m.share_expiry_info({ expiry: info.expiresAt ? fmtTime(info.expiresAt) : m.share_forever() })}</p>
					</div>

					<div class="border-t border-line-soft">
						{#each info.files as fileItem}
							<div class="flex items-center gap-4 px-6 py-4 {info.files.length > 1 ? 'border-b border-line-soft last:border-b-0' : ''}">
								<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-surface-sunken text-ink-4">
									<File size={18} />
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-ink">{fileItem.fileName}</p>
									<p class="text-xs text-ink-3">{fmtSize(fileItem.fileSize)} · {fileItem.mimeType || m.share_unknown_type()}</p>
								</div>
								<div class="flex gap-2">
									{#if isPreviewable(fileItem.mimeType)}
										<a href={fileViewUrl(fileItem.fileSlug)} target="_blank" class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-line px-3 text-sm text-ink-3 hover:bg-surface-sunken">
											{m.share_preview()}
										</a>
									{/if}
									<a href={fileDownloadUrl(fileItem.fileSlug)} class="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-medium text-primary-on hover:bg-primary-hover">
										<Download size={15} /> {m.download()}
									</a>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		</section>
	</div>
</div>
