import type { ProLayoutProps } from "@ant-design/pro-components";

export const Settings: ProLayoutProps & {
  logo?: string;
} = {
  navTheme: "light",
  colorPrimary: "#1890ff",
  layout: "side",
  contentWidth: "Fluid",
  fixedHeader: true,
  fixSiderbar: true,
  colorWeak: false,
  title: "NetDisk Admin",
  logo: "https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg",
  iconfontUrl: "",
  splitMenus: false,
  siderMenuType: "sub",
};

export const proConfig = {
  token: {},
};

export default Settings;
