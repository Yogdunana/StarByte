import React from 'react';
import { Button, Empty, Space, Tag } from 'antd';
import { DeleteOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';
import { FIELD_TYPE_LABELS, type FormField } from '@/components/FormEngine';

interface Props {
  fields: FormField[];
  selected?: string;
  onSelect: (name: string) => void;
  onRemove: (name: string) => void;
  onMove: (name: string, dir: -1 | 1) => void;
  onDropType: (type: string) => void;
}

const FieldList: React.FC<Props> = ({ fields, selected, onSelect, onRemove, onMove, onDropType }) => (
  <div
    onDragOver={(e) => e.preventDefault()}
    onDrop={(e) => {
      const t = e.dataTransfer.getData('field-type');
      if (t) onDropType(t);
    }}
    style={{ minHeight: 280, padding: 8, border: '1px dashed #d9d9d9', borderRadius: 8 }}
  >
    {fields.length === 0 ? <Empty description="从左侧拖入或点击添加字段" /> : null}
    {fields.map((f, i) => (
      <div
        key={f.name}
        onClick={() => onSelect(f.name)}
        style={{
          padding: '8px 12px',
          marginBottom: 8,
          borderRadius: 6,
          cursor: 'pointer',
          background: selected === f.name ? '#e6f4ff' : '#fafafa',
          border: '1px solid #f0f0f0',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
          <span>{f.label || f.name} <Tag>{FIELD_TYPE_LABELS[f.type]}</Tag>{f.required ? <Tag color="red">必填</Tag> : null}</span>
          <Space size={0} onClick={(e) => e.stopPropagation()}>
            <Button size="small" type="text" disabled={i === 0} icon={<ArrowUpOutlined />} onClick={() => onMove(f.name, -1)} />
            <Button size="small" type="text" disabled={i === fields.length - 1} icon={<ArrowDownOutlined />} onClick={() => onMove(f.name, 1)} />
            <Button size="small" type="text" danger icon={<DeleteOutlined />} onClick={() => onRemove(f.name)} />
          </Space>
        </div>
      </div>
    ))}
  </div>
);

export default FieldList;
