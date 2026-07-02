import { useState, useEffect } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { ProLayout } from '@ant-design/pro-components';
import { AvatarDropdown, Footer, LangDropdown } from '@/components';
import { useAuthStore } from '@/utils/auth';

const BasicLayout = () => {
  const navigate = useNavigate();
  const [pathname, setPathname] = useState(window.location.pathname);
  const { user: _user } = useAuthStore();

  useEffect(() => {
    const handler = () => setPathname(window.location.pathname);
    window.addEventListener('popstate', handler);
    return () => window.removeEventListener('popstate', handler);
  }, []);

  return (
    <ProLayout
      title="NetDisk Admin"
      logo="https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg"
      layout="mix"
      navTheme="light"
      fixSiderbar
      location={{ pathname }}
      menuItemRender={(item, dom) => (
        <a onClick={(e) => { e.preventDefault(); navigate(item.path || '/'); }}>{dom}</a>
      )}
      headerContentRender={() => (
        <div className="flex items-center gap-4">
          <LangDropdown />
          <AvatarDropdown />
        </div>
      )}
      footerRender={() => <Footer />}
      menuDataRender={() => [
        {
          path: '/admin/dashboard',
          name: 'Workplace',
          icon: 'DashboardOutlined',
        },
        {
          path: '/admin/users',
          name: 'User Management',
          icon: 'UserOutlined',
        },
        {
          path: '/admin/files',
          name: 'Files Management',
          icon: 'FileOutlined',
        },
        {
          path: '/admin/storage',
          name: 'Storage Analysis',
          icon: 'DatabaseOutlined',
        },
        {
          path: '/admin/physical-files',
          name: 'Physical Files',
          icon: 'FileTextOutlined',
        },
        {
          path: '/admin/activity-logs',
          name: 'Activity Logs',
          icon: 'HistoryOutlined',
        },
        {
          path: '/admin/cleanup',
          name: 'Cleanup Management',
          icon: 'DeleteOutlined',
        },
        {
          path: '/admin/settings',
          name: 'System Settings',
          icon: 'SettingOutlined',
        },
      ]}
    >
      <Outlet />
    </ProLayout>
  );
};

export default BasicLayout;
