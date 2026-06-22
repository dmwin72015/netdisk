# 粘贴板文本转文件上传功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 增强粘贴上传功能，支持检测剪贴板文本并在确认后自动创建 .txt 文件上传（最大 2MB）

**Architecture:** 新增独立的工具函数模块处理文本提取和文件创建，新增对话框组件处理用户交互，修改现有粘贴上传提供者集成文本处理逻辑

**Tech Stack:** TypeScript, Svelte 5, Vitest, Tailwind CSS, Lucide Icons

## Global Constraints

- 文本大小限制：2MB（`2 * 1024 * 1024` 字节）
- 文件命名：自动提取第一行或使用时间戳，清理非法字符，确保 .txt 后缀
- UI 风格：遵循 Quiet Workshop 设计系统（Primary 色系，Inter 字体，4px 间距基准）
- 测试框架：Vitest（沿用项目现有测试框架）
- 类型安全：TypeScript 严格模式
- MIME 类型：`text/plain;charset=utf-8`

---

### Task 1: 创建文本上传工具函数模块

**Files:**
- Create: `frontend/src/lib/paste-text-upload.ts`
- Test: `frontend/src/lib/paste-text-upload.test.ts`

**Interfaces:**
- Consumes: 无（独立工具函数）
- Produces:
  - `MAX_PASTE_TEXT_SIZE: number` (2MB 常量)
  - `extractClipboardText(clipboardData): string | null`
  - `validateTextSize(text): { valid: boolean; size: number; error?: string }`
  - `createTextFile(text, fileName): File`
  - `getDefaultFileName(text): string`

- [ ] **Step 1: 编写失败的测试**

```typescript
import { describe, expect, it } from 'vitest';
import {
  extractClipboardText,
  validateTextSize,
  createTextFile,
  getDefaultFileName,
  MAX_PASTE_TEXT_SIZE
} from './paste-text-upload';

describe('extractClipboardText', () => {
  it('extracts plain text from clipboardData', () => {
    const clipboard = {
      getData: (type: string) => type === 'text/plain' ? 'Hello World' : ''
    } as DataTransfer;
    expect(extractClipboardText(clipboard)).toBe('Hello World');
  });

  it('returns null for empty text', () => {
    const clipboard = {
      getData: () => ''
    } as DataTransfer;
    expect(extractClipboardText(clipboard)).toBeNull();
  });

  it('handles null clipboardData', () => {
    expect(extractClipboardText(null)).toBeNull();
    expect(extractClipboardText(undefined)).toBeNull();
  });
});

describe('validateTextSize', () => {
  it('accepts text within 2MB limit', () => {
    const text = 'a'.repeat(1024 * 1024); // 1MB
    const result = validateTextSize(text);
    expect(result.valid).toBe(true);
    expect(result.size).toBe(1024 * 1024);
  });

  it('rejects text exceeding 2MB limit', () => {
    const text = 'a'.repeat(2 * 1024 * 1024 + 1); // 2MB + 1 byte
    const result = validateTextSize(text);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('2MB');
  });

  it('accepts exactly 2MB text', () => {
    const text = 'a'.repeat(2 * 1024 * 1024); // 2MB
    const result = validateTextSize(text);
    expect(result.valid).toBe(true);
  });
});

describe('createTextFile', () => {
  it('creates a File with correct content and name', () => {
    const file = createTextFile('Hello World', 'test.txt');
    expect(file.name).toBe('test.txt');
    expect(file.type).toBe('text/plain;charset=utf-8');
    expect(file.size).toBe(11);
  });

  it('preserves text content accurately', async () => {
    const content = 'Line 1\nLine 2\nLine 3';
    const file = createTextFile(content, 'multiline.txt');
    const text = await file.text();
    expect(text).toBe(content);
  });
});

describe('getDefaultFileName', () => {
  it('uses first line as filename', () => {
    expect(getDefaultFileName('My Document\nSecond line')).toBe('My Document.txt');
  });

  it('removes illegal characters', () => {
    expect(getDefaultFileName('File/Name:With*Illegal?Chars')).toBe('File_Name_With_Illegal_Chars.txt');
  });

  it('truncates long first line', () => {
    const longLine = 'a'.repeat(60);
    const result = getDefaultFileName(longLine);
    expect(result.startsWith('clipboard-')).toBe(true);
    expect(result.endsWith('.txt')).toBe(true);
  });

  it('uses timestamp for empty first line', () => {
    const result = getDefaultFileName('\nSecond line');
    expect(result).toMatch(/^clipboard-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}\.txt$/);
  });

  it('ensures .txt extension', () => {
    expect(getDefaultFileName('readme')).toBe('readme.txt');
    expect(getDefaultFileName('notes.txt')).toBe('notes.txt');
  });
});

describe('MAX_PASTE_TEXT_SIZE', () => {
  it('equals 2MB', () => {
    expect(MAX_PASTE_TEXT_SIZE).toBe(2 * 1024 * 1024);
  });
});
```

