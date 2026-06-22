---
name: NetDisk
description: 现代、克制、舒适的个人私有云网盘
colors:
  # Primary — 强调色，按事件出现（选中 / 进行中 / 主操作）
  primary: "#2563eb"            # oklch(57% 0.18 252) — 比 Tailwind blue-600 略克制
  primary-hover: "#1d4ed8"      # oklch(50% 0.20 252)
  primary-active: "#1e40af"     # oklch(44% 0.20 252)
  primary-soft: "#eff6ff"       # oklch(97% 0.02 252) — 选中态 / hover 软底
  primary-on: "#ffffff"

  # Neutrals — Cool gray，微微偏 primary 色相（~252°），不是 Tailwind 默认中性灰
  ink: "#0f172a"                # oklch(20% 0.02 252) — 主文本
  ink-2: "#334155"              # oklch(38% 0.02 252) — 次要文本
  ink-3: "#64748b"              # oklch(52% 0.02 252) — 三级 / 辅助文本
  ink-4: "#94a3b8"              # oklch(67% 0.02 252) — 占位 / 弱化图标
  ink-5: "#cbd5e1"              # oklch(82% 0.02 252) — 禁用文本

  line: "#e2e8f0"               # oklch(89% 0.01 252) — 主分隔线
  line-soft: "#f1f5f9"          # oklch(95% 0.005 252) — 极弱分隔
  surface: "#ffffff"            # 内容卡片 / 列表行
  surface-muted: "#f8fafc"      # oklch(98% 0.005 252) — body / 侧边面板
  surface-sunken: "#f1f5f9"     # 凹陷区（输入框 inactive、tag 背景）
  overlay: "#0f172acc"          # 模态遮罩（ink + 80% alpha）

  # Status — 双通道使用（颜色 + 图标 / 文案）
  success: "#16a34a"            # oklch(58% 0.17 145)
  success-soft: "#f0fdf4"
  warning: "#d97706"            # oklch(63% 0.16 65) — 上传警告 / 容量将满
  warning-soft: "#fffbeb"
  danger: "#dc2626"             # oklch(57% 0.22 27) — 删除 / 错误
  danger-soft: "#fef2f2"
  info: "#0284c7"               # oklch(58% 0.14 230)
  info-soft: "#f0f9ff"

  # Roles
  admin-mark: "#d97706"         # 管理员徽章（warning hue 复用）
  star-mark: "#eab308"          # 收藏星标黄（oklch(80% 0.15 90)）
typography:
  display:
    fontFamily: "'Inter Variable', 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif"
    fontSize: "1.5rem"          # 24px — 页面 H1 / 大模块标题
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "-0.015em"
  headline:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "1.125rem"        # 18px — 区块 / Dialog 标题
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: "-0.01em"
  title:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "0.9375rem"       # 15px — 卡片标题 / Toolbar 主文字
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "normal"
  body:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "0.875rem"        # 14px — 列表行、表格正文、按钮、表单标签
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  body-strong:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.5
    letterSpacing: "normal"
  caption:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "0.75rem"         # 12px — 文件大小、时间、辅助提示
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
  caption-strong:
    fontFamily: "{typography.display.fontFamily}"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "normal"
  mono:
    fontFamily: "'JetBrains Mono', 'SF Mono', 'Cascadia Mono', ui-monospace, Menlo, Consolas, monospace"
    fontSize: "0.8125rem"       # 13px — 哈希值、路径、调试态
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: "normal"
rounded:
  none: "0"
  xs: "4px"                     # 复选框、小 chip
  sm: "6px"                     # 输入框、菜单项
  md: "8px"                     # 按钮、卡片、Dialog header 行（项目主流值）
  lg: "12px"                    # 卡片大容器、文件卡片
  xl: "16px"                    # Dialog 容器
  full: "9999px"                # 头像、状态点、收藏标
