<script lang="ts">
	import { onDestroy } from 'svelte';
	import {
		Upload,
		FileVideo,
		Search,
		Loader2,
		CheckCircle2,
		XCircle,
		X,
		Copy,
		ExternalLink
	} from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';
	import { ApiError } from '$lib/api/client';
	import {
		initUpload,
		driveUpload,
		completeUpload,
		computeSHA256,
		computeFeatureHash,
		checkHash,
		claimHash,
		quickCheck,
		type UploadSession
	} from '$lib/api/uploads';

	type Phase = 'checking' | 'uploading' | 'queued' | 'processing' | 'completed' | 'failed';

	type UploadItem = {
		id: string;
		fileName: string;
		fileSize: number;
		phase: Phase;
		uploadPct: number;
		convertPct: number;
		taskId: string | null;
		m3u8Url: string | null;
		errorMsg: string | null;
		abortCtl: AbortController | null;
		es: EventSource | null;
		copied: boolean;
		passive: boolean; // true when item is a duplicate with no active tracking
	};

	let {
		onCompleted,
		resumeSession
	}: { onCompleted?: (taskId: string) => void; resumeSession?: UploadSession | null } = $props();

	let items = $state<UploadItem[]>([]);
	let dragging = $state(false);
	let input: HTMLInputElement | undefined = $state();

	const ALLOWED_EXTS = [
		'.mp4', '.m4v', '.mov', '.3gp', '.3g2',
		'.webm', '.mkv', '.avi', '.flv', '.wmv',
		'.asf', '.ogv', '.ogg', '.mpg', '.mpeg'
	];

	function fmtSize(size: number): string {
		if (size < 1024) return `${size} B`;
		if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
		if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
		return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`;
	}

	function validate(file: File): string | null {
		const lname = file.name.toLowerCase();
		if (!ALLOWED_EXTS.some((ext) => lname.endsWith(ext))) {
			return m.unsupported_format({ formats: ALLOWED_EXTS.join(' ') });
		}
		return null;
	}

	function removeItem(id: string) {
		const item = items.find((i) => i.id === id);
		if (item) {
			item.es?.close();
			item.abortCtl?.abort();
		}
		items = items.filter((i) => i.id !== id);
	}

	function dismissCompleted(id: string) {
		removeItem(id);
	}

	async function startUpload(idx: number, file: File) {
		try {
			// === Stage 1: Quick pre-filter (reads only 3 MB) ===
			const { featureHash, isFullHash } = await computeFeatureHash(file);
			const quickRes = await quickCheck(file.size, featureHash);

			if (!quickRes.found) {
				// No match at all — go straight to upload.
				// For small files, feature hash IS the full hash, pass as X-File-SHA256.
				const sha256ForUpload = isFullHash ? featureHash : undefined;
				return doUpload(idx, file, sha256ForUpload);
			}

			// === Stage 2: Full SHA-256 verification ===
			let fullSHA256: string;
			if (isFullHash) {
				// Small file: feature_hash IS the full hash
				fullSHA256 = featureHash;
			} else {
				// Large file: now compute full SHA-256
				fullSHA256 = await computeSHA256(file);
				if (fullSHA256 !== quickRes.sha256) {
					// Feature hash collision — not actually the same file. Upload normally.
					return doUpload(idx, file, fullSHA256);
				}
			}

			// === Stage 3: File exists — claim or reuse ===
			const checkRes = await checkHash(fullSHA256);
			if (checkRes.exists) {
				if (checkRes.deduped) {
					if (checkRes.transcoded) {
						const claimRes = await claimHash(fullSHA256, file.name, file.size);
						items[idx].uploadPct = 100;
						items[idx].convertPct = 100;
						items[idx].taskId = claimRes.task_id;
						items[idx].phase = 'completed';
						onCompleted?.(claimRes.task_id);
					} else {
						items[idx].errorMsg = m.file_in_use();
						items[idx].phase = 'failed';
					}
				} else {
					items[idx].uploadPct = 100;
					items[idx].taskId = checkRes.task_id!;
					if (checkRes.status === 'completed') {
						items[idx].convertPct = 100;
						items[idx].phase = 'completed';
						onCompleted?.(checkRes.task_id!);
					} else {
						items[idx].phase = checkRes.status === 'processing' ? 'processing' : 'queued';
						items[idx].passive = true;
						onCompleted?.(checkRes.task_id!);
					}
				}
				return;
			}

			// Stage 1 found a candidate but Stage 3 says no file exists — upload normally.
			return doUpload(idx, file, fullSHA256);
		} catch (e) {
			handleError(idx, e);
		}
	}

	async function doUpload(idx: number, file: File, sha256?: string) {
		items[idx].phase = 'uploading';
		const initResult = await initUpload(file.name, file.size, sha256);

		items[idx].abortCtl = new AbortController();
		const result = await driveUpload(initResult.id, file, 0, {
			signal: items[idx].abortCtl!.signal,
			onProgress: (p) => (items[idx].uploadPct = Math.round((p.uploaded / p.total) * 100))
		});
		items[idx].uploadPct = 100;
		items[idx].taskId = result.task_id;
		items[idx].phase = 'queued';
		connectSSE(items[idx], result.task_id);
	}

	async function resumeUpload(idx: number, file: File, session: UploadSession) {
		items[idx].uploadPct = Math.round((session.received_bytes / file.size) * 100);

		try {
			items[idx].abortCtl = new AbortController();
			let result;
			if (session.received_bytes >= file.size) {
				result = await completeUpload(session.id);
			} else {
				result = await driveUpload(session.id, file, session.received_bytes, {
					signal: items[idx].abortCtl!.signal,
					onProgress: (p) => (items[idx].uploadPct = Math.round((p.uploaded / p.total) * 100))
				});
			}
			items[idx].uploadPct = 100;
			items[idx].taskId = result.task_id;
			items[idx].phase = 'queued';
			connectSSE(items[idx], result.task_id);
		} catch (e) {
			handleError(idx, e);
		}
	}

	function handleError(idx: number, e: unknown) {
		if (e instanceof DOMException && e.name === 'AbortError') {
			items[idx].errorMsg = m.cancelled();
		} else {
			items[idx].errorMsg = e instanceof ApiError ? e.message : e instanceof Error ? e.message : m.upload_failed();
		}
		items[idx].phase = 'failed';
	}

	type Frame = {
		task_id: string;
		status: 'pending' | 'processing' | 'completed' | 'failed';
		progress: number;
		m3u8_url?: string;
		error?: string;
	};

	function connectSSE(item: UploadItem, id: string) {
		item.es?.close();
		const token = localStorage.getItem('vf.access') ?? '';
		const url = `/api/v1/tasks/${id}/progress${token ? `?access_token=${encodeURIComponent(token)}` : ''}`;
		item.es = new EventSource(url);

		const apply = (e: MessageEvent) => {
			try {
				const frame = JSON.parse(e.data) as Frame;
				item.convertPct = frame.progress;
				if (frame.status === 'processing') item.phase = 'processing';
				else if (frame.status === 'pending') item.phase = 'queued';
				else if (frame.status === 'completed') {
					item.phase = 'completed';
					item.m3u8Url = frame.m3u8_url ?? null;
				} else if (frame.status === 'failed') {
					item.phase = 'failed';
					item.errorMsg = frame.error ?? m.conversion_failed();
				}
			} catch {
				/* ignore */
			}
		};

		item.es.addEventListener('progress', apply);
		item.es.addEventListener('done', (e) => {
			apply(e as MessageEvent);
			item.es?.close();
			item.es = null;
			if (item.taskId) onCompleted?.(item.taskId);
		});
		item.es.onerror = () => {
			if (item.es && item.es.readyState === EventSource.CLOSED) item.es = null;
		};
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const f = e.dataTransfer?.files?.[0];
		if (!f) return;
		dispatch(f);
	}

	function onPick(e: Event) {
		const files = (e.target as HTMLInputElement).files;
		if (!files) return;
		for (const f of files) dispatch(f);
		(e.target as HTMLInputElement).value = '';
	}

	function dispatch(file: File) {
		// When in resume mode, pass directly to resume flow
		if (resumeSession) {
			const err = validate(file);
			if (err) {
				const item: UploadItem = {
					id: crypto.randomUUID(),
					fileName: file.name,
					fileSize: file.size,
					phase: 'failed',
					uploadPct: 0,
					convertPct: 0,
					taskId: null,
					m3u8Url: null,
					errorMsg: err,
					abortCtl: null,
					es: null,
					copied: false,
					passive: false
				};
				items = [...items, item];
				return;
			}
			if (file.size !== resumeSession.total_size) {
				const item: UploadItem = {
					id: crypto.randomUUID(),
					fileName: file.name,
					fileSize: file.size,
					phase: 'failed',
					uploadPct: 0,
					convertPct: 0,
					taskId: null,
					m3u8Url: null,
					errorMsg: m.file_size_mismatch({ fileSize: file.size, sessionSize: resumeSession.total_size }),
					abortCtl: null,
					es: null,
					copied: false,
					passive: false
				};
				items = [...items, item];
				return;
			}
			if (file.name !== resumeSession.filename) {
				const item: UploadItem = {
					id: crypto.randomUUID(),
					fileName: file.name,
					fileSize: file.size,
					phase: 'failed',
					uploadPct: 0,
					convertPct: 0,
					taskId: null,
					m3u8Url: null,
					errorMsg: m.file_name_mismatch({ fileName: file.name, sessionName: resumeSession.filename }),
					abortCtl: null,
					es: null,
					copied: false,
					passive: false
				};
				items = [...items, item];
				return;
			}
			const item = createItem(file, 'uploading', Math.round((resumeSession.received_bytes / file.size) * 100));
			items = [...items, item];
			void resumeUpload(items.length - 1, file, resumeSession);
			return;
		}

		const err = validate(file);
		if (err) {
			const item: UploadItem = {
				id: crypto.randomUUID(),
				fileName: file.name,
				fileSize: file.size,
				phase: 'failed',
				uploadPct: 0,
				convertPct: 0,
				taskId: null,
				m3u8Url: null,
				errorMsg: err,
				abortCtl: null,
				es: null,
				copied: false,
				passive: false
			};
			items = [...items, item];
			return;
		}

		const item = createItem(file, 'checking');
		items = [...items, item];
		void startUpload(items.length - 1, file);
	}

	function createItem(file: File, phase: Phase, uploadPct = 0, passive = false): UploadItem {
		return {
			id: crypto.randomUUID(),
			fileName: file.name,
			fileSize: file.size,
			phase,
			uploadPct,
			convertPct: 0,
			taskId: null,
			m3u8Url: null,
			errorMsg: null,
			abortCtl: null,
			es: null,
			copied: false,
			passive
		};
	}

	onDestroy(() => {
		for (const item of items) {
			item.es?.close();
			item.abortCtl?.abort();
		}
	});

	async function copyUrl(m3u8Url: string) {
		await navigator.clipboard.writeText(m3u8Url);
	}
</script>

<div class="space-y-3">
	<!-- Always-visible upload zone -->
	<div
		role="button"
		tabindex="0"
		class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {dragging
			? 'border-slate-900 bg-slate-100'
			: 'border-slate-300 bg-white'}"
		ondragover={(e) => {
			e.preventDefault();
			dragging = true;
		}}
		ondragleave={() => (dragging = false)}
		ondrop={onDrop}
		onclick={() => input?.click()}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') input?.click();
		}}
	>
		<div class="flex flex-col items-center gap-2 text-slate-600">
			<Upload size={28} />
			{#if resumeSession}
				<p class="text-sm">{m.resume_select({ filename: resumeSession.filename })}</p>
				<p class="text-xs text-slate-400">
					{m.transfer_progress({ received: fmtSize(resumeSession.received_bytes), total: fmtSize(resumeSession.total_size), pct: Math.round((resumeSession.received_bytes / resumeSession.total_size) * 100) })}
				</p>
			{:else}
				<p class="text-sm">{m.drag_here()}</p>
				<p class="text-xs text-slate-400">
					{m.supported_formats()}
				</p>
			{/if}
		</div>
		<input
			bind:this={input}
			type="file"
			accept="video/*,.mp4,.m4v,.mov,.3gp,.3g2,.webm,.mkv,.avi,.flv,.wmv,.asf,.ogv,.ogg,.mpg,.mpeg"
			class="hidden"
			multiple
			onchange={onPick}
		/>
	</div>

	<!-- Upload progress list -->
	{#if items.length > 0}
		<div class="space-y-2">
			{#each items as item (item.id)}
				<div class="rounded-lg border bg-white p-3">
					<div class="flex items-center gap-2 text-sm text-slate-800">
						<FileVideo size={16} class="shrink-0" />
						<span class="min-w-0 flex-1 truncate font-medium" title={item.fileName}>{item.fileName}</span>
						<span class="shrink-0 text-xs text-slate-400">{fmtSize(item.fileSize)}</span>
						<button
							type="button"
							onclick={() => removeItem(item.id)}
							class="shrink-0 text-slate-300 hover:text-red-600"
							aria-label={m.remove()}
						>
							<X size={14} />
						</button>
					</div>

					{#if item.phase === 'completed'}
						<div class="mt-2 flex items-center gap-2 text-sm text-emerald-700">
							<CheckCircle2 size={16} /> <span>{m.conversion_done()}</span>
						</div>
						{#if item.m3u8Url}
							<div class="mt-2 flex items-center gap-2">
								<code class="flex-1 truncate rounded bg-slate-100 px-2 py-1 text-xs">{item.m3u8Url}</code>
								<button
									type="button"
									onclick={() => copyUrl(item.m3u8Url!)}
									class="flex items-center gap-1 rounded border px-2 py-1 text-xs hover:bg-slate-50"
								>
									<Copy size={12} />
								</button>
							</div>
						{/if}
						<div class="mt-2 flex items-center gap-3 text-xs">
							{#if item.taskId}
								<a href="/videos/{item.taskId}" class="flex items-center gap-1 text-slate-700 underline hover:text-slate-900">
									<ExternalLink size={12} /> {m.play()}
								</a>
							{/if}
						</div>
					{:else if item.phase === 'failed'}
						<div class="mt-2 flex items-start gap-2 text-sm text-red-700">
							<XCircle size={16} class="mt-0.5 shrink-0" />
							<span class="break-all">{item.errorMsg ?? m.failed()}</span>
						</div>
					{:else if item.passive}
						<div class="mt-2 flex items-center gap-2 text-sm text-slate-600">
							<Loader2 size={14} class="animate-spin" />
							<span>
								{item.phase === 'queued' ? m.in_queue() : m.converting()}
							</span>
						</div>
						{#if item.taskId}
							<div class="mt-2 text-xs">
								<a href="/tasks/{item.taskId}" class="text-slate-600 underline hover:text-slate-900">{m.view_progress()}</a>
							</div>
						{/if}
					{:else}
						<div class="mt-2 flex items-center justify-between text-sm">
							<div class="flex items-center gap-2 text-slate-700">
								{#if item.phase === 'checking'}
									<Search size={14} />
								{:else}
									<Loader2 size={14} class="animate-spin" />
								{/if}
								<span>
									{item.phase === 'checking'
										? m.checking()
										: item.phase === 'uploading'
											? m.uploading_status()
											: item.phase === 'queued'
												? m.queued_status()
												: m.converting_status()}
								</span>
							</div>
							{#if item.phase !== 'checking'}
								<span class="text-xs text-slate-500">{item.phase === 'uploading' ? item.uploadPct : item.convertPct}%</span>
							{/if}
						</div>
						{#if item.phase !== 'checking'}
							<div class="mt-1.5 h-2 w-full overflow-hidden rounded bg-slate-200">
								<div
									class="h-full rounded bg-slate-900 transition-all"
									style="width:{item.phase === 'uploading' ? item.uploadPct : item.convertPct}%"
								></div>
							</div>
						{/if}
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