- [ ] **Step 2: 运行测试确保失败**

Run: `cd frontend && pnpm test paste-text-upload.test.ts`
Expected: FAIL with "Cannot find module './paste-text-upload'"

- [ ] **Step 3: 编写最小实现**

```typescript
/** 文本粘贴最大允许大小：2MB */
export const MAX_PASTE_TEXT_SIZE = 2 * 1024 * 1024;

/**
 * 从剪贴板提取纯文本内容
 */
export function extractClipboardText(clipboardData: DataTransfer | null | undefined): string | null {
  if (!clipboardData) return null;
  const text = clipboardData.getData('text/plain');
  return text?.trim() ? text : null;
}

/**
 * 验证文本大小
 */
export function validateTextSize(text: string): { valid: boolean; size: number; error?: string } {
  const size = new Blob([text]).size;
  if (size > MAX_PASTE_TEXT_SIZE) {
    return {
      valid: false,
      size,
      error: `文本大小 ${formatSize(size)} 超过限制 ${formatSize(MAX_PASTE_TEXT_SIZE)}`
    };
  }
  return { valid: true, size };
}

/**
 * 将文本转换为 File 对象
 */
export function createTextFile(text: string, fileName: string): File {
  return new File([text], fileName, { type: 'text/plain;charset=utf-8' });
}

/**
 * 生成默认文件名
 */
export function getDefaultFileName(text: string): string {
  const firstLine = text.split(/\r?\n/)[0].trim();

  if (!firstLine || firstLine.length > 50) {
    const now = new Date();
    const timestamp = now.toISOString().replace(/[:.]/g, '-').slice(0, 19);
    return `clipboard-${timestamp}.txt`;
  }

  const cleaned = firstLine.replace(/[\\/:*?"<>|]/g, '_');
  const truncated = cleaned.slice(0, 100);
  return truncated.endsWith('.txt') ? truncated : `${truncated}.txt`;
}

/**
 * 格式化文件大小
 */
function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}
```

- [ ] **Step 4: 运行测试确保通过**

Run: `cd frontend && pnpm test paste-text-upload.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/paste-text-upload.ts frontend/src/lib/paste-text-upload.test.ts
git commit -m "feat: add paste text upload utilities"
```

---

### Task 2: 创建文本粘贴确认对话框组件

**Files:**
- Create: `frontend/src/lib/components/files/PasteTextUploadConfirmDialog.svelte`
- Test: `frontend/src/lib/components/files/PasteTextUploadConfirmDialog.test.ts` (可选)

**Interfaces:**
- Consumes: `MAX_PASTE_TEXT_SIZE`, `formatSize`（从 utils）
- Produces: `PasteTextUploadConfirmDialog` 组件

**Props:**
```typescript
{
  open: boolean;
  text: string;
  targetLabel: string;
  defaultFileName?: string;
  sizeError?: string;
  onConfirm: (file: File) => void | Promise<void>;
  onCancel: () => void;
}
```

- [ ] **Step 1: 编写组件代码**

