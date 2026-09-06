import React from 'react';
import {
  Cascader, Checkbox, DatePicker, Input, InputNumber, Radio, Rate, Select, Switch, Upload, message,
} from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import dayjs, { type Dayjs } from 'dayjs';
import { uploadFile } from '@/api/file';
import type { FormField, FieldOption } from './types';

interface ControlOption {
  label: string;
  value: string | number;
  children?: ControlOption[];
}

function toOptions(opts: FieldOption[] | undefined): ControlOption[] {
  return (opts || []).map((o) => ({
    label: o.label,
    value: typeof o.value === 'boolean' ? String(o.value) : o.value,
    children: o.children?.length ? toOptions(o.children) : undefined,
  }));
}

function asDayjs(v: unknown): Dayjs | undefined {
  if (dayjs.isDayjs(v)) return v;
  if (typeof v === 'string' && v) {
    const d = dayjs(v);
    return d.isValid() ? d : undefined;
  }
  return undefined;
}

interface ControlProps {
  field: FormField;
  value?: unknown;
  onChange?: (v: unknown) => void;
}

const FieldControl: React.FC<ControlProps> = ({ field, value, onChange }) => {
  const ph = field.placeholder;
  const maxRate = typeof field.props?.max === 'number' ? field.props.max : 5;
  const accept = typeof field.props?.accept === 'string' ? field.props.accept : undefined;
  const maxSize = typeof field.props?.max_size === 'number' ? field.props.max_size : 10;

  switch (field.type) {
    case 'textarea':
      return <Input.TextArea rows={4} placeholder={ph} value={value as string | undefined} onChange={(e) => onChange?.(e.target.value)} />;
    case 'number':
      return <InputNumber style={{ width: '100%' }} placeholder={ph} value={value as number | undefined} onChange={(v) => onChange?.(v)} />;
    case 'select':
      return <Select allowClear placeholder={ph || '请选择'} options={toOptions(field.options)} value={value as string | number | undefined} onChange={(v) => onChange?.(v)} />;
    case 'radio':
      return <Radio.Group options={toOptions(field.options)} value={value} onChange={(e) => onChange?.(e.target.value)} />;
    case 'checkbox':
      return <Checkbox.Group options={toOptions(field.options)} value={value as Array<string | number> | undefined} onChange={(v) => onChange?.(v)} />;
    case 'date':
      return <DatePicker style={{ width: '100%' }} value={asDayjs(value)} onChange={(_, ds) => onChange?.(ds)} />;
    case 'datetime':
      return <DatePicker showTime style={{ width: '100%' }} value={asDayjs(value)} onChange={(_, ds) => onChange?.(ds)} />;
    case 'switch':
      return <Switch checked={Boolean(value)} onChange={(v) => onChange?.(v)} />;
    case 'rating':
      return <Rate count={maxRate} value={typeof value === 'number' ? value : 0} onChange={(v) => onChange?.(v)} />;
    case 'cascader':
      return <Cascader allowClear placeholder={ph || '请选择'} options={toOptions(field.options)} value={value as (string | number)[] | undefined} onChange={(v) => onChange?.(v)} />;
    case 'file': {
      const customRequest: UploadProps['customRequest'] = async (options) => {
        const formData = new FormData();
        formData.append('file', options.file as File);
        try {
          const res = await uploadFile(formData);
          options.onSuccess?.(res);
          onChange?.(res.id);
        } catch (err) {
          options.onError?.(err as Error);
          message.error('上传失败');
        }
      };
      const before: UploadProps['beforeUpload'] = (file) => {
        if (file.size / 1024 / 1024 > maxSize) {
          message.error(`文件不能超过 ${maxSize}MB`);
          return Upload.LIST_IGNORE;
        }
        return true;
      };
      return (
        <Upload maxCount={1} accept={accept} customRequest={customRequest} beforeUpload={before}>
          <a><UploadOutlined /> 上传文件</a>
        </Upload>
      );
    }
    default:
      return <Input placeholder={ph} value={value as string | undefined} onChange={(e) => onChange?.(e.target.value)} />;
  }
};

export default FieldControl;
