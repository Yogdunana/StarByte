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
  text: '单行文本',
  textarea: '多行文本',
  number: '数字',
  select: '下拉选择',
  radio: '单选',
  checkbox: '多选',
  date: '日期',
  datetime: '日期时间',
  file: '文件',
  rating: '评分',
  switch: '开关',
  cascader: '级联',
};

export const FIELD_TYPES = Object.keys(FIELD_TYPE_LABELS) as FormFieldType[];