spacing:
  "0.5": "2px"
  "1": "4px"
  "1.5": "6px"
  "2": "8px"
  "3": "12px"
  "4": "16px"
  "5": "20px"
  "6": "24px"
  "8": "32px"
  "10": "40px"
  "12": "48px"
  "16": "64px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-on}"
    typography: "{typography.body-strong}"
    rounded: "{rounded.md}"
    padding: "0 14px"
    height: "32px"
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
    textColor: "{colors.primary-on}"
  button-primary-active:
    backgroundColor: "{colors.primary-active}"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "0 12px"
    height: "32px"
  button-ghost-hover:
    backgroundColor: "{colors.surface-sunken}"
    textColor: "{colors.ink}"
  button-danger:
    backgroundColor: "{colors.danger}"
    textColor: "{colors.primary-on}"
    typography: "{typography.body-strong}"
    rounded: "{rounded.md}"
    padding: "0 14px"
    height: "32px"
  input-text:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "0 12px"
    height: "36px"
  input-text-focus:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.lg}"
    padding: "16px"
  dialog:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.xl}"
    padding: "0"
  nav-item:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "6px 12px"
    height: "32px"
  nav-item-active:
    backgroundColor: "{colors.primary-soft}"
    textColor: "{colors.primary}"
    typography: "{typography.body-strong}"
  chip:
    backgroundColor: "{colors.surface-sunken}"
    textColor: "{colors.ink-2}"
    typography: "{typography.caption}"
    rounded: "{rounded.full}"
    padding: "2px 10px"
    height: "22px"
---

# Design System: NetDisk

## 1. Overview

**Creative North Star: "The Quiet Workshop"**

NetDisk 不是一个网盘消费产品，是一个**自部署的工作间**。台面（content surface）保持干净，工具（chrome）退到边缘，文件本身——缩略图、文件名、文件结构——才是被反复看到的主角。系统的整体气质是**克制的、可被信任的、桌面优先的**：节奏稳、对比稳、动效短，每个状态都有明确的视觉反馈。

视觉上属于「现代工具软件」的家族（参考 Linear / Vercel / Raycast 的工艺水准），但**不是它们的模仿**：NetDisk 是面向文件、不是面向代码 / 设计 / 启动器的，所以更倾向「Finder 的克制 + Linear 的精度」的中间地带——既不密集到 OA 系统，也不松散到 marketing landing。

这套系统明确**拒绝**几类常见走向：通用 SaaS 模板（紫蓝渐变、emoji icon、glassmorphism 卡片堆）、传统国内网盘的消费向 UI（广告位、强营销、红黄 CTA）、传统企业管理后台（Ant Design Pro 蓝白、密集表格、毫无个性）、玩具感（亮黄亮粉、卡通插画）。

**Key Characteristics:**

- 灰阶为主、强调色按事件出现，**不**按位置预先铺色
- 一套精确的 cool-gray 中性体系（轻微偏蓝 ~252°），不是 Tailwind 默认 gray
- 一种字体（Inter），靠 weight + size 建立层级
- 圆角小而稳定：`md` (8px) 是主流值，`lg` (12px) 用于容器
- 平面优先（flat-by-default），阴影仅在 overlay / floating 层出现
- 动效 150–200ms，cubic-bezier(0.2, 0, 0, 1)，`prefers-reduced-motion` 全覆盖

## 2. Colors

整套调色板是 **cool gray + 一种克制的蓝**：中性色不取 Tailwind 默认 gray（那是 hue=265 偏紫），而是 hue ≈ 252 的冷灰，让强调色融入而不刺眼。强调色只在用户事件需要它的地方出现。

### Primary
- **Workshop Blue** (`#2563eb` / `oklch(57% 0.18 252)`)：主操作按钮、当前选中的导航项、上传 / 选中态、链接、Focus 环。比 Tailwind 默认 `blue-500` 略克制，比 `blue-600` 略亮一线。
- **Workshop Blue Hover** (`#1d4ed8`)：主按钮 hover、链接 hover。
- **Workshop Blue Active** (`#1e40af`)：主按钮 pressed。
- **Workshop Blue Soft** (`#eff6ff`)：当前选中的导航项背景、文件多选高亮底色、上传中状态条软底。**唯一允许大面积出现强调色的场景**。

