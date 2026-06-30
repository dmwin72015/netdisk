import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { useTranslation } from 'react-i18next';
import AdminLayout from './components/AdminLayout';
import LoginPage from './pages/Login';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import UserDetail from './pages/UserDetail';
import Files from './pages/Files';
import Storage from './pages/Storage';
import Settings from './pages/Settings';
import ActivityLogs from './pages/ActivityLogs';
import Cleanup from './pages/Cleanup';
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
              <Route index element={<Dashboard />} />
              <Route path="users" element={<Users />} />
              <Route path="users/:id" element={<UserDetail />} />
              <Route path="files" element={<Files />} />
              <Route path="storage" element={<Storage />} />
              <Route path="settings" element={<Settings />} />
              <Route path="logs" element={<ActivityLogs />} />
              <Route path="cleanup" element={<Cleanup />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;
