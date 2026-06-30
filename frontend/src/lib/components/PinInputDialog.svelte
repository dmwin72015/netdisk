<script lang="ts">
	import { Dialog } from "$lib/ui/dialog";
	import { getPending, getPinValue, setPinValue, closePinDialog } from "$lib/dialog-pin.svelte";
	import * as m from "$lib/paraglide/messages";

	let pending = $derived(getPending());
	let pinVal = $derived(getPinValue());
	let inputs: HTMLInputElement[] = $state([]);
	let contentEl: HTMLElement | undefined = $state();
	let resolveValue: string | null | undefined = undefined;
	let open = $state(false);
	let closing = $state(false);

	$effect(() => {
		if (pending && !closing) {
			open = true;
		}
	});

	function onOpenAutoFocus(e: Event) {
		e.preventDefault();
		requestAnimationFrame(() => focusFirstEmpty());
	}

	function cleanup() {
		const p = pending;
		open = false;
		closing = false;
		closePinDialog();
		if (resolveValue !== undefined) {
			p?.resolve(resolveValue);
			resolveValue = undefined;
		}
	}

	function closeWithAnimation(value: string | null | undefined) {
		if (closing) return;
		closing = true;
		resolveValue = value;
		open = false;
		const el = contentEl;
		if (el) {
			el.addEventListener("animationend", cleanup, { once: true });
		}
		setTimeout(() => {
			if (closing) cleanup();
		}, 300);
	}

	function onOpenChange(isOpen: boolean) {
		if (!isOpen && !closing) {
			closeWithAnimation(null);
		}
	}

	function onConfirm() {
		const val = pinVal;
		if (val.length < 4) return;
		closeWithAnimation(val);
	}

	function onCancel() {
		closeWithAnimation(null);
	}

	function getBoxValue(index: number): string {
		return pinVal[index] ?? "";
	}

	function setBoxValue(index: number, value: string) {
		const digit = value.replace(/\D/g, "").slice(-1);
		const current = pinVal;
		const arr = current.split("");
		arr[index] = digit;
		const next = arr.join("");
		setPinValue(next);

		if (digit && index < 3) {
			inputs[index + 1]?.focus();
		}
	}

	function handleBoxKeydown(index: number, e: KeyboardEvent) {
		if (e.key === "Backspace") {
			if (pinVal[index]) {
				setBoxValue(index, "");
			} else if (index > 0) {
				inputs[index - 1]?.focus();
				setBoxValue(index - 1, "");
			}
		} else if (e.key === "ArrowLeft" && index > 0) {
			e.preventDefault();
			inputs[index - 1]?.focus();
		} else if (e.key === "ArrowRight" && index < 3) {
			e.preventDefault();
			inputs[index + 1]?.focus();
		} else if (e.key === "Enter") {
			e.preventDefault();
			onConfirm();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		const text = (e.clipboardData?.getData("text") ?? "").replace(/\D/g, "").slice(0, 4);
		if (text) {
			setPinValue(text);
			const nextIndex = Math.min(text.length, 3);
			inputs[nextIndex]?.focus();
		}
	}

	function handleContainerClick() {
		focusFirstEmpty();
	}

	function focusFirstEmpty() {
		for (let i = 0; i < 4; i++) {
			if (!pinVal[i]) {
				inputs[i]?.focus();
				return;
			}
		}
		inputs[3]?.focus();
	}

	function getConfirmDisabled(): boolean {
		return pinVal.length < 4;
	}
</script>

{#if pending}
	<Dialog
		{open}
		{onOpenChange}
		{onCancel}
		title={pending?.opts.title ?? ""}
		description={pending?.opts.message}
		showConfirm={false}
		showCancel={false}
		{onOpenAutoFocus}
	>
		{#snippet children()}
			<div bind:this={contentEl}>
				<div
					class="flex items-center justify-center gap-3"
					onclick={handleContainerClick}
					role="group"
					aria-label="PIN input"
				>
					{#each { length: 4 } as _, i}
						<input
							bind:this={inputs[i]}
							type="password"
							inputmode="numeric"
							pattern="[0-9]"
							maxlength="1"
							value={getBoxValue(i)}
							oninput={(e) => setBoxValue(i, e.currentTarget.value)}
							onkeydown={(e) => handleBoxKeydown(i, e)}
							onpaste={handlePaste}
							onfocus={() => {
								if (inputs[i] && pinVal[i]) {
									inputs[i].select();
								}
							}}
							class="h-12 w-10 rounded-lg border border-line bg-surface text-center text-lg font-mono font-medium text-ink outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20"
							aria-label="Digit {i + 1}"
						/>
					{/each}
				</div>

				<div class="mt-5 flex justify-end gap-2">
					<button
						type="button"
						onclick={onCancel}
						class="rounded-lg px-4 py-2 text-sm text-ink-3 transition-colors duration-150 hover:bg-surface-sunken"
					>
						{pending?.opts.cancelText ?? m.cancel()}
					</button>
					<button
						type="button"
						onclick={onConfirm}
						disabled={getConfirmDisabled()}
						class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors duration-150 hover:bg-primary-hover disabled:opacity-40 disabled:hover:bg-primary"
					>
						{pending?.opts.confirmText ?? m.confirm()}
					</button>
				</div>
			</div>
		{/snippet}
	</Dialog>
{/if}
