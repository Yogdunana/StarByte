import React, { useState } from 'react';
import { Button, Popconfirm, message } from 'antd';

export interface ConfirmButtonProps {
  title: string;
  description?: string;
  onConfirm: () => Promise<void>;
  type?: 'default' | 'primary' | 'link' | 'dashed' | 'text';
  danger?: boolean;
  size?: 'small' | 'middle' | 'large';
  disabled?: boolean;
  /** 成功提示消息，设为空字符串则不提示 */
  successMessage?: string;
  /** 失败提示消息，设为空字符串则不提示 */
  errorMessage?: string;
  children: React.ReactNode;
}

/**
 * 确认按钮 — 带 Popconfirm 弹窗，支持异步操作和加载状态
 */
const ConfirmButton: React.FC<ConfirmButtonProps> = ({
  title,
  description,
  onConfirm,
  type = 'default',
  danger = false,
  size = 'middle',
  disabled = false,
  successMessage = '操作成功',
  errorMessage = '操作失败',
  children,
}) => {
  const [loading, setLoading] = useState(false);

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await onConfirm();
      if (successMessage) message.success(successMessage);
    } catch (error) {
      if (errorMessage) {
        const msg = error instanceof Error ? error.message : errorMessage;
        message.error(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Popconfirm
      title={title}
      description={description}
      onConfirm={handleConfirm}
      okText="确定"
      cancelText="取消"
      okButtonProps={{ danger }}
      disabled={disabled || loading}
    >
      <Button type={type} danger={danger} size={size} loading={loading} disabled={disabled}>
        {children}
      </Button>
    </Popconfirm>
  );
};

export default ConfirmButton;
