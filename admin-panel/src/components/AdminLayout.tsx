import { Avatar, Dropdown, Space } from 'antd';
import { ProLayout } from '@ant-design/pro-layout';
import type { MenuDataItem } from '@ant-design/pro-layout';
import {
  DashboardOutlined,
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  SettingOutlined,
  AuditOutlined,
  ToolOutlined,
  HomeOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router';
import { useTranslation } from 'react-i18next';

export default function AdminLayout() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    localStorage.removeItem('nd.access');
    localStorage.removeItem('nd.user');
    navigate('/login');
  };

  const menuItems: MenuDataItem[] = [
    { path: '/admin', name: t('sidebar.dashboard'), icon: <DashboardOutlined /> },
    { path: '/admin/users', name: t('sidebar.users'), icon: <UserOutlined /> },
    {
      path: '/admin/files',
      name: t('sidebar.files'),
      icon: <FileOutlined />,
      children: [
        { path: '/admin/files/user-files', name: t('sidebar.userFiles') },
        { path: '/admin/files/physical', name: t('sidebar.physicalFiles') },
      ],
    },
    { path: '/admin/storage', name: t('sidebar.storage'), icon: <DatabaseOutlined /> },
    { path: '/admin/logs', name: t('sidebar.activityLogs'), icon: <AuditOutlined /> },
    { path: '/admin/cleanup', name: t('sidebar.cleanup'), icon: <ToolOutlined /> },
    { path: '/admin/settings', name: t('sidebar.settings'), icon: <SettingOutlined /> },
  ];

  return (
    <ProLayout
      layout="mix"
      navTheme="light"
      fixSiderbar
      fixedHeader
      logo={<span style={{ color: '#fff', fontSize: 18, fontWeight: 700 }}>Admin</span>}
      title="Admin"
      disableMobile
      location={location}
      menuDataRender={() => menuItems}
      menuItemRender={(item, dom) => (
        <a
          onClick={() => navigate(item.path || '/admin')}
          style={{ textDecoration: 'none' }}
        >
          {dom}
        </a>
      )}
      onMenuHeaderClick={() => navigate('/admin')}
      actionsRender={() => [
        <Dropdown
          key="user"
          menu={{
            items: [
              { key: '/admin', icon: <HomeOutlined />, label: t('common.home') },
              { type: 'divider' },
              { key: 'logout', icon: <LogoutOutlined />, label: t('common.logout'), onClick: handleLogout },
            ],
          }}
          placement="bottomRight"
        >
          <Space style={{ cursor: 'pointer', marginRight: 12 }}>
            <Avatar size={28} icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
          </Space>
        </Dropdown>,
      ]}
      contentStyle={{
        background: '#f0f2f5',
        padding: 20,
        minHeight: 'calc(100vh - 56px)',
      }}
    >
      <Outlet />
    </ProLayout>
  );
}