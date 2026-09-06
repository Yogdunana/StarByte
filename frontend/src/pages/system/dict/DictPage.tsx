import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Card, Select, Space, Tag } from 'antd';
import { getDictItems, getDictTypes, type DictItem, type DictType } from '@/api/dict';
import { useDict, invalidateDict } from '@/hooks/useDict';
import { usePermission } from '@/hooks/usePermission';
import ItemPanel from './ItemPanel';
import TypePanel from './TypePanel';
import './dict.css';

const DictPage: React.FC = () => {
  const canCreate = usePermission('dict:create');
  const canUpdate = usePermission('dict:update');
  const canDelete = usePermission('dict:delete');
  const [types, setTypes] = useState<DictType[]>([]);
  const [items, setItems] = useState<DictItem[]>([]);
  const [selected, setSelected] = useState<DictType>();
  const [typeLoading, setTypeLoading] = useState(false);
  const [itemLoading, setItemLoading] = useState(false);
  const preview = useDict(selected?.code ?? '');
  const itemReqRef = useRef(0);

  const loadTypes = useCallback(async (): Promise<DictType[]> => {
    setTypeLoading(true);
    try {
      const list = await getDictTypes();
      setTypes(list);
      setSelected((prev) => list.find((row) => row.id === prev?.id) || list[0]);
      return list;
    } finally {
      setTypeLoading(false);
    }
  }, []);

  const loadItems = useCallback(async (type?: DictType) => {
    const req = itemReqRef.current + 1;
    itemReqRef.current = req;
    if (!type) {
      setItems([]);
      setItemLoading(false);
      return;
    }
    setItemLoading(true);
    try {
      const list = await getDictItems(type.code, true);
      if (req === itemReqRef.current) {
        setItems(list);
      }
    } finally {
      if (req === itemReqRef.current) {
        setItemLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void loadTypes();
  }, [loadTypes]);
  useEffect(() => {
    void loadItems(selected);
  }, [selected, loadItems]);

  const afterChange = () => {
    const current = selected;
    invalidateDict(current?.code);
    void loadTypes().then((list) => {
      if (current && list.some((row) => row.id === current.id)) {
        void loadItems(current);
        void preview.reload();
      }
    });
  };

  return (
    <div>
      <div className="dict-hero">
        <div>
          <h2>数据字典</h2>
          <p>维护业务枚举类型与选项，公开读接口带 Redis 缓存，前端用 useDict(type) 取启用项。</p>
        </div>
      </div>
      <div className="dict-layout">
        <Card className="dict-card" title="字典类型" extra={<Tag>{types.length} 类</Tag>}>
          <TypePanel
            types={types}
            loading={typeLoading}
            selectedId={selected?.id}
            canCreate={canCreate}
            canUpdate={canUpdate}
            canDelete={canDelete}
            onSelect={setSelected}
            onChanged={afterChange}
          />
        </Card>
        <Card className="dict-card" title={selected ? `${selected.name}（${selected.code}）` : '字典项'}>
          <ItemPanel
            type={selected}
            items={items}
            loading={itemLoading}
            canCreate={canCreate}
            canUpdate={canUpdate}
            canDelete={canDelete}
            onChanged={afterChange}
          />
          <div className="dict-preview">
            <Space direction="vertical" style={{ width: '100%' }}>
              <span>useDict 预览（仅启用项）</span>
              <Select
                style={{ width: '100%' }}
                options={preview.options}
                loading={preview.loading}
                placeholder={preview.labelOf(preview.items[0]?.item_value) || '暂无启用项'}
              />
            </Space>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default DictPage;
