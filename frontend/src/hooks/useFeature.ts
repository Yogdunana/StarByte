import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { DEFAULT_FEATURE_KEYS, evaluateMyFeatures, type FeatureEvaluate } from '@/api/feature';

export interface FeatureState {
  flags: Record<string, boolean>;
  reasons: Record<string, string>;
  variants: Record<string, string>;
  loading: boolean;
  refresh: () => Promise<void>;
}

const empty: FeatureState = {
  flags: {},
  reasons: {},
  variants: {},
  loading: true,
  refresh: async () => undefined,
};

const FeatureContext = createContext<FeatureState>(empty);

export function mapEvaluate(input?: Record<string, FeatureEvaluate> | null): {
  flags: Record<string, boolean>;
  reasons: Record<string, string>;
  variants: Record<string, string>;
} {
  const flags: Record<string, boolean> = {};
  const reasons: Record<string, string> = {};
  const variants: Record<string, string> = {};
  if (!input) return { flags, reasons, variants };
  Object.entries(input).forEach(([key, value]) => {
    flags[key] = Boolean(value?.enabled);
    reasons[key] = value?.reason || '';
    variants[key] = value?.variant || '';
  });
  return { flags, reasons, variants };
}

export function FeatureProvider({
  children,
  keys = DEFAULT_FEATURE_KEYS,
}: {
  children: React.ReactNode;
  keys?: readonly string[];
}) {
  const [flags, setFlags] = useState<Record<string, boolean>>({});
  const [reasons, setReasons] = useState<Record<string, string>>({});
  const [variants, setVariants] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  // Depend on contents, not array identity — default/inline `keys={['cms.public']}`
  // must not recreate refresh every render (that loops setLoading(true) and blanks gates).
  const keySig = keys.join(',');

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const mapped = mapEvaluate(await evaluateMyFeatures(keySig ? keySig.split(',') : undefined));
      setFlags(mapped.flags);
      setReasons(mapped.reasons);
      setVariants(mapped.variants);
    } catch {
      setFlags({});
      setReasons({});
      setVariants({});
    } finally {
      setLoading(false);
    }
  }, [keySig]);

  useEffect(() => { void refresh(); }, [refresh]);

  const value = useMemo(
    () => ({ flags, reasons, variants, loading, refresh }),
    [flags, reasons, variants, loading, refresh],
  );

  return React.createElement(FeatureContext.Provider, { value }, children);
}

export function useFeatureFlags(): FeatureState {
  return useContext(FeatureContext);
}

/** useFeature('cms.public') — 未命中时默认关闭（fail closed）。 */
export function useFeature(key: string): { enabled: boolean; loading: boolean; reason: string; variant: string } {
  const { flags, reasons, variants, loading } = useFeatureFlags();
  return { enabled: Boolean(flags[key]), loading, reason: reasons[key] || '', variant: variants[key] || '' };
}
