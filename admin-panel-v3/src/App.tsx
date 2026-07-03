import { lazy, Suspense } from 'react';
import { createBrowserRouter, Navigate, Outlet } from 'react-router';
import { useAuthGuard } from '@/utils/auth';
import AdminLayout from '@/components/AdminLayout';
import PageLoading from '@/components/PageLoading';

const LoginPage = lazy(() => import('@/features/login/pages/Login'));
const DashboardPage = lazy(() => import('@/features/dashboard/pages/Dashboard'));
const UsersPage = lazy(() => import('@/features/users/pages/Users'));
const UserDetailPage = lazy(() => import('@/features/users/pages/UserDetail'));
const FilesPage = lazy(() => import('@/features/files/pages/Files'));
const PhysicalFilesPage = lazy(() => import('@/features/physical-files/pages/PhysicalFiles'));
const PhysicalFileDetailPage = lazy(() => import('@/features/physical-files/pages/PhysicalFileDetail'));
const StoragePage = lazy(() => import('@/features/storage/pages/Storage'));
const SettingsPage = lazy(() => import('@/features/settings/pages/Settings'));
const ActivityLogsPage = lazy(() => import('@/features/activity-logs/pages/ActivityLogs'));
const CleanupPage = lazy(() => import('@/features/cleanup/pages/Cleanup'));
const NotFoundPage = lazy(() => import('@/features/not-found/pages/NotFound'));

function ProtectedRoute() {
  const { checking, authorized } = useAuthGuard();
  if (checking) {
    return <PageLoading />;
  }
  return authorized ? <Outlet /> : null;
}

function SuspenseLayout() {
  return (
    <Suspense fallback={<PageLoading />}>
      <Outlet />
    </Suspense>
  );
}

const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <Suspense fallback={<PageLoading />}>
        <LoginPage />
      </Suspense>
    ),
  },
  {
    path: '/',
    element: <ProtectedRoute />,
    children: [
      {
        element: <SuspenseLayout />,
        children: [
          {
            element: <AdminLayout />,
            children: [
              { index: true, element: <Navigate to="/dashboard" replace /> },
              { path: 'dashboard', element: <DashboardPage /> },
              { path: 'users', element: <UsersPage /> },
              { path: 'users/:id', element: <UserDetailPage /> },
              { path: 'files', element: <Navigate to="/files/user-files" replace /> },
              { path: 'files/user-files', element: <FilesPage /> },
              { path: 'files/physical', element: <PhysicalFilesPage /> },
              { path: 'files/physical/:id', element: <PhysicalFileDetailPage /> },
              { path: 'storage', element: <StoragePage /> },
              { path: 'settings', element: <SettingsPage /> },
              { path: 'logs', element: <ActivityLogsPage /> },
              { path: 'cleanup', element: <CleanupPage /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
]);

export default router;