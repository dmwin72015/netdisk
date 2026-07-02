import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { ProLayout } from "@ant-design/pro-components";
import { AvatarDropdown, Footer, LangDropdown } from "@/components";
import { useAuthStore } from "@/utils/auth";
import { useTranslation } from "react-i18next";
import {
  DashboardOutlined,
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  AuditOutlined,
  ToolOutlined,
  SettingOutlined,
  HomeOutlined,
  LogoutOutlined,
} from "@ant-design/icons";
import { Dropdown, Avatar, Space } from "antd";
import defaultSettings from "@/config/defaultSettings";

const BasicLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user: _user } = useAuthStore();
  const { t } = useTranslation();

  const handleLogout = () => {
    localStorage.removeItem("nd.access");
    localStorage.removeItem("nd.user");
    sessionStorage.removeItem("nd.access");
    sessionStorage.removeItem("nd.user");
    navigate("/login");
  };

  const menuItems = [
    {
      path: "/admin/dashboard",
      name: t("sidebar.dashboard"),
      icon: <DashboardOutlined />,
    },
    { path: "/admin/users", name: t("sidebar.users"), icon: <UserOutlined /> },
    {
      path: "/admin/files",
      name: t("sidebar.files"),
      icon: <FileOutlined />,
      children: [
        { path: "/admin/files", name: t("sidebar.userFiles") },
        { path: "/admin/physical-files", name: t("sidebar.physicalFiles") },
      ],
    },
    {
      path: "/admin/storage",
      name: t("sidebar.storage"),
      icon: <DatabaseOutlined />,
    },
    {
      path: "/admin/activity-logs",
      name: t("sidebar.activityLogs"),
      icon: <AuditOutlined />,
    },
    {
      path: "/admin/cleanup",
      name: t("sidebar.cleanup"),
      icon: <ToolOutlined />,
    },
    {
      path: "/admin/settings",
      name: t("sidebar.settings"),
      icon: <SettingOutlined />,
    },
  ];

  return (
    <ProLayout
      {...defaultSettings}
      location={location}
      menuDataRender={() => menuItems}
      menuItemRender={(item, dom) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            navigate(item.path || "/admin/dashboard");
          }}
        >
          {dom}
        </a>
      )}
      onMenuHeaderClick={() => navigate("/admin/dashboard")}
      headerContentRender={() => (
        <div className="flex items-center gap-4">
          <LangDropdown />
          <AvatarDropdown />
        </div>
      )}
      actionsRender={() => [
        <Dropdown
          key="user"
          menu={{
            items: [
              {
                key: "/admin/dashboard",
                icon: <HomeOutlined />,
                label: t("common.home"),
              },
              { type: "divider" },
              {
                key: "logout",
                icon: <LogoutOutlined />,
                label: t("common.logout"),
                onClick: handleLogout,
              },
            ],
          }}
          placement="bottomRight"
        >
          <Space style={{ cursor: "pointer", marginRight: 12 }}>
            <Avatar
              size={28}
              icon={<UserOutlined />}
              style={{ backgroundColor: "#1677ff" }}
            />
          </Space>
        </Dropdown>,
      ]}
      contentStyle={{
        padding: 20,
        minHeight: "calc(100vh - 56px)",
      }}
      footerRender={() => <Footer />}
    >
      <Outlet />
    </ProLayout>
  );
};

export default BasicLayout;
