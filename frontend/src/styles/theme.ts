import type { ThemeConfig } from 'antd';

/**
 * Ant Design 主题配置
 */
const theme: ThemeConfig = {
  token: {
    colorPrimary: '#2563eb',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#2563eb',
    borderRadius: 10,
    fontSize: 14,
  },
  components: {
    Layout: {
      headerBg: '#ffffff',
      headerHeight: 64,
      headerPadding: '0 16px',
      siderBg: '#001529',
      bodyBg: '#f0f2f5',
    },
    Menu: {
      darkItemBg: '#001529',
      darkSubMenuItemBg: '#000c17',
    },
    Table: {
      headerBg: '#fafafa',
      headerColor: '#000000',
      borderColor: '#f0f0f0',
    },
    Card: {
      borderRadiusLG: 8,
    },
  },
};

export default theme;
