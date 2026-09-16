import { Grid } from 'antd';

/** Aligns with antd Grid: md=768, lg=992. */
export interface Viewport {
  /** width < 768 */
  phone: boolean;
  /** 768 ≤ width < 992 */
  tablet: boolean;
  /** width < 992 — drawer chrome, no persistent sider */
  compact: boolean;
  /** width ≥ 992 */
  desktop: boolean;
}

export function useViewport(): Viewport {
  const screens = Grid.useBreakpoint();
  const phone = !screens.md;
  const desktop = !!screens.lg;
  const tablet = !phone && !desktop;
  return { phone, tablet, compact: !desktop, desktop };
}
