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

const PropPanel: React.FC<Props> = ({ field, allNames, onChange }) => {
  if (!field) {
    return <div style={{ color: '#999' }}>选择一个字段以编辑属性</div>;
  }
  const needsOptions = ['select', 'radio', 'checkbox', 'cascader'].includes(field.type);
  const optionsText = (field.options || []).map((o) => `${o.label}=${String(o.value)}`).join('\n');

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
        <Form.Item label="选项（每行 标签=值）">
          <Input.TextArea
            rows={4}
            value={optionsText}
            onChange={(e) => {
              const options = e.target.value.split('\n').map((line) => {
                const [label, raw] = line.split('=');
                if (!label?.trim()) return null;
                const value = raw === undefined ? label.trim() : raw.trim();
                const num = Number(value);
                return { label: label.trim(), value: value !== '' && !Number.isNaN(num) && String(num) === value ? num : value };
              }).filter((x): x is { label: string; value: string | number } => x !== null);
              onChange({ options });
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