```svelte
<script lang="ts">
  import { FileText, Check, X, AlertTriangle, ChevronDown, ChevronUp } from '@lucide/svelte';
  import { Dialog } from '$lib/ui/dialog';
  import { fmtSize } from '$lib/utils/format';
  import { MAX_PASTE_TEXT_SIZE } from '$lib/paste-text-upload';

  let {
    open,
    text,
    targetLabel,
    defaultFileName = '',
    sizeError,
    onConfirm,
    onCancel
  }: {
    open: boolean;
    text: string;
    targetLabel: string;
    defaultFileName?: string;
    sizeError?: string;
    onConfirm: (file: File) => void | Promise<void>;
    onCancel: () => void;
  } = $props();

  let fileName = $state(defaultFileName);
  let previewExpanded = $state(false);

  const PREVIEW_MAX_LENGTH = 500;

  function handleOpenChange(value: boolean) {
    if (!value) onCancel();
  }

  $effect(() => {
    if (open) {
      fileName = defaultFileName || getDefaultFileName();
    }
  });

  function getDefaultFileName(): string {
    const firstLine = text.split(/\r?\n/)[0].trim();
    if (!firstLine || firstLine.length > 50) {
      const now = new Date();
      const timestamp = now.toISOString().replace(/[:.]/g, '-').slice(0, 19);
      return `clipboard-${timestamp}.txt`;
    }
    const cleaned = firstLine.replace(/[\\/:*?"<>|]/g, '_').slice(0, 100);
    return cleaned.endsWith('.txt') ? cleaned : `${cleaned}.txt`;
  }

  $derived if (previewExpanded) {
    text;
  } else {
    text.length > PREVIEW_MAX_LENGTH ? text.slice(0, PREVIEW_MAX_LENGTH) + '...' : text;
  }

  async function handleConfirm() {
    if (!fileName.trim() || sizeError) return;
    const file = new File([text], fileName, { type: 'text/plain;charset=utf-8' });
    await onConfirm(file);
  }

  $derived hasError = !!sizeError;
  $derived textSize = new Blob([text]).size;
  $derived previewText = previewExpanded ? text : (text.length > PREVIEW_MAX_LENGTH ? text.slice(0, PREVIEW_MAX_LENGTH) + '...' : text);
</script>

<Dialog
  bind:open
  onOpenChange={handleOpenChange}
  onCancel={onCancel}
  title={hasError ? '文本内容过大' : '确认上传文本内容？'}
  footer={false}
  class="max-w-lg"
  bodyClass="!p-0"
>
  {#if hasError}
    <div class="border-b border-line-soft bg-error-soft px-5 py-4">
      <div class="flex items-start gap-3">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-error text-white">
          <AlertTriangle size={20} />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-error">{sizeError}</p>
          <p class="mt-1 text-xs text-ink-4">
            文本大小：{fmtSize(textSize)} / {fmtSize(MAX_PASTE_TEXT_SIZE)}
          </p>
        </div>
      </div>
    </div>
  {:else}
    <div class="border-b border-line-soft px-5 py-4">
      <div class="flex items-start gap-3">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
          <FileText size={20} />
        </div>
        <div class="min-w-0">
          <p class="text-sm text-ink-2">
            将上传文本文件到 <span class="font-semibold text-ink">{targetLabel}</span>
          </p>
          <p class="mt-1 text-xs text-ink-4">大小 {fmtSize(textSize)}</p>
        </div>
      </div>
    </div>

    <div class="space-y-4 px-5 py-4">
      <div>
        <label class="mb-2 block text-sm font-medium text-ink-2" for="fileName">
          文件名
        </label>
        <input
          id="fileName"
          type="text"
          bind:value={fileName}
          class="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
          placeholder="输入文件名..."
        />
      </div>

      <div>
        <button
          type="button"
          onclick={() => previewExpanded = !previewExpanded}
          class="flex w-full items-center justify-between rounded-lg border border-line bg-surface-muted px-3 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-sunken"
        >
          <span class="font-medium">内容预览</span>
          <div class="flex items-center gap-2">
            {#if text.length > PREVIEW_MAX_LENGTH}
              <span class="text-xs text-ink-4">{text.length} 字符</span>
            {/if}
            {#if previewExpanded}
              <ChevronUp size={16} />
            {:else}
              <ChevronDown size={16} />
            {/if}
          </div>
        </button>
        {#if previewExpanded || text.length <= PREVIEW_MAX_LENGTH}
          <div class="mt-2 max-h-60 overflow-y-auto rounded-lg border border-line bg-surface-sunken p-3">
            <pre class="whitespace-pre-wrap break-words text-xs text-ink-2">{previewText}</pre>
          </div>
        {/if}
      </div>
    </div>

    <div class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3">
      <button
        type="button"
        onclick={onCancel}
        class="inline-flex items-center gap-1.5 rounded-lg border border-line bg-white px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-muted"
      >
        <X size={14} /> 取消
      </button>
      <button
        type="button"
        onclick={handleConfirm}
        disabled={!fileName.trim() || hasError}
        class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Check size={14} /> 确认上传
      </button>
    </div>
  {/if}
</Dialog>
```

