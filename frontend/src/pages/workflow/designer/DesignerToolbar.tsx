import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Button, Space, Tag } from 'antd';
import {
  SaveOutlined,
  CloudUploadOutlined,
  EyeOutlined,
  ImportOutlined,
  ExportOutlined,
  UndoOutlined,
  RedoOutlined,
  FolderOpenOutlined,
  PlusOutlined,
} from '@ant-design/icons';

interface DesignerToolbarProps {
  title: string;
  previewMode: boolean;
  saving: boolean;
  publishing: boolean;
  canUndo: boolean;
  canRedo: boolean;
  onSaveDraft: () => void;
  onPublish: () => void;
  onTogglePreview: () => void;
  onImport: () => void;
  onExport: () => void;
  onUndo: () => void;
  onRedo: () => void;
  onOpen: () => void;
  onNew: () => void;
}

const DesignerToolbar: React.FC<DesignerToolbarProps> = ({
  title,
  previewMode,
  saving,
  publishing,
  canUndo,
  canRedo,
  onSaveDraft,
  onPublish,
  onTogglePreview,
  onImport,
  onExport,
  onUndo,
  onRedo,
  onOpen,
  onNew,
}) => {
  useLocale();
  return (
    <div className="designer-toolbar">
      <Space>
        <strong>{title || tx('未命名流程')}</strong>
        {previewMode && <Tag color="blue">{tx('预览')}</Tag>}
      </Space>
      <Space wrap>
        <Button icon={<PlusOutlined />} onClick={onNew}>
          {tx('新建')}
        </Button>
        <Button icon={<FolderOpenOutlined />} onClick={onOpen}>
          {tx('打开')}
        </Button>
        <Button icon={<SaveOutlined />} loading={saving} onClick={onSaveDraft}>
          {tx('保存草稿')}
        </Button>
        <Button
          type="primary"
          icon={<CloudUploadOutlined />}
          loading={publishing}
          onClick={onPublish}
        >
          {tx('发布')}
        </Button>
        <Button icon={<EyeOutlined />} onClick={onTogglePreview}>
          {previewMode ? tx('返回编辑') : tx('预览')}
        </Button>
        <Button icon={<ImportOutlined />} disabled={previewMode} onClick={onImport}>
          {tx('导入')}
        </Button>
        <Button icon={<ExportOutlined />} onClick={onExport}>
          {tx('导出')}
        </Button>
        <Button icon={<UndoOutlined />} disabled={previewMode || !canUndo} onClick={onUndo}>
          {tx('撤销')}
        </Button>
        <Button icon={<RedoOutlined />} disabled={previewMode || !canRedo} onClick={onRedo}>
          {tx('重做')}
        </Button>
      </Space>
    </div>
  );
};

export default DesignerToolbar;
