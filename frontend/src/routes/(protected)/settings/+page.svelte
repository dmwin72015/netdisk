<script lang="ts">
	import {
		getShowSystemDirs, setShowSystemDirs,
		getUploadConcurrency, setUploadConcurrency,
		getDuplicateStrategy, setDuplicateStrategy,
		getDirectoryUnlockTtlHours, setDirectoryUnlockTtlHours,
		exportPreferences, importPreferences,
	} from '$lib/stores/file-preferences.svelte';
	import { UPLOAD_FILE_CONCURRENCY } from '$lib/upload-concurrency';
	import { Eye, Upload, FileWarning, Lock, FileJson } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';

	let importInput: HTMLInputElement | undefined = $state();
	let importing = $state(false);

	const ttlOptions = [
		{ value: 1, label: '1 小时' },
		{ value: 2, label: '2 小时' },
		{ value: 6, label: '6 小时' },
		{ value: 24, label: '24 小时' },
		{ value: -1, label: '永久（当前登录状态）' },
	];

	function downloadSettingsJson() {
		const blob = new Blob([JSON.stringify(exportPreferences(), null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = 'netdisk-settings.json';
		link.click();
		URL.revokeObjectURL(url);
	}

	async function onImportSettings(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		importing = true;
		try {
			const raw = await file.text();
			await importPreferences(JSON.parse(raw));
			toast.success('设置已恢复');
		} catch {
			toast.error('恢复设置失败，请选择有效的 JSON 文件');
		} finally {
			importing = false;
		}
	}
</script>

<div class="mx-auto max-w-2xl space-y-8">
	<h1 class="text-xl font-semibold text-gray-900">{m.file_settings()}</h1>

	<!-- Display -->
	<section>
		<div class="mb-3 flex items-center gap-2">
			<Eye size={16} class="text-gray-400" />
			<h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">{m.display_settings()}</h2>
		</div>
		<div class="rounded-xl border border-gray-100 bg-white shadow-sm">
			<div class="flex items-center justify-between gap-4 px-5 py-4">
				<div class="min-w-0">
					<p class="text-sm font-medium text-gray-800">{m.show_system_dirs()}</p>
					<p class="mt-1 text-xs leading-5 text-gray-500">{m.show_system_dirs_desc()}</p>
				</div>
				<input
					type="checkbox"
					checked={getShowSystemDirs()}
					onchange={(e) => setShowSystemDirs((e.currentTarget as HTMLInputElement).checked)}
					class="h-4 w-4 shrink-0 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
				/>
			</div>
		</div>
	</section>

	<!-- Directory Lock -->
	<section>
		<div class="mb-3 flex items-center gap-2">
			<Lock size={16} class="text-gray-400" />
			<h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">目录加锁</h2>
		</div>
		<div class="rounded-xl border border-gray-100 bg-white shadow-sm">
			<div class="flex items-center justify-between gap-4 px-5 py-4">
				<div class="min-w-0">
					<p class="text-sm font-medium text-gray-800">密码有效期</p>
					<p class="mt-1 text-xs leading-5 text-gray-500">目录解锁后，在当前登录会话中的有效时间。</p>
				</div>
				<select
					value={getDirectoryUnlockTtlHours()}
					onchange={(e) => setDirectoryUnlockTtlHours(parseInt((e.currentTarget as HTMLSelectElement).value, 10))}
					class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
				>
					{#each ttlOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>
		</div>
	</section>

	<!-- Backup -->
	<section>
		<div class="mb-3 flex items-center gap-2">
			<FileJson size={16} class="text-gray-400" />
			<h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">配置备份</h2>
		</div>
		<div class="rounded-xl border border-gray-100 bg-white shadow-sm">
			<div class="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
				<div class="min-w-0">
					<p class="text-sm font-medium text-gray-800">导出 / 恢复设置</p>
					<p class="mt-1 text-xs leading-5 text-gray-500">将当前设置导出为 JSON，或从 JSON 文件恢复。</p>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					<button
						type="button"
						onclick={downloadSettingsJson}
						class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
					>
						导出 JSON
					</button>
					<button
						type="button"
						disabled={importing}
						onclick={() => importInput?.click()}
						class="rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
					>
						{importing ? '恢复中...' : '从 JSON 恢复'}
					</button>
					<input bind:this={importInput} type="file" accept="application/json,.json" class="hidden" onchange={onImportSettings} />
				</div>
			</div>
		</div>
	</section>

	<!-- Upload -->
	<section>
		<div class="mb-3 flex items-center gap-2">
			<Upload size={16} class="text-gray-400" />
			<h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">{m.upload_title()}</h2>
		</div>
		<div class="rounded-xl border border-gray-100 bg-white shadow-sm">
			<div class="flex items-center justify-between gap-4 px-5 py-4">
				<div class="min-w-0">
					<p class="text-sm font-medium text-gray-800">{m.upload_concurrency()}</p>
					<p class="mt-1 text-xs leading-5 text-gray-500">{m.upload_concurrency_desc()}</p>
				</div>
				<span class="flex items-center gap-3">
					<input
						type="range"
						min="1"
						max={UPLOAD_FILE_CONCURRENCY}
						step="1"
						value={getUploadConcurrency()}
						oninput={(e) => setUploadConcurrency(parseInt((e.currentTarget as HTMLInputElement).value, 10))}
						class="h-1.5 w-24 appearance-none rounded-full bg-gray-200 accent-blue-600"
					/>
					<span class="w-6 text-center text-sm font-medium text-gray-700">{getUploadConcurrency()}</span>
				</span>
			</div>
		</div>
	</section>

	<!-- Duplicates -->
	<section>
		<div class="mb-3 flex items-center gap-2">
			<FileWarning size={16} class="text-gray-400" />
			<h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">{m.duplicate_strategy()}</h2>
		</div>
		<div class="rounded-xl border border-gray-100 bg-white shadow-sm">
			<div class="border-b border-gray-100 px-5 py-4">
				<p class="text-sm leading-5 text-gray-500">{m.duplicate_strategy_desc()}</p>
			</div>
			<div class="space-y-1 px-2 py-2">
				{#each [
					['prompt', m.duplicate_strategy_prompt()],
					['overwrite', m.duplicate_strategy_overwrite()],
					['keep_both', m.duplicate_strategy_keep_both()],
					['skip', m.duplicate_strategy_skip()],
				] as [value, label]}
					<button
						type="button"
						onclick={() => setDuplicateStrategy(value)}
						class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors {getDuplicateStrategy() === value ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-50'}"
					>
						<span class="flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2 {getDuplicateStrategy() === value ? 'border-blue-600' : 'border-gray-300'}">
							{#if getDuplicateStrategy() === value}
								<span class="h-2 w-2 rounded-full bg-blue-600"></span>
							{/if}
						</span>
						<span class="text-sm">{label}</span>
					</button>
				{/each}
			</div>
		</div>
	</section>
</div>
