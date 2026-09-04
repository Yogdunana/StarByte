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
}) => (
  <div className="designer-toolbar">
    <Space>
      <strong>{title || '未命名流程'}</strong>
      {previewMode && <Tag color="blue">预览</Tag>}
    </Space>
    <Space wrap>
      <Button icon={<PlusOutlined />} onClick={onNew}>
        新建
      </Button>
      <Button icon={<FolderOpenOutlined />} onClick={onOpen}>
        打开
      </Button>
      <Button icon={<SaveOutlined />} loading={saving} onClick={onSaveDraft}>
        保存草稿
      </Button>
      <Button type="primary" icon={<CloudUploadOutlined />} loading={publishing} onClick={onPublish}>
        发布
      </Button>
      <Button icon={<EyeOutlined />} onClick={onTogglePreview}>
        {previewMode ? '返回编辑' : '预览'}
      </Button>
      <Button icon={<ImportOutlined />} disabled={previewMode} onClick={onImport}>
        导入
      </Button>
      <Button icon={<ExportOutlined />} onClick={onExport}>
        导出
      </Button>
      <Button icon={<UndoOutlined />} disabled={previewMode || !canUndo} onClick={onUndo}>
        撤销
      </Button>
      <Button icon={<RedoOutlined />} disabled={previewMode || !canRedo} onClick={onRedo}>
        重做
      </Button>
    </Space>
  </div>
);

export default DesignerToolbar;