### Neutral (Cool Ink Ramp)
- **Ink** (`#0f172a`)：主文本，文件名、表单填写值、Dialog 标题。
- **Ink-2** (`#334155`)：次要文本、未激活的导航 label、Toolbar 主文字。
- **Ink-3** (`#64748b`)：三级文本、辅助说明、面包屑非末段。
- **Ink-4** (`#94a3b8`)：占位符、未选中状态图标。
- **Ink-5** (`#cbd5e1`)：禁用态文本。

- **Surface** (`#ffffff`)：内容卡片 / 列表行 / Dialog body 的底色。
- **Surface Muted** (`#f8fafc`)：body 整体底色、侧边面板。**比 Tailwind `gray-50` 略冷一线，统一在 252° hue 上**。
- **Surface Sunken** (`#f1f5f9`)：输入框 inactive 背景、chip 背景、键盘 kbd 标签。
- **Line** (`#e2e8f0`)：主分隔线、卡片边框、表格行分隔。
- **Line Soft** (`#f1f5f9`)：极弱分隔（仅在密集列表里用，几乎贴近背景）。
- **Overlay** (`rgba(15, 23, 42, 0.8)`)：模态遮罩。

### Tertiary (Status & Roles)
- **Success Green** (`#16a34a`)：操作成功、上传完成。配 `success-soft` (`#f0fdf4`) 做温和提示底色。
- **Warning Amber** (`#d97706`)：上传警告、容量将满、管理员徽章。配 `warning-soft` (`#fffbeb`)。
- **Danger Red** (`#dc2626`)：删除操作、错误提示。配 `danger-soft` (`#fef2f2`) 做警告框。**不**用红色做装饰性强调。
- **Info Sky** (`#0284c7`)：纯信息提示。配 `info-soft` (`#f0f9ff`)。
- **Star Yellow** (`#eab308`)：收藏星标。**唯一**的黄色用途。

### Named Rules

**The One Voice Rule.** Workshop Blue 在任一屏幕上的视觉面积不超过 **10%**。它是用户事件的回音，不是装饰。Navbar 上的 logo bg、Hero 区的大色块、卡片左侧 stripe 这类「预先铺色」的用法**全部禁止**。

**The Cool Hue Rule.** 中性灰统一在 hue ≈ 252° 的冷灰上（偏 primary 色相 0.005–0.02 chroma）。不混用 Tailwind 默认 `gray-*`（偏紫）、`zinc-*`（中性）、`slate-*`（接近但不完全相同的色相）。所有中性色变量必须从本节定义的 Ink / Surface / Line 体系里取。

**The Two-Channel Status Rule.** 任何状态信息（错误、警告、成功）必须同时用**颜色 + 图标 / 文案**两个通道传达。不允许「只用红色文字」这种单通道传达——色盲用户拿不到这个信号。

## 3. Typography

**Display / Body / Label Font:** **Inter Variable**（fallback：`-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`）

**Mono Font:** **JetBrains Mono**（fallback：`'SF Mono', ui-monospace, Menlo, Consolas, monospace`）

**Character.** 一种字体走天下。Inter 是当代产品 UI 的标准答案：x-height 充足、数字等宽、weight 轴丰富。靠 **weight (400 / 500 / 600) + size + ink ramp** 建立层级，**不**靠多字体或大对比字号。Mono 字体仅在哈希、路径、调试态出现，不在常规 UI 里使用。

### Hierarchy
- **Display** (600, 1.5rem / 24px, line-height 1.25, letter-spacing -0.015em)：页面 H1，例如「文件」「媒体库」「账户中心」页头。
- **Headline** (600, 1.125rem / 18px, line-height 1.35, letter-spacing -0.01em)：Dialog 标题、设置区块大标题。
- **Title** (500, 0.9375rem / 15px, line-height 1.4)：卡片标题、Toolbar 主文字、Empty State 主文。
- **Body** (400, 0.875rem / 14px, line-height 1.5)：列表行、表格正文、按钮、表单标签、菜单项（**项目主流字号，绝大多数 UI 都在这一档**）。
- **Body Strong** (500, 0.875rem)：当前选中的导航项、强调的内联文本。
- **Caption** (400, 0.75rem / 12px, line-height 1.4)：文件大小、时间戳、辅助提示。
- **Caption Strong** (500, 0.75rem)：chip 内文字、状态徽章。
- **Mono** (400, 0.8125rem / 13px)：SHA、路径、URL、调试值。

