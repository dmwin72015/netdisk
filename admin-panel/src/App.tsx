import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { useTranslation } from 'react-i18next';
import AdminLayout from './components/AdminLayout';
import { LoginPage } from './features/login';
import { DashboardPage } from './features/dashboard';
import { UsersPage, UserDetailPage } from './features/users';
import { FilesPage } from './features/files';
import { StoragePage } from './features/storage';
import { SettingsPage } from './features/settings';
import { ActivityLogsPage } from './features/activity-logs';
import { CleanupPage } from './features/cleanup';
import { useAuthGuard } from './utils/auth';

dayjs.locale('zh-cn');

function ProtectedRoute() {
  const { checking, authorized } = useAuthGuard();
  const { t } = useTranslation();
  if (checking) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <div>{t('common.loading')}</div>
      </div>
    );
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
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/admin" element={<ProtectedRoute />}>
            <Route element={<AdminLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="users" element={<UsersPage />} />
              <Route path="users/:id" element={<UserDetailPage />} />
              <Route path="files" element={<FilesPage />} />
              <Route path="storage" element={<StoragePage />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="logs" element={<ActivityLogsPage />} />
              <Route path="cleanup" element={<CleanupPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;
