<script lang="ts">
	import {
		getShowSystemDirs, setShowSystemDirs,
		getUploadConcurrency, setUploadConcurrency,
		getDuplicateStrategy, setDuplicateStrategy,
	} from '$lib/stores/file-preferences.svelte';
	import { UPLOAD_FILE_CONCURRENCY } from '$lib/upload-concurrency';
	import { Eye, Upload, FileWarning } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';
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
