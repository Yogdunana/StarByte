import { roleDisplayName } from '@/utils/roleDisplayName';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Modal, Popconfirm, Select, Space, Table } from 'antd';
import { tx, useLocale } from '@/i18n/text';
import { getDepartmentTree } from '@/api/department';
import { selectRoles } from '@/store/slices/userSlice';
import { useDispatch, useSelector } from 'react-redux';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';
import { useTranslation } from 'react-i18next';
import { addRoleMember, getRoleMembers, removeRoleMember, type RoleMember } from '@/api/role';
import { getUserList } from '@/api/user';
import { usePermissions } from '@/hooks/usePermission';
import type { Role, Department } from '@/types/api';

export default function RoleMembers({ role, onClose }: { role: Role; onClose: () => void }) {
  useLocale();
  const { t } = useTranslation();
  const roles = useSelector(selectRoles);
  const leadership = ['president', 'vice_president', 'center_director', 'minister'].includes(
    role.code,
  );
  const scoped = leadership && role.code !== 'president';
  const [departmentIds, setDepartmentIds] = useState<string[]>([]);
  const [departments, setDepartments] = useState<{ value: string; label: string }[]>([]);
  useEffect(() => {
    if (!scoped || !roles.includes('super_admin')) return;
    void getDepartmentTree()
      .then((rows) => {
        const options: { value: string; label: string }[] = [];
        const walk = (items: Department[], parent?: string) =>
          items.forEach((d) => {
            if (role.code === 'minister' ? !!(d.parent_id || parent) : !(d.parent_id || parent))
              options.push({ value: d.id, label: d.name });
            if (d.children) walk(d.children, d.id);
          });
        walk(rows);
        setDepartments(options);
      })
      .catch(() => setDepartments([]));
  }, [scoped, role.code, roles]);
  const dispatch = useDispatch<AppDispatch>();
  const [canAssign, canReadUsers] = usePermissions(['role:assign', 'user:read']);
  const mutable = canAssign && (leadership ? roles.includes('super_admin') : !role.is_system);
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
      await addRoleMember(role.id, userId, departmentIds);
      void dispatch(fetchCurrentUser());
      setUserId(undefined);
      setDepartmentIds([]);
      await load();
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal
      open
      title={`${t('rbac.members', '角色成员')} · ${roleDisplayName(role)}`}
      onCancel={onClose}
      footer={null}
      width={720}
    >
      {mutable && canReadUsers && (
        <Space wrap style={{ marginBottom: 16, width: '100%' }}>
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
            style={{ width: 'min(260px, calc(100vw - 96px))' }}
          />
          {scoped && (
            <Select
              mode="multiple"
              aria-label={tx('任职范围')}
              placeholder={tx('任职范围')}
              value={departmentIds}
              onChange={setDepartmentIds}
              options={departments}
              style={{ width: 'min(300px, calc(100vw - 96px))' }}
            />
          )}
          <Button
            loading={busy}
            disabled={
              !userId ||
              (['minister', 'center_director', 'vice_president'].includes(role.code) &&
                departmentIds.length === 0)
            }
            onClick={() => void add().catch(() => undefined)}
          >
            {t('rbac.addMember', '添加成员')}
          </Button>
        </Space>
      )}
      <Table<RoleMember>
        rowKey="id"
        scroll={{ x: 760 }}
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
            title: tx('任职范围'),
            render: (_, user) =>
              (user.department_ids ?? [])
                .map((id) => departments.find((d) => d.value === id)?.label || id)
                .join('、') || '—',
          },
          {
            title: t('common.actions', '操作'),
            render: (_, user) =>
              mutable && (
                <Space wrap>
                  {scoped && (
                    <Button
                      onClick={() => {
                        setOptions([{ value: user.id, label: user.real_name || user.username }]);
                        setUserId(user.id);
                        setDepartmentIds(user.department_ids ?? []);
                      }}
                    >
                      {tx('编辑任职范围')}
                    </Button>
                  )}
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
                </Space>
              ),
          },
        ]}
      />
    </Modal>
  );
}