### Named Rules

**The Fixed Scale Rule.** 产品 UI 字号是**固定 rem**，不用 `clamp()`。用户在桌面上以一致 DPI 浏览，流体字号在侧边栏会缩到失真。响应式靠**结构折叠**（侧栏收起、表格列减少），不靠字号。

**The Tabular Number Rule.** 所有展示数字（文件大小、字节数、时间、进度百分比）必须用 `font-variant-numeric: tabular-nums`，避免数字宽度抖动。

**The No-Display-Font-in-Labels Rule.** 不允许在按钮、表单 label、菜单项、表格表头里使用比 Body 更大或更细的字号去「设计感」。Body (14px / 400) 在 99% 的场合是正确答案。

## 4. Elevation

**Flat-by-default。** 内容平面层（surface / surface-muted）之间**只**靠颜色与 1px 边框区分，**不**用阴影。阴影只出现在 **floating 层**（Dropdown / Popover / Dialog / Toast），用来明确「这一层悬浮在内容之上」。

NetDisk 的「层级感」由 cool gray 的两层底色（`surface` 与 `surface-muted` 与 `surface-sunken`）+ 1px `line` 边框承担，而不是阴影堆叠。这避免了 SaaS 卡片堆里常见的浮夸感。

### Shadow Vocabulary

- **shadow-pop** (`0 1px 2px rgba(15, 23, 42, 0.06), 0 4px 12px rgba(15, 23, 42, 0.08)`)：Dropdown / Popover / Context Menu。轻盈、贴近触发点。
- **shadow-dialog** (`0 10px 38px rgba(15, 23, 42, 0.18), 0 10px 20px rgba(15, 23, 42, 0.10)`)：Dialog / Drawer。明确的离台感。
- **shadow-toast** (`0 8px 24px rgba(15, 23, 42, 0.12)`)：Toast 通知。

### Named Rules

**The Flat-At-Rest Rule.** 卡片、表格、列表行在 rest 状态**没有阴影**。允许的视觉反馈：边框颜色变化（hover → primary 软色、active → primary）、背景色变化（hover → surface-sunken）。不允许 hover 时浮起。

**The No-Glassmorphism Rule.** `backdrop-filter: blur()` **仅允许**用于 Navbar 的 `bg-white/80 backdrop-blur-sm`（已有用法，保留）和模态遮罩。**禁止**在卡片、面板、按钮、tooltip 上使用 glassmorphism——那是通用 SaaS 模板的 saturated 反射。

## 5. Components

每个组件至少要有 default / hover / focus-visible / active / disabled 五个状态。loading / error / empty / selected 按需。所有状态切换走 150–200ms 缓动，`cubic-bezier(0.2, 0, 0, 1)`（ease-out-quart）。

### Buttons
- **Shape:** 圆角 `md` (8px)。**所有按钮统一这一值**，不允许 `rounded-full` pill 出现在按钮上（pill 留给 chip）。
- **Height:** Primary / Ghost / Danger 均为 `32px`（h-8），icon-only `28px`。**不**用 36px+ 的大按钮——这是 marketing 移植到产品的常见 tell。
- **Primary:** `bg-primary text-white px-3.5 h-8 font-medium`，hover → `primary-hover`，active → `primary-active`，focus-visible → 2px primary-soft 外环。
- **Ghost:** `text-ink-2 px-3 h-8`，hover → `bg-surface-sunken text-ink`。导航、Toolbar 默认按钮形态。
- **Danger:** 同 Primary 形态，色用 `danger`。**只**用于「删除 / 永久删除 / 清空回收站」三类不可逆操作，**不**用于常规危险提示。
- **Disabled:** `text-ink-5`，无 hover 反应，cursor: not-allowed。
- **Loading:** 按钮内联 spinner + 文案，按钮宽度锁定（避免抖动）。

