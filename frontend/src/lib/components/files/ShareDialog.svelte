<script lang="ts">
	import { Check, Copy, File, LoaderCircle, QrCode, RefreshCw } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { browser } from '$app/environment';
	import { Dialog } from '$lib/ui/dialog';
	import { createShare, type ShareItem } from '$lib/api/shares';
	import type { NormalizedFile } from '$lib/types/file';
	import { copyToClipboard, fmtSize } from '$lib/utils/format';
	import { qrCodeDataUrl } from '$lib/utils/qrcode';

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
			toast.error('请选择自定义截止时间');
			return;
		}
		if (privateShare && passwordCode.trim() === '') {
			toast.error('请输入提取码');
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
			toast.success('分享链接已生成');
		} catch (error) {
			console.error(error);
			toast.error('生成分享失败');
		} finally {
			creating = false;
		}
	}

	async function copyLink() {
		if (!shareLink) return;
		const ok = await copyToClipboard(shareLink);
		if (ok) {
			copied = true;
			toast.success('链接已复制');
			setTimeout(() => (copied = false), 1200);
		}
	}

	async function copyPasswordCode() {
		if (!share?.hasPassword) return;
		const ok = await copyToClipboard(passwordCode);
		if (ok) toast.success('提取码已复制');
	}
</script>

<Dialog bind:open title="分享文件" description={files.length === 1 ? files[0]?.name : `${files.length} 个文件`} footer={false} class="max-w-3xl">
	<div class="grid gap-6 md:grid-cols-[minmax(0,1fr)_220px]">
		<div class="space-y-5">
			{#if files.length > 0}
				<div class="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3">
					{#if files.length === 1}
						<p class="truncate text-sm font-medium text-gray-900">{files[0].name}</p>
						<p class="mt-1 text-xs text-gray-500">{fmtSize(totalSize)} · {files[0].mimeType || '未知类型'}</p>
					{:else}
						<p class="text-sm font-medium text-gray-900">共 {files.length} 个文件</p>
						<p class="mt-1 text-xs text-gray-500">总计 {fmtSize(totalSize)}</p>
						<div class="mt-2 max-h-24 space-y-1 overflow-y-auto">
							{#each files as f}
								<p class="truncate text-xs text-gray-500"><File size={12} class="inline" /> {f.name}</p>
							{/each}
						</div>
					{/if}
				</div>
			{/if}

			{#if !share}
				<section class="space-y-2">
					<p class="text-sm font-medium text-gray-800">有效期</p>
					<div class="grid grid-cols-3 gap-2 sm:grid-cols-5">
						{#each [
							['1d', '1天'],
							['7d', '7天'],
							['30d', '30天'],
							['forever', '永久'],
							['custom', '自定义'],
						] as [value, label]}
							<button
								type="button"
								onclick={() => (expiryOption = value as ExpiryOption)}
								class="rounded-lg border px-3 py-2 text-sm transition-colors {expiryOption === value ? 'border-blue-500 bg-blue-50 text-blue-700' : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'}"
							>
								{label}
							</button>
						{/each}
					</div>
					{#if expiryOption === 'custom'}
						<input
							type="datetime-local"
							bind:value={customExpiresAt}
							class="h-10 w-full rounded-lg border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
						/>
					{/if}
				</section>

				<section class="space-y-3 rounded-xl border border-gray-100 p-4">
					<label class="flex items-center justify-between gap-4">
						<span>
							<span class="block text-sm font-medium text-gray-800">私密分享</span>
							<span class="mt-1 block text-xs text-gray-500">开启后访问者需要输入提取码</span>
						</span>
						<input type="checkbox" bind:checked={privateShare} class="h-4 w-4 rounded border-gray-300 text-blue-600" />
					</label>
					{#if privateShare}
						<div class="flex gap-2">
							<input
								bind:value={passwordCode}
								maxlength="16"
								class="h-10 min-w-0 flex-1 rounded-lg border border-gray-200 px-3 text-sm uppercase outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
							/>
							<button type="button" onclick={refreshPasswordCode} class="inline-flex h-10 items-center gap-1.5 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 hover:bg-gray-50">
								<RefreshCw size={14} /> 换一个
							</button>
						</div>
					{/if}
				</section>

				<div class="flex justify-end gap-2 border-t border-gray-100 pt-4">
					<button type="button" onclick={() => (open = false)} class="h-9 rounded-lg px-4 text-sm text-gray-600 hover:bg-gray-100">取消</button>
					<button type="button" onclick={submitShare} disabled={creating} class="inline-flex h-9 items-center gap-2 rounded-lg bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
						{#if creating}<LoaderCircle size={15} class="animate-spin" />{/if}
						生成链接
					</button>
				</div>
			{:else}
				<section class="space-y-4">
					<div>
						<p class="mb-2 text-sm font-medium text-gray-800">分享链接</p>
						<div class="flex gap-2">
							<input readonly value={shareLink} class="h-10 min-w-0 flex-1 rounded-lg border border-gray-200 bg-gray-50 px-3 text-sm text-gray-700" />
							<button type="button" onclick={copyLink} class="inline-flex h-10 items-center gap-1.5 rounded-lg bg-blue-600 px-3 text-sm font-medium text-white hover:bg-blue-700">
								{#if copied}<Check size={14} />{:else}<Copy size={14} />{/if} 复制
							</button>
						</div>
					</div>
					{#if share.hasPassword}
						<div>
							<p class="mb-2 text-sm font-medium text-gray-800">提取码</p>
							<div class="flex gap-2">
								<input readonly value={passwordCode} class="h-10 w-32 rounded-lg border border-gray-200 bg-gray-50 px-3 text-sm tracking-widest text-gray-700" />
								<button type="button" onclick={copyPasswordCode} class="inline-flex h-10 items-center gap-1.5 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 hover:bg-gray-50">
									<Copy size={14} /> 复制提取码
								</button>
							</div>
						</div>
					{/if}
					<p class="text-xs text-gray-500">有效期：{share.expiresAt ? new Date(share.expiresAt).toLocaleString() : '永久'}</p>
				</section>
			{/if}
		</div>

		<aside class="flex min-h-52 items-center justify-center rounded-xl border border-dashed border-gray-200 bg-gray-50 p-4">
			{#if qrCodeUrl}
				<img src={qrCodeUrl} alt="分享二维码" class="h-44 w-44 rounded-lg bg-white p-2 shadow-sm" />
			{:else if shareLink}
				<div class="text-center text-gray-400">
					<QrCode size={44} class="mx-auto mb-3" />
					<p class="text-sm">链接较长，无法生成二维码</p>
				</div>
			{:else}
				<div class="text-center text-gray-400">
					<QrCode size={44} class="mx-auto mb-3" />
					<p class="text-sm">生成链接后显示二维码</p>
				</div>
			{/if}
		</aside>
	</div>
</Dialog>
