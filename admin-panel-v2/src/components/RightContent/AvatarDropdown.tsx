import {
  LogoutOutlined,
  UserOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { Avatar } from 'antd';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/utils/auth';
import HeaderDropdown from '../HeaderDropdown';

const menuItems: MenuProps['items'] = [
  {
    key: 'logout',
    icon: <LogoutOutlined />,
    label: '退出登录',
    danger: true,
  },
];

export const AvatarDropdown: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  const onMenuClick: MenuProps['onClick'] = (event) => {
    if (event.key === 'logout') {
      logout();
      navigate('/login');
    }
  };

  if (!user) {
    return <Avatar size="small" icon={<UserOutlined />} />;
  }

  return (
    <HeaderDropdown
      placement="bottomRight"
      arrow
      menu={{
        selectedKeys: [],
        onClick: onMenuClick,
        items: menuItems,
      }}
    >
      <Avatar size="small" icon={<UserOutlined />} />
    </HeaderDropdown>
  );
};
