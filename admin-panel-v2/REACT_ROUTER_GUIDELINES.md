# React Router v7+ 使用规范

> 本规范覆盖 `react-router-dom` v7 / React Router v7 Data Mode。

## 1. 版本与包引用

- 当前使用 `react-router-dom`（不是 `react-router`）。
- v7 起 `react-router-dom` 依然存在并导出全部 API，不要误以为要换包。
- 所有组件和 Hooks 统一从 `react-router-dom` 导入。

```ts
import {
  createBrowserRouter,
  RouterProvider,
  Outlet,
  Navigate,
  useNavigate,
  useParams,
  useSearchParams,
  useLocation,
} from 'react-router-dom';
```

## 2. 路由创建：Data Mode（唯一模式）

- 必须使用 `createBrowserRouter` + `<RouterProvider>`，禁止使用 `<BrowserRouter>` + `<Routes>` + `<Route>`（旧式 Declarative Mode）。
- 路由器在入口文件创建，通过 `<RouterProvider router={router} />` 注入。

```ts
// App.tsx
import { createBrowserRouter } from 'react-router-dom';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <Home /> },
      { path: 'users', element: <Users /> },
    ],
  },
]);
```

## 3. 入口文件必须挂载 DOM

- 使用 React 18+ `createRoot`，不能只 export 组件不挂载。
- 入口文件职责：渲染 `<RouterProvider>` + 全局 Provider（QueryClient / i18n / Theme）。

```ts
// src/main.tsx
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { router } from '@/App';

createRoot(document.getElementById('root')!).render(
  <RouterProvider router={router} />
);
```

## 4. 布局与嵌套：Outlet

- 父路由用 `element` 放布局组件，子路由在 `children` 里。
- 布局组件内部用 `<Outlet />` 渲染子路由。

```ts
// router
{
  path: '/admin',
  element: <ProtectedRoute><AdminLayout /></ProtectedRoute>,
  children: [
    { path: 'dashboard', element: <Dashboard /> },
  ],
}

// AdminLayout.tsx
const AdminLayout = () => {
  return (
    <ProLayout>
      <Outlet />
    </ProLayout>
  );
};
```

## 5. 导航：禁止 window.location

- 编程式导航用 `useNavigate()`。
- 声明式跳转用 `<Link to="...">` 或 `<NavLink>`。
- 禁止 `window.location.href`、`window.location.replace`（会整页刷新，破坏 SPA）。

```ts
const navigate = useNavigate();
navigate('/users', { replace: true });

// 或
<Link to="/users">Users</Link>
```

## 6. 路由守卫：Navigate 组件

- 用 `<Navigate to="/login" replace />` 实现跳转，不要用 `window.location.href`。
- 守卫组件返回 `<Navigate>` 或 `children`，不要返回 `null` 后手动跳转。

```ts
const ProtectedRoute = ({ children }) => {
  const token = getToken();
  if (!token) return <Navigate to="/login" replace />;
  return children;
};
```

## 7. 懒加载：lazy + Suspense

- 路由级懒加载用 `React.lazy` + `Suspense`。
- fallback 放在路由配置的 `element` 里，不要放在布局组件外层导致整个布局消失。

```ts
const Dashboard = lazy(() => import('@/features/dashboard/pages/Dashboard'));

{
  path: 'dashboard',
  element: (
    <Suspense fallback={<Loading />}>
      <Dashboard />
    </Suspense>
  ),
}
```

## 8. 路径别名

- `@` → `src/`
- `@root` → 项目根目录（用于访问 `package.json`、`vite.config.ts` 等）

```ts
import { useAuthStore } from '@/utils/auth';
import packageJson from '@root/package.json';
```

## 9. 搜索参数与动态路由

- 动态路由参数用 `useParams`。
- 搜索参数用 `useSearchParams`。
- 当前位置用 `useLocation`。

```ts
const { id } = useParams<{ id: string }>();
const [searchParams, setSearchParams] = useSearchParams();
const location = useLocation();
```

## 10. 常见错误清单

| 错误写法 | 正确写法 | 原因 |
|---------|---------|------|
| `window.location.href = '/login'` | `<Navigate to="/login" replace />` | 避免整页刷新 |
| `export default App` 不挂载 | `createRoot(...).render(<RouterProvider ...>)` | 必须挂载到 DOM |
| `<BrowserRouter>` + `<Routes>` | `createBrowserRouter` + `<RouterProvider>` | v7 Data Mode 是唯一推荐模式 |
| 路由组件内 `window.location.pathname` 监听 | 用 `useLocation` 获取路径 | 不侵入浏览器 API |
| `import { HashRouter }` 混用 | 整个项目统一 `createBrowserRouter` | 避免 hash/浏览器历史混用 |