### Inputs
- **Style:** 圆角 `md` (8px)，高 `36px` (h-9)，1px `line` 边框，bg `surface`。inner padding 12px。
- **Focus:** 边框 → `primary`，外加 2px `primary-soft` 外环。**不**用 box-shadow 替代边框（避免错位）。
- **Error:** 边框 → `danger`，错误文案在下方以 caption + `danger` 显示，行高保留位避免布局抖动。
- **Disabled:** bg → `surface-sunken`，text → `ink-4`。

### Cards (文件卡片、媒体卡片、设置卡片)
- **Corner:** `lg` (12px)。
- **Background:** `surface`，1px `line` 边框。
- **Padding:** 内部 `16px` (p-4)，紧凑卡片 `12px`。
- **Selected:** 边框 → `primary`，bg → `primary-soft`。
- **Hover:** **仅**在交互卡（可点击的文件 / 媒体）上：bg → `surface-muted`，边框保持 `line`。**不**做位移 / 阴影浮起。
- **Nested cards are forbidden.** 文件夹卡片里放更小的文件卡片这种结构必须重新设计。

### Dialogs
- **Shape:** 圆角 `xl` (16px)。
- **Header:** 1px `line-soft` 下边框，padding `12px 20px`，title 用 Title 字号 (15/500)，关闭按钮 ghost 形态。
- **Body:** padding `16px 20px`，overflow-auto。
- **Footer:** 1px `line-soft` 上边框，padding `12px 20px`，按钮右对齐，主按钮在右。
- **Overlay:** `rgba(15, 23, 42, 0.5)`，进入 / 退出走 200ms 淡入淡出。
- **Dialog 已存在的 close 动画必须保留**（最近修复过这个回归，见 commit `7620669`）。

### Navigation (Navbar)
- 顶部 sticky `h-14` (56px)，bg `surface/80` + `backdrop-blur-sm`，下边框 `line`。
- 居中 `max-w-6xl`，左：logo + brand，中：主导航三项（文件 / 媒体 / 相册），右：搜索 + 账户下拉 + 语言切换。
- **导航项 default:** `text-ink-2`，hover → `bg-surface-sunken text-ink`。
- **导航项 active:** `bg-primary-soft text-primary font-medium`。**唯一**允许 primary-soft 大面积出现的场景之一。
- 移动端（< 768px）：折叠为汉堡菜单 + 下拉，**不**做底部 Tab Bar（这不是 mobile-first 产品）。

### Dropdown / Popover / Context Menu
- bg `surface`，圆角 `md` (8px)，shadow `shadow-pop`，1px `line` 边框。
- item padding `6px 10px`，hover → `bg-surface-sunken`。
- 分隔线 1px `line-soft`。
- 进入 / 退出走 150ms 缩放 + 淡入（`scale: 0.96 → 1`, `opacity: 0 → 1`），origin 跟随触发位置。

### Chips / Tags / Status Badges
- **Shape:** `rounded-full`，高 `22px`，padding `2px 10px`。
- **Default:** bg `surface-sunken`，text `ink-2`，caption-strong 字号。
- **Status 变体:** `success` / `warning` / `danger` / `info` 各配 soft 底色 + 同色深字（保证 ≥4.5:1）。
- **Star (收藏标):** 单独的纯色填充图标，**不**做圆角背景包裹。

### File Row (列表视图，签名组件)
- 行高 `44px`（密集模式 36px），左：缩略图 / MimeIcon 16px，中：文件名（body），右：大小（caption mono tabular）、修改时间（caption）、操作菜单。
- hover → `bg-surface-sunken`，selected → `bg-primary-soft`，selected hover → 同色保持（不变深）。
- 主体一行截断，**不**换行。文件名 hover 显示 native tooltip。
- 右键 / 长按弹 Context Menu（已有用法），单击选中、双击进入 / 预览。