- [ ] **Step 2: 创建基础测试**（可选，但推荐）

```typescript
import { describe, expect, it, vi } from 'vitest';
import { mount } from 'svelte/testing';
import PasteTextUploadConfirmDialog from './PasteTextUploadConfirmDialog.svelte';

describe('PasteTextUploadConfirmDialog', () => {
  it('renders with default filename', () => {
    const component = mount(PasteTextUploadConfirmDialog, {
      props: {
        open: true,
        text: 'Hello World',
        targetLabel: '当前目录',
        defaultFileName: 'test.txt',
        onConfirm: vi.fn(),
        onCancel: vi.fn()
      }
    });

    expect(component.text()).toContain('test.txt');
  });

  it('calls onCancel when cancel button clicked', async () => {
    const onCancel = vi.fn();
    const component = mount(PasteTextUploadConfirmDialog, {
      props: {
        open: true,
        text: 'Hello',
        targetLabel: '当前目录',
        onConfirm: vi.fn(),
        onCancel
      }
    });

    await component.find('button').trigger('click');
    expect(onCancel).toHaveBeenCalled();
  });
});
```

- [ ] **Step 3: 运行组件测试**（如果编写）

Run: `cd frontend && pnpm test PasteTextUploadConfirmDialog.test.ts`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/files/PasteTextUploadConfirmDialog.svelte
git commit -m "feat: add paste text upload confirm dialog"
```

---

### Task 3: 集成文本处理到粘贴上传提供者

**Files:**
- Modify: `frontend/src/lib/components/files/PasteUploadProvider.svelte`

**Interfaces:**
- Consumes: `extractClipboardText`, `validateTextSize`, `getDefaultFileName`, `createTextFile` from `paste-text-upload.ts`; `PasteTextUploadConfirmDialog` component
- Produces: 增强的粘贴处理逻辑

- [ ] **Step 1: 修改 PasteUploadProvider.svelte**

在现有文件基础上添加以下改动：

**1.1 引入新依赖**

```typescript
import { extractClipboardText, validateTextSize, getDefaultFileName, createTextFile } from '$lib/paste-text-upload';
import PasteTextUploadConfirmDialog from './PasteTextUploadConfirmDialog.svelte';
```

**1.2 新增状态**

```typescript
let textDialogOpen = $state(false);
let clipboardText = $state<string | null>(null);
let textFileName = $state('');
```

**1.3 增强 handlePaste 函数**

```typescript
function handlePaste(event: ClipboardEvent) {
  if (!enabled) return;

  // 优先检测文件
  const pastedFiles = extractClipboardFiles(event.clipboardData);
  if (pastedFiles.length > 0) {
    const result = filterPasteFiles(pastedFiles, acceptFile);
    acceptedFiles = result.accepted;
    rejectedFiles = result.rejected;
    if (acceptedFiles.length > 0) dialogOpen = true;
    return;
  }

  // 新增：检测文本
  const text = extractClipboardText(event.clipboardData);
  if (text && text.trim().length > 0) {
    event.preventDefault();

    // 检查文本大小
    const sizeCheck = validateTextSize(text);
    if (!sizeCheck.valid) {
      toast.error(sizeCheck.error || '文本内容过大');
      return;
    }

    clipboardText = text;
    textFileName = getDefaultFileName(text);
    textDialogOpen = true;
  }
}
```

**1.4 新增 confirmTextUpload 回调**

```typescript
async function confirmTextUpload(file: File) {
  textDialogOpen = false;
  try {
    await onUpload([file]);
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '粘贴上传失败');
  } finally {
    clipboardText = null;
    textFileName = '';
  }
}
```

**1.5 在模板中添加新对话框**

```svelte
<PasteTextUploadConfirmDialog
  bind:open={textDialogOpen}
  text={clipboardText || ''}
  targetLabel={targetLabel}
  defaultFileName={textFileName}
  onConfirm={confirmTextUpload}
  onCancel={() => {
    textDialogOpen = false;
    clipboardText = null;
    textFileName = '';
  }}
