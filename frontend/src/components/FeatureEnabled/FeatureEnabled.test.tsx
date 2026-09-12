import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import FeatureEnabled from './FeatureEnabled';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@/hooks/useFeature', () => ({
  useFeature: vi.fn(),
}));

import { useFeature } from '@/hooks/useFeature';

describe('FeatureEnabled', () => {
  it('renders children when enabled', () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: true, loading: false, reason: 'boolean' });
    render(<FeatureEnabled flag="cms.public"><span>visible</span></FeatureEnabled>);
    expect(screen.getByText('visible')).toBeInTheDocument();
  });

  it('hides children while loading', () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: false, loading: true, reason: '' });
    const { container } = render(<FeatureEnabled flag="cms.public"><span>hidden</span></FeatureEnabled>);
    expect(container).toBeEmptyDOMElement();
  });

  it('shows fallback when disabled', () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: false, loading: false, reason: 'disabled' });
    render(<FeatureEnabled flag="cms.public" fallback={<span>off</span>}><span>on</span></FeatureEnabled>);
    expect(screen.getByText('off')).toBeInTheDocument();
    expect(screen.queryByText('on')).toBeNull();
  });
});
