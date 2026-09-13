import { roleDisplayName } from '@/utils/roleDisplayName';
import RoleMembers from './RoleMembers';
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
  Tree,
  message,
} from 'antd';
import type { DataNode } from 'antd/es/tree';
import { useTranslation } from 'react-i18next';
import { useDispatch } from 'react-redux';
import {
  assignRolePermissions,
  createRole,
  deleteRole,
  getPermissionTree,
  getRoleDetail,
  getRoleList,
  updateRole,
} from '@/api/role';
import { usePermissions } from '@/hooks/usePermission';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';
import type { CreateRoleParams, Permission, Role, UpdateRoleParams } from '@/types/api';

const permissionNodes = (items: Permission[]): DataNode[] =>
  items.map((p) => ({
    key: p.id,
    title: `${p.name} (${p.code})`,
    children: permissionNodes(p.children || []),
  }));

const RolePage: React.FC = () => {
  const { t } = useTranslation();
  const dispatch = useDispatch<AppDispatch>();
  const [canCreate, canUpdate, canDelete, canAssign, canReadPermissions] = usePermissions([
    'role:create',
    'role:update',
    'role:delete',
    'role:assign',
    'permission:read',
  ]);
  const [rows, setRows] = useState<Role[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<Role | null | undefined>();
  const [members, setMembers] = useState<Role>();
  const [assigning, setAssigning] = useState<Role>();
  const [tree, setTree] = useState<DataNode[]>([]);
  const [checked, setChecked] = useState<React.Key[]>([]);
  const [form] = Form.useForm<CreateRoleParams & UpdateRoleParams>();
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await getRoleList({ page, page_size: 20, keyword });
      setRows(result.list);
      setTotal(result.total);
    } finally {
      setLoading(false);
    }
  }, [page, keyword]);
  useEffect(() => {
    void load().catch(() => undefined);
  }, [load]);
  const edit = async (row?: Role) => {
    const detail = row ? await getRoleDetail(row.id) : null;
    form.resetFields();
    form.setFieldsValue(detail ? { ...detail, parent_id: detail.parent_id || undefined } : {});
    setEditing(detail);
  };
  const save = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editing)
        await updateRole(
          editing.id,
          editing.is_system ? { name: values.name, description: values.description } : values,
        );
      else await createRole(values as CreateRoleParams);
      setEditing(undefined);
      await load();
      void dispatch(fetchCurrentUser());
      message.success(t('common.saveSuccess', '保存成功'));
    } finally {
      setSaving(false);
    }
  };
  const assign = async (row: Role) => {
    const [detail, permissions] = await Promise.all([getRoleDetail(row.id), getPermissionTree()]);
    setTree(permissionNodes(permissions));
    setChecked(detail.permission_ids || []);
    setAssigning(detail);
  };
  const savePermissions = async () => {
    if (!assigning) return;
    setSaving(true);
    try {
      await assignRolePermissions(assigning.id, checked.map(String));
      setAssigning(undefined);
      void dispatch(fetchCurrentUser());
      message.success(t('common.saveSuccess', '保存成功'));
    } finally {
      setSaving(false);
    }
  };
  return (
    <Card
      title={t('rbac.roles', '角色管理')}
      extra={
        canCreate && (
          <Button type="primary" onClick={() => void edit().catch(() => undefined)}>
            {t('common.create', '新增')}
          </Button>
        )
      }
    >
      <Input.Search
        aria-label={t('rbac.searchRoles', '搜索角色')}
        placeholder={t('rbac.searchRoles', '搜索角色')}
        allowClear
        onSearch={(v) => {
          setKeyword(v);
          setPage(1);
        }}
        style={{ maxWidth: 360, marginBottom: 16 }}
      />
      <Table<Role>
        rowKey="id"
        scroll={{ x: 960 }}
        dataSource={rows}
        loading={loading}
        pagination={{
          current: page,
          pageSize: 20,
          total,
          onChange: setPage,
          showSizeChanger: false,
        }}
        columns={[
          {
            title: t('common.name', '名称'),
            width: 180,
            render: (_, role) => roleDisplayName(role),
          },
          { title: t('rbac.code', '编码'), dataIndex: 'code', width: 220 },
          {
            title: t('common.status', '状态'),
            width: 160,
            render: (_, row) => (
              <Space wrap>
                <Tag>
                  {row.status === 0 ? t('common.enabled', '启用') : t('common.disabled', '禁用')}
                </Tag>
                {row.is_system && <Tag>{t('rbac.system', '系统内置')}</Tag>}
              </Space>
            ),
          },
          {
            title: t('common.actions', '操作'),
            width: 400,
            render: (_, row) => (
              <Space wrap>
                <Button onClick={() => setMembers(row)}>{t('rbac.members', '角色成员')}</Button>
                {canUpdate && (
                  <Button onClick={() => void edit(row).catch(() => undefined)}>
                    {t('common.edit', '编辑')}
                  </Button>
                )}
                {canAssign && canReadPermissions && (
                  <Button
                    disabled={row.is_system}
                    onClick={() => void assign(row).catch(() => undefined)}
                  >
                    {t('rbac.assignPermissions', '分配权限')}
                  </Button>
                )}
                {canDelete && !row.is_system && (
                  <Popconfirm
                    title={t('rbac.confirmDeleteRole', '确定删除该角色？')}
                    onConfirm={async () => {
                      await deleteRole(row.id);
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
      {members && (
        <RoleMembers key={members.id} role={members} onClose={() => setMembers(undefined)} />
      )}
      <Modal
        title={editing ? t('rbac.editRole', '编辑角色') : t('rbac.createRole', '新增角色')}
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
            rules={[{ required: true, whitespace: true, max: 50 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="code"
            label={t('rbac.code', '编码')}
            rules={[{ required: true, whitespace: true, max: 50 }]}
          >
            <Input disabled={editing?.is_system} />
          </Form.Item>
          <Form.Item
            name="description"
            label={t('common.description', '说明')}
            rules={[{ max: 255 }]}
          >
            <Input.TextArea />
          </Form.Item>
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
      <Modal
        title={`${t('rbac.assignPermissions', '分配权限')} · ${assigning?.name || ''}`}
        open={!!assigning}
        onCancel={() => setAssigning(undefined)}
        onOk={() => void savePermissions().catch(() => undefined)}
        confirmLoading={saving}
      >
        <Tree
          checkable
          checkStrictly
          checkedKeys={checked}
          treeData={tree}
          onCheck={(keys) => setChecked(Array.isArray(keys) ? keys : keys.checked)}
        />
      </Modal>
    </Card>
  );
};
export default RolePage;
