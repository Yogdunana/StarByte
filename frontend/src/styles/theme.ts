import type { ThemeConfig } from 'antd';

const lightTheme: ThemeConfig = {
  token: {
    colorPrimary: '#405cac', colorInfo: '#405cac', colorSuccess: '#28764e',
    colorWarning: '#b77826', colorError: '#c34c40',
    colorText: '#253550', colorTextSecondary: '#586b86',
    colorBgLayout: '#edf1f8', colorBgContainer: '#fcfdff',
    colorBorder: '#d4deed', colorBorderSecondary: '#e3e9f3',
    borderRadius: 10, fontSize: 14, controlHeight: 40,
    fontFamily: '"Avenir Next", "PingFang SC", "Noto Sans SC", "Microsoft YaHei", sans-serif',
  },
  components: {
    Layout: { headerBg: '#edf1f8', headerHeight: 72, siderBg: 'transparent', bodyBg: '#edf1f8' },
    Menu: {
      itemBg: 'transparent', subMenuItemBg: 'transparent', itemColor: '#586b86',
      itemSelectedBg: '#e0e9fb', itemSelectedColor: '#334f99', itemHoverBg: '#edf2fc',
      itemBorderRadius: 10, itemHeight: 44, itemMarginInline: 10,
    },
    Table: { headerBg: '#edf2fa', headerColor: '#536782', borderColor: '#e3e9f3', cellPaddingBlock: 14 },
    Card: { borderRadiusLG: 20, headerFontSize: 16 },
    Button: { primaryShadow: '0 4px 12px rgba(64,92,172,.18)' },
  },
};
export const darkComponents: ThemeConfig['components'] = {
  ...lightTheme.components,
  Layout: { headerBg: '#151d2c', headerHeight: 72, siderBg: 'transparent', bodyBg: '#151d2c' },
  Menu: {
    itemBg: 'transparent', subMenuItemBg: 'transparent', itemColor: '#b4c2d7',
    itemSelectedBg: '#304469', itemSelectedColor: '#d2dfff', itemHoverBg: '#2b3b54',
    itemBorderRadius: 10, itemHeight: 44, itemMarginInline: 10,
  },
  Table: { headerBg: '#29364c', headerColor: '#c3d0e4', borderColor: '#34465f', cellPaddingBlock: 14 },
};
export const darkTokens: ThemeConfig['token'] = {
  ...lightTheme.token, colorPrimary: '#adc4ff', colorInfo: '#adc4ff',
  colorText: '#e4ebf8', colorTextSecondary: '#aabbd3',
  colorBgLayout: '#151d2c', colorBgContainer: '#202c40',
  colorBorder: '#465773', colorBorderSecondary: '#34465f',
};
export default lightTheme;
