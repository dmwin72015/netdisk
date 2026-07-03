import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { useAuthGuard } from '@/utils/auth';
import AdminLayout from '@/components/AdminLayout';
import PageLoading from '@/components/PageLoading';
import ErrorBoundary from '@/components/ErrorBoundary';

dayjs.locale('zh-cn');

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

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <ErrorBoundary>
        <BrowserRouter>
          <Suspense fallback={<PageLoading />}>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/" element={<ProtectedRoute />}>
                <Route element={<AdminLayout />}>
                  <Route index element={<Navigate to="/dashboard" replace />} />
                  <Route path="dashboard" element={<DashboardPage />} />
                  <Route path="users" element={<UsersPage />} />
                  <Route path="users/:id" element={<UserDetailPage />} />
                  <Route path="files" element={<Navigate to="/files/user-files" replace />} />
                  <Route path="files/user-files" element={<FilesPage />} />
                  <Route path="files/physical" element={<PhysicalFilesPage />} />
                  <Route path="files/physical/:id" element={<PhysicalFileDetailPage />} />
                  <Route path="storage" element={<StoragePage />} />
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="logs" element={<ActivityLogsPage />} />
                  <Route path="cleanup" element={<CleanupPage />} />
                </Route>
              </Route>
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </Suspense>
        </BrowserRouter>
      </ErrorBoundary>
    </ConfigProvider>
  );
}

export default App;
