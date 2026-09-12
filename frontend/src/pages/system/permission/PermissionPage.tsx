import React, { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  TreeSelect,
  message,
} from 'antd';
import type { DataNode } from 'antd/es/tree';
import { useTranslation } from 'react-i18next';
import {
  createPermission,
  deletePermission,
  getPermissionTree,
  updatePermission,
  type PermissionInput,
} from '@/api/role';
import { usePermissions } from '@/hooks/usePermission';
import type { Permission } from '@/types/api';

const PermissionPage: React.FC = () => {
  const { t } = useTranslation();
  const [canCreate, canUpdate, canDelete] = usePermissions([
    'permission:create',
    'permission:update',
    'permission:delete',
  ]);
  const [rows, setRows] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<Permission | null | undefined>();
  const [form] = Form.useForm<PermissionInput>();
  const kind = Form.useWatch('type', form);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setRows(await getPermissionTree());
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load().catch(() => undefined);
  }, [load]);
  const edit = (row?: Permission) => {
    form.resetFields();
    form.setFieldsValue(row ? { ...row, parent_id: row.parent_id || undefined } : { type: 'api' });
    setEditing(row || null);
  };
  const save = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editing) await updatePermission(editing.id, values);
      else await createPermission(values);
      setEditing(undefined);
      await load();
      message.success(t('common.saveSuccess', '保存成功'));
    } finally {
      setSaving(false);
    }
  };
  const parentOptions = (items: Permission[]): (DataNode & { value: string })[] =>
    items.map((p) => ({
      key: p.id,
      title: p.name,
      value: p.id,
      children: parentOptions(p.children || []),
    }));
  return (
    <Card
      title={t('rbac.permissions', '权限管理')}
      extra={
        canCreate && (
          <Button type="primary" onClick={() => edit()}>
            {t('common.create', '新增')}
          </Button>
        )
      }
    >
      <Table<Permission>
        rowKey="id"
        dataSource={rows}
        loading={loading}
        pagination={false}
        columns={[
          { title: t('common.name', '名称'), dataIndex: 'name' },
          { title: t('rbac.code', '编码'), dataIndex: 'code' },
          {
            title: t('common.type', '类型'),
            dataIndex: 'type',
            render: (value: string) => t(`rbac.type.${value}`, value),
          },
          {
            title: t('common.status', '状态'),
            render: (_, row) => (
              <Tag>
                {row.status === 0 ? t('common.enabled', '启用') : t('common.disabled', '禁用')}
              </Tag>
            ),
          },
          {
            title: t('common.actions', '操作'),
            render: (_, row) => (
              <Space>
                {canUpdate && <Button onClick={() => edit(row)}>{t('common.edit', '编辑')}</Button>}
                {canDelete && !row.is_system && (
                  <Popconfirm
                    title={t('rbac.confirmDeletePermission', '确定删除该权限？')}
                    onConfirm={async () => {
                      await deletePermission(row.id);
                      await load();
                    }}
                  >
                    <Button danger>{t('common.delete', '删除')}</Button>
                  </Popconfirm>
                )}
              </Space>
            ),
          },
        ]}
      />
      <Modal
        title={
          editing ? t('rbac.editPermission', '编辑权限') : t('rbac.createPermission', '新增权限')
        }
        open={editing !== undefined}
        onCancel={() => setEditing(undefined)}
        onOk={() => void save().catch(() => undefined)}
        confirmLoading={saving}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item
            name="name"
            label={t('common.name', '名称')}
            rules={[{ required: true, whitespace: true, max: 100 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="code"
            label={t('rbac.code', '编码')}
            rules={[{ required: true, whitespace: true, max: 100 }]}
          >
            <Input disabled={!!editing} />
          </Form.Item>
          <Form.Item name="type" label={t('common.type', '类型')} rules={[{ required: true }]}>
            <Select
              disabled={!!editing}
              options={['menu', 'button', 'api'].map((value) => ({
                value,
                label: t(`rbac.type.${value}`, value),
              }))}
            />
          </Form.Item>
          {!editing && (
            <Form.Item name="parent_id" label={t('rbac.parentPermission', '父级权限')}>
              <TreeSelect allowClear treeData={parentOptions(rows)} />
            </Form.Item>
          )}
          <Form.Item
            name="description"
            label={t('common.description', '说明')}
            rules={[{ max: 255 }]}
          >
            <Input.TextArea />
          </Form.Item>
          {kind === 'menu' && (
            <Form.Item name="path" label={t('rbac.path', '页面路径')} rules={[{ max: 255 }]}>
              <Input />
            </Form.Item>
          )}
          {kind === 'api' && !editing && (
            <>
              <Form.Item name="api_method" label={t('rbac.method', '请求方法')}>
                <Select
                  allowClear
                  options={['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((value) => ({
                    value,
                    label: value,
                  }))}
                />
              </Form.Item>
              <Form.Item
                name="api_path"
                label={t('rbac.apiPath', '接口路径')}
                rules={[{ max: 255 }]}
              >
                <Input />
              </Form.Item>
            </>
          )}
          {editing && (
            <Form.Item name="status" label={t('common.status', '状态')}>
              <Select
                disabled={editing.is_system}
                options={[
                  { value: 0, label: t('common.enabled', '启用') },
                  { value: 1, label: t('common.disabled', '禁用') },
                ]}
              />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </Card>
  );
};
export default PermissionPage;
