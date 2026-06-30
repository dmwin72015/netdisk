import { useState, useEffect } from 'react';
import { Layout, Menu, Button, Avatar, Dropdown, Space, theme } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  SettingOutlined,
  AuditOutlined,
  ToolOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  HomeOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem('nd.admin.sidebar') === 'true';
    } catch {
      return false;
    }
  });
  const navigate = useNavigate();
  const location = useLocation();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  useEffect(() => {
    localStorage.setItem('nd.admin.sidebar', String(collapsed));
  }, [collapsed]);

  const handleLogout = () => {
    localStorage.removeItem('nd.access');
    localStorage.removeItem('nd.user');
    navigate('/login');
  };

  const menuItems: MenuProps['items'] = [
    { key: '/admin', icon: <HomeOutlined />, label: 'Home' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: handleLogout,
    },
  ];

  const sideMenuItems: MenuProps['items'] = [
    { key: '/admin', icon: <DashboardOutlined />, label: 'Dashboard' },
    { key: '/admin/users', icon: <UserOutlined />, label: 'Users' },
    { key: '/admin/files', icon: <FileOutlined />, label: 'Files' },
    { key: '/admin/storage', icon: <DatabaseOutlined />, label: 'Storage' },
    { key: '/admin/logs', icon: <AuditOutlined />, label: 'Activity Logs' },
    { key: '/admin/cleanup', icon: <ToolOutlined />, label: 'Cleanup' },
    { key: '/admin/settings', icon: <SettingOutlined />, label: 'Settings' },
  ];

  const handleMenuClick = (info: { key: string }) => {
    navigate(info.key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          background: '#1a1a2e',
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          <span
            style={{
              color: '#fff',
              fontSize: 20,
              fontWeight: 700,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
            }}
          >
            {collapsed ? 'A' : 'Admin'}
          </span>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={sideMenuItems}
          onClick={handleMenuClick}
          style={{ background: 'transparent', borderRight: 0 }}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'all 0.2s' }}>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
            position: 'sticky',
            top: 0,
            zIndex: 1,
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: 16, width: 48, height: 48 }}
          />
          <Dropdown menu={{ items: menuItems }} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} />
            </Space>
          </Dropdown>
        </Header>
        <Content
          style={{
            margin: 24,
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
