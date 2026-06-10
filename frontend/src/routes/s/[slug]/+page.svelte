<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getPublicShare, publicShareFileUrl, verifyPublicShare, type PublicShareInfo } from '$lib/api/shares';
	import { fmtSize, fmtTime } from '$lib/utils/format';
	import { Download, File, KeyRound, LoaderCircle, ShieldCheck } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';

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
			if (error instanceof ApiError && error.status === 404) {
				errorMessage = '分享不存在、已取消或已过期';
			} else {
				errorMessage = '加载分享失败';
			}
		} finally {
			loading = false;
		}
	}

	async function verifyPassword() {
		if (!info || verifying) return;
		if (!passwordCode.trim()) {
			toast.error('请输入提取码');
			return;
		}
		verifying = true;
		try {
			info = await verifyPublicShare(info.slug, passwordCode.trim());
			verified = true;
			toast.success('提取码验证成功');
		} catch {
			verified = false;
			toast.error('提取码错误');
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
	<title>{info ? (isMultiFile ? `${info.files.length} 个文件` : info.files[0]?.fileName) : '文件分享'}</title>
</svelte:head>

<div class="min-h-screen bg-[#f5f7fb] px-4 py-10 text-gray-900">
	<div class="mx-auto max-w-4xl space-y-5">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-sm shadow-blue-200">
				<ShieldCheck size={20} />
			</div>
			<div>
				<h1 class="text-lg font-semibold">Netdisk 文件分享</h1>
				<p class="text-sm text-gray-500">打开链接即可查看或下载分享文件</p>
			</div>
		</div>

		<section class="overflow-hidden rounded-3xl border border-gray-100 bg-white shadow-sm shadow-gray-100/80">
			{#if loading}
				<div class="flex items-center justify-center py-24">
					<LoaderCircle size={26} class="animate-spin text-gray-300" />
				</div>
			{:else if errorMessage}
				<div class="flex flex-col items-center justify-center px-6 py-24 text-center">
					<File size={48} class="mb-4 text-gray-300" />
					<p class="text-base text-gray-600">{errorMessage}</p>
				</div>
			{:else if info}
				{#if info.hasPassword && !verified}
					<form class="mx-auto max-w-sm space-y-4 px-6 py-16" onsubmit={(event) => { event.preventDefault(); void verifyPassword(); }}>
						<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-blue-600">
							<KeyRound size={22} />
						</div>
						<div class="text-center">
							<p class="text-sm font-medium text-gray-800">此分享需要提取码</p>
							<p class="mt-1 text-xs text-gray-500">请输入分享者提供的提取码后查看文件</p>
						</div>
						<input
							bind:value={passwordCode}
							maxlength="16"
							placeholder="请输入提取码"
							class="h-11 w-full rounded-xl border border-gray-200 px-4 text-center text-sm tracking-widest outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
						/>
						<button type="submit" disabled={verifying} class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-60">
							{#if verifying}<LoaderCircle size={16} class="animate-spin" />{/if}
							验证并查看
						</button>
					</form>
				{:else}
					<div class="px-6 py-5">
						<p class="text-lg font-semibold text-gray-900">{isMultiFile ? `共 ${info.files.length} 个文件` : info.files[0]?.fileName}</p>
						<p class="mt-1 text-sm text-gray-500">有效期 {info.expiresAt ? fmtTime(info.expiresAt) : '永久'}</p>
					</div>

					<div class="border-t border-gray-100">
						{#each info.files as fileItem}
							<div class="flex items-center gap-4 px-6 py-4 {info.files.length > 1 ? 'border-b border-gray-50 last:border-b-0' : ''}">
								<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gray-50 text-gray-400">
									<File size={18} />
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-gray-900">{fileItem.fileName}</p>
									<p class="text-xs text-gray-500">{fmtSize(fileItem.fileSize)} · {fileItem.mimeType || '未知类型'}</p>
								</div>
								<div class="flex gap-2">
									{#if isPreviewable(fileItem.mimeType)}
										<a href={fileViewUrl(fileItem.fileSlug)} target="_blank" class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 hover:bg-gray-50">
											预览
										</a>
									{/if}
									<a href={fileDownloadUrl(fileItem.fileSlug)} class="inline-flex h-9 items-center gap-1.5 rounded-lg bg-blue-600 px-3 text-sm font-medium text-white hover:bg-blue-700">
										<Download size={15} /> 下载
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
