import { useCallback, useEffect, useMemo, useState } from 'react';
import { getDictItems, type DictItem } from '@/api/dict';
import type { Option } from '@/types/api';

const cache = new Map<string, DictItem[]>();

export function invalidateDict(typeCode?: string): void {
  if (typeCode) {
    cache.delete(typeCode);
    return;
  }
  cache.clear();
}

export function useDict(typeCode: string) {
  const [items, setItems] = useState<DictItem[]>(() => cache.get(typeCode) ?? []);
  const [loading, setLoading] = useState(!cache.has(typeCode));

  const reload = useCallback(async () => {
    if (!typeCode) {
      setItems([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const list = await getDictItems(typeCode);
      cache.set(typeCode, list);
      setItems(list);
    } finally {
      setLoading(false);
    }
  }, [typeCode]);

  useEffect(() => {
    if (cache.has(typeCode)) {
      setItems(cache.get(typeCode) ?? []);
      setLoading(false);
      return;
    }
    void reload();
  }, [typeCode, reload]);

  const options = useMemo<Option[]>(
    () => items.map((item) => ({ label: item.item_label, value: item.item_value })),
    [items],
  );

  const labelOf = useCallback(
    (value: string | number | undefined) => {
      if (value === undefined || value === null) return '';
      const key = String(value);
      return items.find((item) => item.item_value === key)?.item_label ?? key;
    },
    [items],
  );

  return { items, options, labelOf, loading, reload };
}