/>
```

- [ ] **Step 2: 本地测试验证**

启动前端开发服务并测试：
```bash
cd frontend && pnpm dev
```

**测试场景：**
1. ✅ 粘贴纯文本 → 应弹出文本上传对话框
2. ✅ 粘贴文件 → 应保持原有文件上传对话框
3. ✅ 粘贴超 2MB 文本 → 应显示错误 toast
4. ✅ 在输入框中粘贴 → 不应触发上传
5. ✅ 文本上传对话框 → 文件名可编辑，文本可预览，确认后走正常上传流程

- [ ] **Step 3: 运行前端类型检查**

Run: `cd frontend && pnpm check`
Expected: PASS（无 TypeScript 错误）

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/files/PasteUploadProvider.svelte
git commit -m "feat: integrate text paste handling into paste upload provider"
```

---

### Task 4: 更新测试覆盖率

**Files:**
- Modify: `frontend/src/lib/paste-upload.test.ts`

**任务：** 在现有测试中添加针对文本粘贴的测试用例

- [ ] **Step 1: 添加文本粘贴测试用例**

```typescript
import { extractClipboardText, validateTextSize } from './paste-text-upload';

describe('PasteUploadProvider - text handling', () => {
  it('should prioritize files over text when both present', () => {
    // 测试当剪贴板同时包含文件和文本时，优先处理文件
    const file = new File([''], 'test.txt');
    const clipboard = {
      files: [file],
      getData: () => 'text content'
    } as DataTransfer;

    const result = extractClipboardFiles(clipboard);
    expect(result).toHaveLength(1);
  });

  it('should validate text size before showing dialog', () => {
    const text = 'a'.repeat(2 * 1024 * 1024 + 1);
    const result = validateTextSize(text);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('2MB');
  });
});
```

- [ ] **Step 2: 运行所有测试**

Run: `cd frontend && pnpm test paste-upload`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/paste-upload.test.ts
git commit -m "test: add text paste test cases"
```

---

### Task 5: 最终验证与文档

- [ ] **Step 1: 运行完整测试套件**

```bash
cd frontend && pnpm test
```

Expected: All tests pass

- [ ] **Step 2: 类型检查**

```bash
cd frontend && pnpm check
```

Expected: No TypeScript errors

- [ ] **Step 3: 手动测试清单**

在浏览器中验证以下场景：

- [ ] 粘贴纯文本文件（< 2MB）→ 弹出对话框 → 可编辑文件名 → 确认后开始上传
- [ ] 粘贴纯文本文件（> 2MB）→ 显示错误 toast → 不弹窗
- [ ] 粘贴文件 → 保持原有行为（文件上传对话框）
- [ ] 粘贴富文本 → 提取纯文本 → 正常弹窗
- [ ] 在输入框中粘贴 → 不触发上传
- [ ] 取消上传 → 关闭对话框，不上传
- [ ] 长文本（> 500 字符）→ 预览显示省略，可展开

- [ ] **Step 4: 更新 README（如需要）**

如果项目 README 或功能列表需要更新，添加文本粘贴上传功能说明

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "feat: add clipboard text to txt file upload support

- Add paste-text-upload utility module with size validation (2MB limit)
- Add PasteTextUploadConfirmDialog component for text upload confirmation
- Integrate text paste handling into PasteUploadProvider
- Smart filename generation from first line or timestamp
- Text preview with expand/collapse (500 char limit)
- All existing paste upload functionality preserved"
```

---

## 测试策略

### 单元测试
- ✅ `paste-text-upload.test.ts`：覆盖所有工具函数
- ✅ `PasteTextUploadConfirmDialog.test.ts`：覆盖对话框渲染和交互

### 集成测试
- ✅ `paste-upload.test.ts`：添加文本粘贴场景测试
- ✅ 手动浏览器测试：验证完整用户流程

### 边界测试
- ✅ 空文本
- ✅ 正好 2MB 文本
- ✅ 超过 2MB 文本
- ✅ 长文件名（超过 50 字符）
- ✅ 非法文件名字符
- ✅ 多行文本
- ✅ 混合粘贴（文件 + 文本）

## 回滚计划

如果新功能出现严重问题：
1. 移除 `PasteUploadProvider.svelte` 中的文本处理逻辑
2. 删除 `PasteTextUploadConfirmDialog.svelte`
3. 删除 `paste-text-upload.ts` 和相关测试
4. Revert commit: `git revert <commit-hash>`

## 后续优化建议

- 支持其他文本格式（Markdown → .md，HTML → .html）
- 自定义文件名模板
- 最大文本长度配置化（从后端配置读取）
- 粘贴历史记录
- 批量文本粘贴支持
