import React, { useMemo } from 'react';
import { Button, Form } from 'antd';
import FieldControl from './FieldControl';
import { fieldRules } from './rules';
import type { FormEngineProps } from './types';
import { isFieldVisible } from './visible';

const FormEngine: React.FC<FormEngineProps> = ({
  schema,
  initialValues,
  onChange,
  onSubmit,
  loading,
  layout = 'vertical',
  submitText = '提交',
  showSubmit = true,
}) => {
  const [form] = Form.useForm<Record<string, unknown>>();
  const defaults = useMemo(() => {
    const acc: Record<string, unknown> = { ...(initialValues || {}) };
    schema.fields.forEach((f) => {
      if (acc[f.name] === undefined && f.default !== undefined && f.default !== null) {
        acc[f.name] = f.default;
      }
    });
    return acc;
  }, [schema.fields, initialValues]);

  return (
    <Form
      form={form}
      layout={layout}
      initialValues={defaults}
      onValuesChange={(_, all) => onChange?.(all)}
      onFinish={(values) => {
        void onSubmit?.(values);
      }}
    >
      <Form.Item shouldUpdate noStyle>
        {() => {
          const values = form.getFieldsValue(true) as Record<string, unknown>;
          return schema.fields.map((field) => (
            isFieldVisible(field, values) ? (
              <Form.Item key={field.name} name={field.name} label={field.label} rules={fieldRules(field)}>
                <FieldControl field={field} />
              </Form.Item>
            ) : null
          ));
        }}
      </Form.Item>
      {showSubmit ? (
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>{submitText}</Button>
        </Form.Item>
      ) : null}
    </Form>
  );
};

export default FormEngine;
