import { tx } from '@/i18n/text';
export type FormFieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'select'
  | 'radio'
  | 'checkbox'
  | 'date'
  | 'datetime'
  | 'file'
  | 'rating'
  | 'switch'
  | 'cascader';

export type VisibleOperator = '==' | '!=' | '>' | '<' | 'in';

export interface FieldOption {
  label: string;
  value: string | number | boolean;
  children?: FieldOption[];
}

export interface FieldValidation {
  min_length?: number;
  max_length?: number;
  min_value?: number;
  max_value?: number;
  pattern?: string;
  message?: string;
}

export interface VisibleWhen {
  field: string;
  operator: VisibleOperator;
  value: string | number | boolean | Array<string | number | boolean>;
}

export interface FormField {
  name: string;
  label: string;
  type: FormFieldType;
  required?: boolean;
  placeholder?: string;
  default?: string | number | boolean | null;
  options?: FieldOption[];
  validation?: FieldValidation;
  visible_when?: VisibleWhen | null;
  props?: Record<string, string | number | boolean>;
}

export interface FormSchema {
  fields: FormField[];
}

export interface FormEngineProps {
  schema: FormSchema;
  initialValues?: Record<string, unknown>;
  onChange?: (values: Record<string, unknown>) => void;
  onSubmit?: (values: Record<string, unknown>) => void | Promise<void>;
  loading?: boolean;
  layout?: 'horizontal' | 'vertical' | 'inline';
  submitText?: string;
  showSubmit?: boolean;
}

export const FIELD_TYPE_LABELS: Record<FormFieldType, string> = {
  get text() {
    return tx('单行文本');
  },
  get textarea() {
    return tx('多行文本');
  },
  get number() {
    return tx('数字');
  },
  get select() {
    return tx('下拉选择');
  },
  get radio() {
    return tx('单选');
  },
  get checkbox() {
    return tx('多选');
  },
  get date() {
    return tx('日期');
  },
  get datetime() {
    return tx('日期时间');
  },
  get file() {
    return tx('文件');
  },
  get rating() {
    return tx('评分');
  },
  get switch() {
    return tx('开关');
  },
  get cascader() {
    return tx('级联');
  },
};

export const FIELD_TYPES = Object.keys(FIELD_TYPE_LABELS) as FormFieldType[];
