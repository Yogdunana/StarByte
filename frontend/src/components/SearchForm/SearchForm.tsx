import React, { useMemo } from 'react';
import { Form, Input, Select, DatePicker, Button, Space, Row, Col } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import type { SearchField } from '@/types/common';

export interface SearchFormProps {
  fields: SearchField[];
  onSearch: (values: Record<string, unknown>) => void;
  onReset?: () => void;
  loading?: boolean;
  colSpan?: number; // 默认 6（一行4个字段）
}

const { RangePicker } = DatePicker;

/**
 * 搜索表单 — 支持动态字段配置（text/select/date/daterange/number）
 */
const SearchForm: React.FC<SearchFormProps> = ({
  fields,
  onSearch,
  onReset,
  loading = false,
  colSpan = 6,
}) => {
  const [form] = Form.useForm();

  const internalFields = useMemo(
    () => fields.filter((f) => f && f.name && f.label),
    [fields],
  );

  const handleSearch = async () => {
    const values = await form.validateFields();
    // 处理日期范围
    const processed: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(values)) {
      if (Array.isArray(value) && value.length === 2 && dayjs.isDayjs(value[0])) {
        processed[`${key}_start`] = value[0].format('YYYY-MM-DD');
        processed[`${key}_end`] = value[1].format('YYYY-MM-DD');
      } else if (dayjs.isDayjs(value)) {
        processed[key] = value.format('YYYY-MM-DD');
      } else {
        processed[key] = value;
      }
    }
    onSearch(processed);
  };

  const handleReset = () => {
    form.resetFields();
    onReset?.();
  };

  const renderField = (field: SearchField) => {
    const placeholder = field.placeholder || `请输入${field.label}`;

    switch (field.type) {
      case 'select':
        return (
          <Select
            placeholder={placeholder}
            options={field.options}
            allowClear
            style={{ width: '100%' }}
          />
        );
      case 'date':
        return <DatePicker style={{ width: '100%' }} placeholder={placeholder} />;
      case 'daterange':
        return <RangePicker style={{ width: '100%' }} />;
      case 'number':
        return <Input type="number" placeholder={placeholder} allowClear />;
      default:
        return <Input placeholder={placeholder} allowClear />;
    }
  };

  if (internalFields.length === 0) return null;

  return (
    <Form form={form} layout="inline" style={{ marginBottom: 16 }}>
      <Row gutter={[16, 16]} style={{ width: '100%', rowGap: 12 }}>
        {internalFields.map((field) => (
          <Col key={field.name} span={field.width || colSpan}>
            <Form.Item
              name={field.name}
              label={field.label}
              initialValue={field.initialValue}
              style={{ marginBottom: 0, width: '100%' }}
            >
              {renderField(field)}
            </Form.Item>
          </Col>
        ))}
        <Col flex="auto" style={{ textAlign: 'right' }}>
          <Space>
            <Button type="primary" icon={<SearchOutlined />} loading={loading} onClick={handleSearch}>
              搜索
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              重置
            </Button>
          </Space>
        </Col>
      </Row>
    </Form>
  );
};

export default SearchForm;
