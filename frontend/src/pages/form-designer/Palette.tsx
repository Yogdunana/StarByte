import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Button, List } from 'antd';
import { FIELD_TYPE_LABELS, FIELD_TYPES, type FormFieldType } from '@/components/FormEngine';

interface Props {
  onAdd: (type: FormFieldType) => void;
}

const Palette: React.FC<Props> = ({ onAdd }) => {
  useLocale();
  return (
    <div>
      <div style={{ fontWeight: 600, marginBottom: 8 }}>{tx('字段类型')}</div>
      <List
        size="small"
        dataSource={FIELD_TYPES}
        renderItem={(t) => (
          <List.Item
            draggable
            onDragStart={(e) => {
              e.dataTransfer.setData('field-type', t);
            }}
            actions={[
              <Button key="add" type="link" size="small" onClick={() => onAdd(t)}>
                {tx('添加')}
              </Button>,
            ]}
          >
            {FIELD_TYPE_LABELS[t]}
          </List.Item>
        )}
      />
    </div>
  );
};

export default Palette;
