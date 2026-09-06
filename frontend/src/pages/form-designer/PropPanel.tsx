import React from 'react';
import { Form, Input, InputNumber, Select, Switch } from 'antd';
import { FIELD_TYPE_LABELS, FIELD_TYPES, type FormField, type VisibleOperator } from '@/components/FormEngine';

interface Props {
  field?: FormField;
  allNames: string[];
  onChange: (patch: Partial<FormField>) => void;
}

const operators: { value: VisibleOperator; label: string }[] = [
  { value: '==', label: '等于' },
  { value: '!=', label: '不等于' },
  { value: '>', label: '大于' },
  { value: '<', label: '小于' },
  { value: 'in', label: '包含于' },
];

function parseFieldOptions(text: string): { label: string; value: string | number }[] {
  const chunks: string[] = [];
  for (const line of text.split('\n')) {
    const parts = line.split(/[,，;；]+/).map((s) => s.trim()).filter(Boolean);
    if (parts.length > 1 && parts.every((p) => p.includes('='))) {
      chunks.push(...parts);
    } else if (parts.length === 1) {
      chunks.push(parts[0]);
    } else if (line.trim()) {
      chunks.push(line.trim());
    }
  }
  return chunks.map((line) => {
    const idx = line.indexOf('=');
    const label = (idx >= 0 ? line.slice(0, idx) : line).trim();
    const raw = idx >= 0 ? line.slice(idx + 1).trim() : label;
    if (!label) return null;
    const num = Number(raw);
    return { label, value: raw !== '' && !Number.isNaN(num) && String(num) === raw ? num : raw };
  }).filter((x): x is { label: string; value: string | number } => x !== null);
}

const PropPanel: React.FC<Props> = ({ field, allNames, onChange }) => {
  const fieldName = field?.name ?? '';
  const serializedOptions = (field?.options || []).map((o) => `${o.label}=${String(o.value)}`).join('\n');
  const [optionsDraft, setOptionsDraft] = React.useState({ fieldName, text: serializedOptions });
  if (optionsDraft.fieldName !== fieldName) {
    setOptionsDraft({ fieldName, text: serializedOptions });
  }

  if (!field) {
    return <div style={{ color: '#999' }}>选择一个字段以编辑属性</div>;
  }
  const needsOptions = ['select', 'radio', 'checkbox', 'cascader'].includes(field.type);

  return (
    <Form layout="vertical" size="small">
      <Form.Item label="字段名"><Input value={field.name} onChange={(e) => onChange({ name: e.target.value })} /></Form.Item>
      <Form.Item label="标签"><Input value={field.label} onChange={(e) => onChange({ label: e.target.value })} /></Form.Item>
      <Form.Item label="类型">
        <Select value={field.type} options={FIELD_TYPES.map((t) => ({ value: t, label: FIELD_TYPE_LABELS[t] }))} onChange={(type) => onChange({ type })} />
      </Form.Item>
      <Form.Item label="占位提示"><Input value={field.placeholder} onChange={(e) => onChange({ placeholder: e.target.value })} /></Form.Item>
      <Form.Item label="必填"><Switch checked={Boolean(field.required)} onChange={(required) => onChange({ required })} /></Form.Item>
      {needsOptions ? (
        <Form.Item label="选项（每行一项，或用逗号分隔：大一=1,大二=2）">
          <Input.TextArea
            rows={4}
            value={optionsDraft.text}
            onChange={(e) => {
              const text = e.target.value;
              setOptionsDraft({ fieldName, text });
              onChange({ options: parseFieldOptions(text) });
            }}
          />
        </Form.Item>
      ) : null}
      {(field.type === 'text' || field.type === 'textarea') ? (
        <>
          <Form.Item label="最小长度"><InputNumber min={0} value={field.validation?.min_length} onChange={(n) => onChange({ validation: { ...field.validation, min_length: n ?? undefined } })} /></Form.Item>
          <Form.Item label="最大长度"><InputNumber min={0} value={field.validation?.max_length} onChange={(n) => onChange({ validation: { ...field.validation, max_length: n ?? undefined } })} /></Form.Item>
          <Form.Item label="正则"><Input value={field.validation?.pattern} onChange={(e) => onChange({ validation: { ...field.validation, pattern: e.target.value } })} /></Form.Item>
        </>
      ) : null}
      {(field.type === 'number' || field.type === 'rating') ? (
        <>
          <Form.Item label="最小值"><InputNumber value={field.validation?.min_value} onChange={(n) => onChange({ validation: { ...field.validation, min_value: n ?? undefined } })} /></Form.Item>
          <Form.Item label="最大值"><InputNumber value={field.validation?.max_value} onChange={(n) => onChange({ validation: { ...field.validation, max_value: n ?? undefined } })} /></Form.Item>
        </>
      ) : null}
      {field.type === 'rating' ? (
        <Form.Item label="星级上限"><InputNumber min={1} max={10} value={typeof field.props?.max === 'number' ? field.props.max : 5} onChange={(n) => onChange({ props: { ...field.props, max: n || 5 } })} /></Form.Item>
      ) : null}
      {field.type === 'file' ? (
        <>
          <Form.Item label="accept"><Input value={typeof field.props?.accept === 'string' ? field.props.accept : ''} onChange={(e) => onChange({ props: { ...field.props, accept: e.target.value } })} /></Form.Item>
          <Form.Item label="大小上限 MB"><InputNumber min={1} value={typeof field.props?.max_size === 'number' ? field.props.max_size : 10} onChange={(n) => onChange({ props: { ...field.props, max_size: n || 10 } })} /></Form.Item>
        </>
      ) : null}
      <Form.Item label="条件显示依赖字段">
        <Select
          allowClear
          value={field.visible_when?.field}
          options={allNames.filter((n) => n !== field.name).map((n) => ({ value: n, label: n }))}
          onChange={(name) => onChange({ visible_when: name ? { field: name, operator: field.visible_when?.operator || '==', value: field.visible_when?.value ?? '' } : null })}
        />
      </Form.Item>
      {field.visible_when ? (
        <>
          <Form.Item label="运算符">
            <Select value={field.visible_when.operator} options={operators} onChange={(operator) => onChange({ visible_when: { ...field.visible_when!, operator } })} />
          </Form.Item>
          <Form.Item label="比较值">
            <Input value={String(field.visible_when.value ?? '')} onChange={(e) => onChange({ visible_when: { ...field.visible_when!, value: e.target.value } })} />
          </Form.Item>
        </>
      ) : null}
    </Form>
  );
};

export default PropPanel;
