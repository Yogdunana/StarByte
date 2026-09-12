import type { ThemeConfig } from 'antd';

const fontFamily = '"Avenir Next", "PingFang SC", "Noto Sans SC", "Microsoft YaHei", sans-serif';

const lightTheme: ThemeConfig = {
  token: {
    colorPrimary: '#2f3530',
    colorInfo: '#2f3530',
    colorSuccess: '#3d6b4f',
    colorWarning: '#9a7040',
    colorError: '#a54a42',
    colorText: '#1c1b18',
    colorTextSecondary: '#6b6860',
    colorBgLayout: '#f3f2ee',
    colorBgContainer: '#fafaf7',
    colorBorder: '#ddd9d1',
    colorBorderSecondary: '#eeebe5',
    borderRadius: 10,
    fontSize: 14,
    controlHeight: 40,
    fontFamily,
    boxShadow: '0 1px 0 rgba(28, 27, 24, 0.06)',
    boxShadowSecondary: '0 8px 24px rgba(28, 27, 24, 0.06)',
  },
  components: {
    Layout: { headerBg: '#fafaf7', headerHeight: 64, siderBg: '#fafaf7', bodyBg: '#f3f2ee' },
    Menu: {
      itemBg: 'transparent',
      subMenuItemBg: 'transparent',
      itemColor: '#6b6860',
      itemSelectedBg: '#e4e7e2',
      itemSelectedColor: '#1c1b18',
      itemHoverBg: '#eeebe5',
      itemBorderRadius: 8,
      itemHeight: 40,
      itemMarginInline: 8,
    },
    Table: { headerBg: '#eeebe5', headerColor: '#6b6860', borderColor: '#eeebe5', cellPaddingBlock: 14 },
    Card: { borderRadiusLG: 14, headerFontSize: 16 },
    Button: { primaryShadow: 'none' },
  },
};

export const darkComponents: ThemeConfig['components'] = {
  ...lightTheme.components,
  Layout: { headerBg: '#1c1c1a', headerHeight: 64, siderBg: '#1c1c1a', bodyBg: '#121211' },
  Menu: {
    itemBg: 'transparent',
    subMenuItemBg: 'transparent',
    itemColor: '#a8a49a',
    itemSelectedBg: '#2a2a26',
    itemSelectedColor: '#eceae4',
    itemHoverBg: '#222220',
    itemBorderRadius: 8,
    itemHeight: 40,
    itemMarginInline: 8,
  },
  Table: { headerBg: '#222220', headerColor: '#a8a49a', borderColor: '#33322e', cellPaddingBlock: 14 },
};

export const darkTokens: ThemeConfig['token'] = {
  ...lightTheme.token,
  colorPrimary: '#d4d0c8',
  colorInfo: '#d4d0c8',
  colorText: '#eceae4',
  colorTextSecondary: '#a8a49a',
  colorBgLayout: '#121211',
  colorBgContainer: '#1c1c1a',
  colorBorder: '#33322e',
  colorBorderSecondary: '#222220',
  boxShadow: '0 1px 0 rgba(0, 0, 0, 0.28)',
  boxShadowSecondary: '0 10px 28px rgba(0, 0, 0, 0.32)',
};

export default lightTheme;