### Upload Task Panel (签名组件)
- 浮动在右下，固定宽度 `380px`，可最小化为 chip。
- 每条任务一行：文件名 + 进度条 + 状态图标 + 操作（暂停 / 重试 / 取消）。
- 进度条高 `4px`，圆角 `full`，bg `surface-sunken`，fill `primary`（warning 状态用 `warning`，error 用 `danger`）。
- **进度条用 transform: scaleX() 动画**，不动 width（避免布局回流）。

## 6. Do's and Don'ts

### Do:
- **Do** 使用本规范定义的 cool-gray 中性体系（`ink-*` / `surface-*` / `line-*`），它们都对齐到 hue ≈ 252°。
- **Do** 在所有数字展示场景启用 `font-variant-numeric: tabular-nums`。
- **Do** 把按钮统一在 `h-8` (32px) + `rounded-md` (8px)。
- **Do** 让强调色 Workshop Blue 在屏幕上的视觉面积保持在 **10% 以内**（One Voice Rule）。
- **Do** 在 floating 层（Dropdown / Dialog / Toast）才使用阴影。
- **Do** 给每个交互元素提供 default / hover / focus-visible / active / disabled 五态。
- **Do** 在错误 / 警告状态同时使用颜色 + 图标 + 文案三通道。
- **Do** 长任务（上传 / 转码 / 删除批量）给出明确的「在哪一步 / 还要多久 / 失败可重试」反馈。
- **Do** 在 `prefers-reduced-motion: reduce` 下把所有 transform / scale 动画降级为即时切换或交叉淡出。
- **Do** 用 `border-line` 1px 边框 + 两层底色（`surface` / `surface-muted` / `surface-sunken`）构建层次，而不是阴影。

### Don't:
- **Don't** 用紫蓝渐变 Hero / 大色块装饰背景 / `background-clip: text` 渐变文字——这是 **通用 SaaS 模板** 的 saturated 反射，PRODUCT.md 已将其列为头号 anti-reference。
- **Don't** 把 emoji 当作功能图标。所有功能图标走 `@lucide/svelte`，统一描边 1.5px，size 14 / 15 / 16 三档。
- **Don't** 在卡片 / 列表 / 按钮上使用 `backdrop-filter: blur()`（glassmorphism）。仅允许在 Navbar 与 Modal Overlay 上使用。
- **Don't** 写 `border-left: 4px solid <color>` 这类侧条 stripe 作 callout / alert 装饰——属于「绝对禁止」清单。
- **Don't** 在产品里放营销页的 hero-metric 组件（大数字 + 小标签 + 渐变 accent）。这是 SaaS cliché。
- **Don't** 仿照传统国内网盘做广告位、强营销弹窗、红黄 CTA、礼包入口。
- **Don't** 仿照 Ant Design Pro 做蓝白企业后台：标题下面铺 primary 横条、表格表头蓝灰填充、Tab 下划线 4px——一个都不行。
- **Don't** 用卡通插画、亮黄亮粉、过度圆角（`rounded-2xl` 以上）、bouncy 弹簧动效。
- **Don't** 用 Tailwind 默认 `gray-*` / `slate-*` / `zinc-*` 类名。中性色必须走本规范定义的 `ink-*` / `surface-*` / `line-*`。
- **Don't** 用 `clamp()` 流体字号。产品 UI 字号固定。
- **Don't** 用大于 `rounded-xl` (16px) 的圆角，除了头像与状态点。
- **Don't** 嵌套卡片。卡片里只能放列表 / 表单 / 文本，**不**能再放卡片。
- **Don't** 给 rest 态卡片加阴影。Flat-At-Rest Rule。
- **Don't** 写「向上浮起」(`translateY(-2px)` + 加深阴影) 的 hover 反馈。Linear / Vercel / Raycast 都不这么做。
- **Don't** 用纯装饰 entrance 动画扫过整页。动效必须服务于因果。
- **Don't** 用 `border-radius` 在按钮上做 `rounded-full` (pill)。pill 留给 chip 和头像。
- **Don't** 把 display 字号或 light weight 用在按钮 / 表单标签 / 菜单项上。Body (14/400) 在 99% 的场合是正确答案。
