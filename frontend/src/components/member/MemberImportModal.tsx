import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useState } from 'react';
import { Button, Input, Modal, Space, Steps, Table, Tag, Upload, message } from 'antd';
import { importMembers, previewMemberImport } from '@/api/member';
import type { MemberImportResult, MemberImportRowResult, MemberImportRowStatus } from '@/types/api';
import { exportCSV } from '@/utils/download';
import { MEMBER_IMPORT_TEMPLATE_HEADER, parseMemberImportDraft } from '@/utils/memberImport';

interface Props {
  open: boolean;
  onClose: () => void;
  onImported: () => void;
}

const STATUS_META: Record<MemberImportRowStatus, { label: () => string; color: string }> = {
  create: { label: () => tx('新建'), color: 'green' },
  update: { label: () => tx('更新'), color: 'blue' },
  skip: { label: () => tx('无变化'), color: 'default' },
  error: { label: () => tx('失败'), color: 'red' },
};

/**
 * 「批量录入」弹窗三步走：贴名单 → 看后端给的逐行预判 → 确认写入。
 *
 * 预检放在这里是刻意的：协会职务只能系统管理员录、部门名能不能对上、
 * 学号有没有重复，这些都得问过后端才知道，本地猜一遍会给出假的安心感。
 */
const MemberImportModal: React.FC<Props> = ({ open, onClose, onImported }) => {
  useLocale();
  const [step, setStep] = useState(0);
  const [text, setText] = useState('');
  const [result, setResult] = useState<MemberImportResult | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) {
      setStep(0);
      setText('');
      setResult(null);
      setBusy(false);
    }
  }, [open]);

  const readFile = (file: File) => {
    const reader = new FileReader();
    reader.onload = () => setText(String(reader.result ?? ''));
    reader.readAsText(file);
    return false;
  };

  const downloadTemplate = () => {
    exportCSV([MEMBER_IMPORT_TEMPLATE_HEADER], 'member-import-template');
  };

  const goPreview = useCallback(async () => {
    const rows = parseMemberImportDraft(text);
    if (rows.length === 0) {
      message.warning(tx('没能识别出任何一行，请对照模板检查表头'));
      return;
    }
    setBusy(true);
    try {
      const preview = await previewMemberImport(rows);
      setResult(preview);
      setStep(1);
    } finally {
      setBusy(false);
    }
  }, [text]);

  const goImport = async () => {
    const rows = parseMemberImportDraft(text);
    if (rows.length === 0) return;
    setBusy(true);
    try {
      const imported = await importMembers(rows);
      setResult(imported);
      setStep(2);
      onImported();
    } finally {
      setBusy(false);
    }
  };

  const columns = [
    { title: tx('行号'), dataIndex: 'row', width: 64 },
    { title: tx('学号'), dataIndex: 'student_no' },
    { title: tx('姓名'), dataIndex: 'real_name' },
    { title: tx('部门'), dataIndex: 'department' },
    { title: tx('角色'), dataIndex: 'role' },
    {
      title: tx('结果'),
      dataIndex: 'status',
      render: (status: MemberImportRowStatus) => (
        <Tag color={STATUS_META[status].color}>{STATUS_META[status].label()}</Tag>
      ),
    },
    { title: tx('说明'), dataIndex: 'message' },
  ];

  const summary = result
    ? tx('新建 {{value0}} 人，更新 {{value1}} 人，无变化 {{value2}} 人，失败 {{value3}} 人', {
        value0: result.created,
        value1: result.updated,
        value2: result.skipped,
        value3: result.failed,
      })
    : '';

  return (
    <Modal
      title={tx('批量录入成员')}
      open={open}
      width={860}
      onCancel={onClose}
      maskClosable={!busy}
      footer={
        step === 0 ? (
          <Space>
            <Button onClick={downloadTemplate}>{tx('下载模板')}</Button>
            <Button onClick={onClose}>{tx('取消')}</Button>
            <Button type="primary" loading={busy} onClick={() => void goPreview()}>
              {tx('下一步')}
            </Button>
          </Space>
        ) : step === 1 ? (
          <Space>
            <Button onClick={() => setStep(0)}>{tx('返回修改')}</Button>
            <Button type="primary" loading={busy} onClick={() => void goImport()}>
              {tx('确认录入')}
            </Button>
          </Space>
        ) : (
          <Button type="primary" onClick={onClose}>
            {tx('完成')}
          </Button>
        )
      }
    >
      <Steps
        size="small"
        current={step}
        style={{ marginBottom: 16 }}
        items={[
          { title: tx('粘贴名单') },
          { title: tx('确认预判结果') },
          { title: tx('录入完成') },
        ]}
      />
      {step === 0 && (
        <>
          <Space style={{ marginBottom: 12 }} wrap>
            <Upload accept=".csv,.txt,.tsv" beforeUpload={readFile} showUploadList={false}>
              <Button>{tx('选择 CSV 文件')}</Button>
            </Upload>
            <Button onClick={downloadTemplate}>{tx('下载模板')}</Button>
          </Space>
          <Input.TextArea
            rows={10}
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder={tx('也可以直接把表格里框选的内容粘到这里（支持 CSV / Tab 分隔）')}
          />
        </>
      )}
      {step > 0 && result && (
        <>
          <p style={{ marginBottom: 12 }}>{summary}</p>
          <Table<MemberImportRowResult>
            rowKey="row"
            size="small"
            dataSource={result.results}
            columns={columns}
            pagination={{ pageSize: 8 }}
          />
          {step === 2 && (
            <p style={{ marginTop: 12, color: '#666' }}>
              {tx('录进来的人第一次用统一身份认证登录时会自动绑到自己的账号，把登录地址发给他们就行。')}
            </p>
          )}
        </>
      )}
    </Modal>
  );
};

export default MemberImportModal;
