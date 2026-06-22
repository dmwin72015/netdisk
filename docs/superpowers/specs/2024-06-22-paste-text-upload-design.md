# 粘贴板文本自动转文件上传功能设计

## 需求概述

增强现有的粘贴上传功能，使其能够：
- 检测剪贴板中的**文本内容**
- **检查文本大小**，超过 2MB 则拒绝并提示用户
- 自动将文本转换为 `.txt` 文件（最大 2MB）
- 弹出对话框让用户**输入文件名**（可编辑，有默认值）
- 用户**确认后才上传**，给予充分的控制权

## 用户流程

```
用户按下 Ctrl+V
    ↓
检测剪贴板内容
    ↓
┌─────────────┬─────────────┐
│  是文件      │  是文本      │
│  (现有逻辑)  │  (新增逻辑)  │
└─────────────┴─────────────┘
              ↓
       检查文本大小
              ↓
        ┌────────────┐
        │ ≤ 2MB？    │
        └────┬────┬───┘
             │否  │是
        ┌────┘    └──────────┐
        ↓                    ↓
   显示错误提示          弹出对话框
   "文本内容不能          - 显示文件名输入框
   超过 2MB"            （默认：剪贴板内容.txt）
   返回                  - 显示文本内容预览
                         - 用户可编辑文件名
                               ↓
                        用户点击"确认上传"或"取消"
                               ↓
                        确认 → 创建 Blob → 转为 File → 走现有上传流程
                        取消 → 关闭对话框
```

## 技术方案

### 1. 新增工具函数：`paste-text-upload.ts`

**位置：** `frontend/src/lib/paste-text-upload.ts`

**常量：**
```typescript
/** 文本粘贴最大允许大小：2MB */
export const MAX_PASTE_TEXT_SIZE = 2 * 1024 * 1024;
```

**功能：**
- `extractClipboardText(clipboardData): string | null`
  从剪贴板提取纯文本内容，优先使用 `clipboardData.getData('text/plain')`

- `validateTextSize(text: string): { valid: boolean; size: number; error?: string }`
  检查文本大小，超过 2MB 返回错误信息

- `createTextFile(text: string, fileName: string): File`
  将文本内容转换为 `File` 对象，默认 MIME 类型为 `text/plain;charset=utf-8`

- `getDefaultFileName(text: string): string`
  生成默认文件名：
  - 提取文本第一行作为文件名基础
  - 移除不合法文件名字符
  - 限制长度（最多 100 字符）
  - 如果第一行太长或为空，使用时间戳：`clipboard-YYYY-MM-DD-HHmmss.txt`
  - 确保以 `.txt` 结尾

### 2. 新增组件：`PasteTextUploadConfirmDialog.svelte`

**位置：** `frontend/src/lib/components/files/PasteTextUploadConfirmDialog.svelte`

**Props：**
```typescript
{
  open: boolean;
  text: string;              // 剪贴板文本内容
  targetLabel: string;       // 目标位置（如"当前目录"）
  defaultFileName?: string;  // 默认文件名
  sizeError?: string;        // 大小错误信息（如果有）
  onConfirm: (file: File) => void | Promise<void>;
  onCancel: () => void;
}
```

**UI 结构：**
```
┌─────────────────────────────────────────┐
│  确认上传文本内容？          [X]         │
├─────────────────────────────────────────┤
│  📄 将上传文本文件到 "当前目录"         │
│                                         │
│  文件名：                               │
│  ┌─────────────────────────────────┐   │
│  │ 我的笔记.txt                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  内容预览（可折叠）：                   │
│  ┌─────────────────────────────────┐   │
│  │ 这是从剪贴板复制的文本内容...      │   │
│  │ （显示前 500 字符）              │   │
│  └─────────────────────────────────┘   │
│                                         │
│           [取消]    [确认上传]          │
└─────────────────────────────────────────┘
```

**特性：**
- 文件名输入框：默认填充建议文件名，用户可编辑
- 内容预览：默认折叠，可展开查看完整内容（最多显示 500 字符，超过显示省略）
- 如果文本内容过长（> 500 字符），预览区域显示字符计数
- 键盘支持：Enter 确认，Escape 取消
- **超过 2MB 时显示错误状态，禁用确认按钮并提示用户**

**错误状态 UI：**
```
┌─────────────────────────────────────────┐
│  确认上传文本内容？          [X]         │
├─────────────────────────────────────────┤
│  ⚠️  文本内容过大                        │
│  文本大小不能超过 2MB，                   │
│  当前文本大小为 2.5MB                    │
│                                         │
│           [取消]                        │
└─────────────────────────────────────────┘
```

### 3. 修改现有组件：`PasteUploadProvider.svelte`

**改动点：**

1. **引入新依赖：**
```typescript
import { extractClipboardText, createTextFile, getDefaultFileName } from '$lib/paste-text-upload';
import PasteTextUploadConfirmDialog from './PasteTextUploadConfirmDialog.svelte';
```

2. **新增状态：**
```typescript
let textDialogOpen = $state(false);
let clipboardText = $state<string | null>(null);
let textFileName = $state('');
```

3. **增强 `handlePaste` 逻辑：**
```typescript
function handlePaste(event: ClipboardEvent) {
  if (!enabled) return;

  // 优先检测文件
  const pastedFiles = extractClipboardFiles(event.clipboardData);
  if (pastedFiles.length > 0) {
    // 现有文件处理逻辑
    const result = filterPasteFiles(pastedFiles, acceptFile);
    acceptedFiles = result.accepted;
    rejectedFiles = result.rejected;
    if (acceptedFiles.length > 0) dialogOpen = true;
    return;
  }

  // 新增：检测文本
  const text = extractClipboardText(event.clipboardData);
  if (text && text.trim().length > 0) {
    // 检查文本大小
    const sizeCheck = validateTextSize(text);
    if (!sizeCheck.valid) {
      event.preventDefault();
      toast.error(sizeCheck.error || '文本内容过大');
      return;
    }

    event.preventDefault();
    clipboardText = text;
    textFileName = getDefaultFileName(text);
    textDialogOpen = true;
  }
}
```

