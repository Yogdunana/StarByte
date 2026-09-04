import React from 'react';
import { Button, Input, Select, Space } from 'antd';

export interface FileFilterValue {
  keyword: string;
  category?: string;
}

interface FileFilterBarProps {
  value: FileFilterValue;
  onChange: (value: FileFilterValue) => void;
  onSearch: () => void;
  onReset: () => void;
}

const FileFilterBar: React.FC<FileFilterBarProps> = ({ value, onChange, onSearch, onReset }) => {
  return (
    <Space wrap style={{ marginBottom: 16 }}>
      <Input
        allowClear
        placeholder="搜索文件名"
        value={value.keyword}
        style={{ width: 220 }}
        onChange={(e) => onChange({ ...value, keyword: e.target.value })}
        onPressEnter={onSearch}
      />
      <Select
        allowClear
        placeholder="分类"
        style={{ width: 140 }}
        value={value.category}
        onChange={(category) => onChange({ ...value, category })}
        options={[
          { value: 'image', label: '图片' },
          { value: 'document', label: '文档' },
          { value: 'video', label: '视频' },
        ]}
      />
      <Button type="primary" onClick={onSearch}>
        查询
      </Button>
      <Button onClick={onReset}>重置</Button>
    </Space>
  );
};

export default FileFilterBar;
