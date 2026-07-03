import { useState, useEffect } from 'react';
import { Dropdown, Typography } from 'antd';
import { ProLayout, SettingDrawer } from '@ant-design/pro-layout';
import type { ProSettings } from '@ant-design/pro-layout';
import type { MenuDataItem } from '@ant-design/pro-layout';
import {
  DashboardOutlined,
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  SettingOutlined,
  AuditOutlined,
  ToolOutlined,
  LogoutOutlined,
  BellOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation, Link } from 'react-router';
import { getStoredUser, logout } from '@/utils/auth';

const { Text } = Typography;

const STORAGE_KEY = 'admin.layout-settings';

type LayoutSettings = ProSettings & {
  colorPrimary?: string;
  colorWeak?: boolean;
};

const defaultSettings: LayoutSettings = {
  navTheme: 'light',
  layout: 'side',
  contentWidth: 'Fluid',
  fixedHeader: false,
  fixSiderbar: true,
  colorPrimary: '#1890ff',
  colorWeak: false,
  splitMenus: false,
};

function loadSettings(): LayoutSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return { ...defaultSettings, ...JSON.parse(raw) };
  } catch { /* ignore */ }
  return defaultSettings;
}

const routeConfig: MenuDataItem[] = [
  { path: '/dashboard', name: '仪表盘', icon: <DashboardOutlined /> },
  { path: '/users', name: '用户管理', icon: <UserOutlined /> },
  {
    path: '/files',
    name: '文件管理',
    icon: <FileOutlined />,
    children: [
      { path: '/files/user-files', name: '用户文件' },
      { path: '/files/physical', name: '物理文件' },
    ],
  },
  { path: '/storage', name: '存储概览', icon: <DatabaseOutlined /> },
  { path: '/logs', name: '操作日志', icon: <AuditOutlined /> },
  { path: '/cleanup', name: '清理工具', icon: <ToolOutlined /> },
  { path: '/settings', name: '系统设置', icon: <SettingOutlined /> },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const user = getStoredUser();
  const [settings, setSettings] = useState<LayoutSettings>(loadSettings);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  }, [settings]);

  return (
    <>
      <ProLayout
        title="Admin Panel V3"
        logo={
          <svg viewBox="0 0 24 24" width="28" height="28" fill={settings.colorPrimary || '#1890ff'}>
            <path d="M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.18L19.18 7.5 12 10.82 4.82 7.5 12 4.18z" />
          </svg>
        }
        layout={settings.layout}
        fixSiderbar={settings.fixSiderbar}
        fixedHeader={settings.fixedHeader}
        location={location}
        navTheme={settings.navTheme}
        contentWidth={settings.contentWidth}
        splitMenus={settings.splitMenus}
        siderWidth={220}
        menu={{
          locale: false,
          defaultOpenAll: true,
          autoClose: false,
        }}
        route={{ children: routeConfig }}
        menuItemRender={(item, dom) => (
          <Link to={item.path || '/dashboard'}>{dom}</Link>
        )}
        subMenuItemRender={(_, dom) => dom}
        avatarProps={{
          src: undefined,
          icon: <UserOutlined />,
          size: 'small',
          title: user?.username || 'Admin',
          render: (_, dom) => (
            <Dropdown
              menu={{
                items: [
                  { key: 'center', icon: <UserOutlined />, label: '个人中心' },
                  { type: 'divider' },
                  {
                    key: 'logout',
                    icon: <LogoutOutlined />,
                    label: '退出登录',
                    onClick: () => {
                      logout();
                      navigate('/login');
                    },
                  },
                ],
              }}
            >
              {dom}
            </Dropdown>
          ),
        }}
        actionsRender={() => [
          <QuestionCircleOutlined key="help" className="text-base mr-3" />,
          <BellOutlined key="bell" className="text-base mr-3" />,
        ]}
        footerRender={() => (
          <div className="text-center py-4">
            <Text type="secondary">
              Admin Panel V3 &copy; {new Date().getFullYear()}
            </Text>
          </div>
        )}
        contentStyle={{}}
      >
        <Outlet />
      </ProLayout>
      <SettingDrawer
        pathname={location.pathname}
        settings={settings}
        onSettingChange={(newSettings) => setSettings(newSettings as LayoutSettings)}
        disableUrlParams
        hideCopyButton
        hideHintAlert
        enableDarkTheme
        colorList={[
          { key: '蓝色', color: '#1890ff' },
          { key: '紫色', color: '#722ed1' },
          { key: '红色', color: '#f5222d' },
          { key: '火山', color: '#fa541c' },
          { key: '日暮', color: '#fa8c16' },
          { key: '绿色', color: '#52c41a' },
          { key: '青色', color: '#13c2c2' },
          { key: '蓝色系', color: '#2f54eb' },
        ]}
      />
    </>
  );
}