4. **新增 `confirmTextUpload` 回调：**
```typescript
async function confirmTextUpload() {
  if (!clipboardText) return;

  textDialogOpen = false;
  const file = createTextFile(clipboardText, textFileName);
  await onUpload([file]);  // 复用现有的上传流程

  clipboardText = null;
  textFileName = '';
}
```

### 4. 文件命名策略

**默认文件名生成规则：**

```typescript
function getDefaultFileName(text: string): string {
  // 1. 提取第一行
  const firstLine = text.split(/\r?\n/)[0].trim();

  // 2. 如果第一行太长（> 50 字符）或为空，使用时间戳
  if (!firstLine || firstLine.length > 50) {
    const now = new Date();
    const timestamp = now.toISOString().replace(/[:.]/g, '-').slice(0, 19);
    return `clipboard-${timestamp}.txt`;
  }

  // 3. 清理不合法字符：\ / : * ? " < > |
  const cleaned = firstLine.replace(/[\\/:*?"<>|]/g, '_');

  // 4. 限制长度
  const truncated = cleaned.slice(0, 100);

  // 5. 确保 .txt 后缀
  return truncated.endsWith('.txt') ? truncated : `${truncated}.txt`;
}
```

**示例：**
| 剪贴板内容 | 默认文件名 |
|-----------|-----------|
| `Hello World` | `Hello World.txt` |
| `TODO: 完成文档\n- 步骤1\n- 步骤2` | `TODO: 完成文档.txt` |
| `Very long text... (超过50字符)` | `clipboard-2024-06-22T14-30-00.txt` |
| ``（空）`` | `clipboard-2024-06-22T14-30-00.txt` |

### 5. 集成到现有上传流程

文本文件创建后，通过 `onUpload([file])` 传递给 `upload-manager.svelte.ts`，**完全复用现有上传逻辑**，包括：
- 分片上传
- SHA-256 校验
- 秒传/去重
- 断点续传
- 上传队列管理

## 错误处理

| 场景 | 处理方式 |
|------|---------|
| 剪贴板为空文本 | 忽略，不弹窗 |
| **文本超过 2MB** | **弹出错误提示，告知用户"文本内容不能超过 2MB"** |
| 文件名输入为空 | 自动回退到默认文件名 |
| 文件名包含非法字符 | 自动替换为下划线 |
| 文件名过长 | 截断到 100 字符 |
| 文本编码问题 | 使用 `text/plain;charset=utf-8` |
| 取消上传 | 关闭对话框，不做任何操作 |

## 边界情况

1. **同时有文件和文本：** 优先处理文件（保持现有行为）
2. **纯富文本格式：** 使用 `text/plain` 提取纯文本，忽略格式
3. **剪贴板图片：** 走现有文件处理逻辑
4. **在可编辑元素中粘贴：** `isEditablePasteTarget` 已处理，不拦截
5. **文本超过 2MB：** 显示错误 toast 提示"文本内容不能超过 2MB"，不弹窗
6. **文本正好 2MB：** 正常弹窗，允许上传（≤ 2MB）

## UI 设计规范

遵循现有设计系统（Quiet Workshop）：

- **对话框：** 使用现有 `Dialog` 组件
- **图标：** Lucide 图标（`FileText` 或 `Clipboard`）
- **颜色：** 主色调使用 `primary` 系列
- **字体：** Inter，使用 Tailwind 类名
- **间距：** 4px 基准（`gap-3`、`px-5`、`py-4` 等）
- **圆角：** `rounded-xl`、`rounded-lg`
- **按钮：** 主按钮 `bg-primary`，次按钮 `border border-line bg-white`

## 测试覆盖

### `paste-text-upload.test.ts`

```typescript
describe('extractClipboardText', () => {
  it('提取纯文本内容');
  it('处理空剪贴板');
  it('忽略富文本格式，只返回纯文本');
});

describe('createTextFile', () => {
  it('正确创建 File 对象');
  it('设置正确的 MIME 类型');
  it('文件名和内容匹配');
});

describe('getDefaultFileName', () => {
  it('使用第一行作为文件名');
  it('移除非法字符');
  it('第一行过长时使用时间戳');
  it('空文本使用时间戳');
  it('确保 .txt 后缀');
});
```

### `PasteTextUploadConfirmDialog.test.ts`

```typescript
describe('PasteTextUploadConfirmDialog', () => {
  it('显示默认文件名');
  it('允许用户编辑文件名');
  it('点击确认触发 onConfirm');
  it('点击取消触发 onCancel');
  it('显示文本内容预览');
  it('长文本显示省略和字符计数');
});
```

## 实现步骤

1. ✅ 创建 `paste-text-upload.ts` 工具函数
2. ✅ 编写工具函数测试
3. ✅ 创建 `PasteTextUploadConfirmDialog.svelte` 组件
4. ✅ 编写组件测试（可选）
5. ✅ 修改 `PasteUploadProvider.svelte` 集成文本处理
6. ✅ 本地测试验证
7. ✅ 更新 `PasteUploadProvider.test.ts`（如果需要）

## 后续优化（可选）

- [ ] 支持其他文本格式（Markdown、HTML）自动选择扩展名
- [ ] 支持自定义默认文件名模板
- [ ] 支持设置最大文本长度限制
- [ ] 添加历史记录（最近使用的文件名）
- [ ] 支持批量文本粘贴（多个文本块）
