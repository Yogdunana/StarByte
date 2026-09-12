import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Modal, Popconfirm, Select, Space, Table } from 'antd';
import { useDispatch } from 'react-redux';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';
import { useTranslation } from 'react-i18next';
import { addRoleMember, getRoleMembers, removeRoleMember, type RoleMember } from '@/api/role';
import { getUserList } from '@/api/user';
import { usePermissions } from '@/hooks/usePermission';
import type { Role } from '@/types/api';

export default function RoleMembers({ role, onClose }: { role: Role; onClose: () => void }) {
  const { t } = useTranslation();
  const dispatch = useDispatch<AppDispatch>();
  const [canAssign, canReadUsers] = usePermissions(['role:assign', 'user:read']);
  const mutable = canAssign && !role.is_system;
  const [rows, setRows] = useState<RoleMember[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [busy, setBusy] = useState(false);
  const [userId, setUserId] = useState<string>();
  const [options, setOptions] = useState<{ value: string; label: string }[]>([]);
  const requestNumber = useRef(0);
  const load = useCallback(async () => {
    const data = await getRoleMembers(role.id, page);
    setRows(data.list);
    setTotal(data.total);
  }, [role.id, page]);
  useEffect(() => {
    void load().catch(() => undefined);
  }, [load]);
  const search = async (keyword: string) => {
    const number = ++requestNumber.current;
    const result = await getUserList({ page: 1, page_size: 20, keyword, status: 0 });
    if (number === requestNumber.current)
      setOptions(
        result.list.map((user) => ({
          value: user.id,
          label: `${user.real_name} (${user.username})`,
        })),
      );
  };
  const add = async () => {
    if (!userId) return;
    setBusy(true);
    try {
      await addRoleMember(role.id, userId);
      void dispatch(fetchCurrentUser());
      setUserId(undefined);
      await load();
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal
      open
      title={`${t('rbac.members', '角色成员')} · ${role.name}`}
      onCancel={onClose}
      footer={null}
      width={720}
    >
      {mutable && canReadUsers && (
        <Space style={{ marginBottom: 16 }}>
          <Select
            aria-label={t('rbac.selectUser', '搜索并选择用户')}
            placeholder={t('rbac.selectUser', '搜索并选择用户')}
            showSearch
            filterOption={false}
            options={options}
            value={userId}
            onChange={setUserId}
            onSearch={(value) => void search(value).catch(() => undefined)}
            onFocus={() => void search('').catch(() => undefined)}
            style={{ minWidth: 260 }}
          />
          <Button
            loading={busy}
            disabled={!userId}
            onClick={() => void add().catch(() => undefined)}
          >
            {t('rbac.addMember', '添加成员')}
          </Button>
        </Space>
      )}
      <Table<RoleMember>
        rowKey="id"
        dataSource={rows}
        pagination={{
          current: page,
          total,
          pageSize: 20,
          onChange: setPage,
          showSizeChanger: false,
        }}
        columns={[
          { title: t('profile.name', '姓名'), dataIndex: 'real_name' },
          { title: t('profile.username', '用户名'), dataIndex: 'username' },
          {
            title: t('common.actions', '操作'),
            render: (_, user) =>
              mutable && (
                <Popconfirm
                  title={t('rbac.confirmRemoveMember', '确定移除该角色成员？')}
                  onConfirm={async () => {
                    await removeRoleMember(role.id, user.id);
                    void dispatch(fetchCurrentUser());
                    await load();
                  }}
                >
                  <Button danger>{t('common.remove', '移除')}</Button>
                </Popconfirm>
              ),
          },
        ]}
      />
    </Modal>
  );
}
