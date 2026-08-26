import React, { useState } from 'react';
import { Upload, message, Image } from 'antd';
import { UploadOutlined, PlusOutlined } from '@ant-design/icons';
import type { UploadFile, UploadProps } from 'antd/es/upload/interface';
import { getToken } from '@/utils/storage';

export interface FileUploadProps {
  accept?: string;
  maxSize?: number; // MB
  maxCount?: number;
  action?: string; // 上传 URL
  value?: UploadFile[];
  onChange?: (files: UploadFile[]) => void;
  listType?: 'text' | 'picture' | 'picture-card';
  multiple?: boolean;
}

/**
 * 文件上传组件 — 支持类型/大小/数量限制和预览
 */
const FileUpload: React.FC<FileUploadProps> = ({
  accept,
  maxSize = 10,
  maxCount = 1,
  action = '/api/v1/files/upload',
  value,
  onChange,
  listType = 'text',
  multiple = false,
}) => {
  const [fileList, setFileList] = useState<UploadFile[]>(value || []);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewImage, setPreviewImage] = useState('');

  const handleBeforeUpload: UploadProps['beforeUpload'] = (file) => {
    // 类型检查
    if (accept) {
      const types = accept.split(',').map((t) => t.trim());
      const fileType = file.type;
      const ext = '.' + file.name.split('.').pop()?.toLowerCase();
      const matched = types.some(
        (t) => fileType === t || ext === t || (t.startsWith('.') && ext === t) || (t.endsWith('/*') && fileType.startsWith(t.slice(0, -1))),
      );
      if (!matched) {
        message.error(`不支持的文件类型: ${ext || fileType}`);
        return Upload.LIST_IGNORE;
      }
    }

    // 大小检查
    if (file.size / 1024 / 1024 > maxSize) {
      message.error(`文件大小不能超过 ${maxSize}MB`);
      return Upload.LIST_IGNORE;
    }

    return true;
  };

  const handleChange: UploadProps['onChange'] = ({ fileList: newFileList }) => {
    setFileList(newFileList);
    onChange?.(newFileList);
  };

  const handlePreview = async (file: UploadFile) => {
    if (file.url) {
      setPreviewImage(file.url);
      setPreviewOpen(true);
    } else if (file.originFileObj) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setPreviewImage(e.target?.result as string);
        setPreviewOpen(true);
      };
      reader.readAsDataURL(file.originFileObj);
    }
  };

  const uploadProps: UploadProps = {
    name: 'file',
    action,
    accept,
    multiple,
    maxCount,
    listType,
    fileList,
    beforeUpload: handleBeforeUpload,
    onChange: handleChange,
    onPreview: handlePreview,
    headers: {
      Authorization: `Bearer ${getToken()}`,
    },
  };

  return (
    <>
      <Upload {...uploadProps}>
        {listType === 'picture-card' && fileList.length >= maxCount ? null : (
          listType === 'picture-card' ? (
            <div>
              <PlusOutlined />
              <div style={{ marginTop: 8 }}>上传</div>
            </div>
          ) : (
            <button style={{ border: '1px dashed #d9d9d9', padding: '4px 15px', borderRadius: 6, background: '#fff', cursor: 'pointer' }}>
              <UploadOutlined /> 点击上传
            </button>
          )
        )}
      </Upload>
      {previewOpen && (
        <Image
          style={{ display: 'none' }}
          preview={{
            visible: previewOpen,
            onVisibleChange: (visible) => setPreviewOpen(visible),
          }}
          src={previewImage}
        />
      )}
    </>
  );
};

export default FileUpload;
