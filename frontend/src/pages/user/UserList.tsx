import { tx, useLocale } from '@/i18n/text';
import React, { useState, useEffect, useCallback } from 'react';
import { Table, Card, Button, Input, Select, Space, Modal, Form, message } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';

import { getUserList, UserListItem, createUser, updateUser, deleteUser } from '@/api/user';
import { getUserColumns } from './userColumns';

function isFormValidateError(error: unknown): boolean {
  return typeof error === 'object' && error !== null && 'errorFields' in error;
}

const UserList: React.FC = () => {
  useLocale();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<UserListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<number | undefined>();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<UserListItem | null>(null);
  const [form] = Form.useForm();

  const [filters, setFilters] = useState<{ keyword?: string; status?: number }>({});

  // 加载用户列表
  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getUserList({
        page,
        page_size: pageSize,
        ...filters,
      });
      setData(res.list);
      setTotal(res.total);
    } catch (error) {
      message.error(tx('加载用户列表失败'));
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, filters]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  // 搜索
  const handleSearch = () => {
    setPage(1);
    setFilters({ keyword, status });
  };

  // 重置
  const handleReset = () => {
    setKeyword('');
    setStatus(undefined);
    setPage(1);
    setFilters({});
  };

  // 新增
  const handleAdd = () => {
    setEditingUser(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 编辑
  const handleEdit = (record: UserListItem) => {
    setEditingUser(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  // 删除
  const handleDelete = (record: UserListItem) => {
    Modal.confirm({
      title: tx('确认删除'),
      content: tx('确定要删除用户 "{{value0}}" 吗？', { value0: record.username }),
      onOk: async () => {
        try {
          await deleteUser(record.id);
          message.success(tx('删除成功'));
          loadUsers();
        } catch (error) {
          message.error(tx('删除失败'));
        }
      },
    });
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingUser) {
        await updateUser(editingUser.id, values);
        message.success(tx('更新成功'));
      } else {
        await createUser(values);
        message.success(tx('创建成功'));
      }
      setModalVisible(false);
      loadUsers();
    } catch (error: unknown) {
      if (isFormValidateError(error)) return;
      message.error(editingUser ? tx('更新失败') : tx('创建失败'));
    }
  };

  const columns = getUserColumns({
    onEdit: handleEdit,
    onDelete: handleDelete,
  });

  return (
    <Card>
      {/* 搜索栏 */}
      <div style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
        <Input
          placeholder={tx('搜索用户名/姓名/邮箱')}
          prefix={<SearchOutlined />}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 240 }}
          onPressEnter={handleSearch}
        />
        <Select
          placeholder={tx('状态')}
          value={status}
          onChange={setStatus}
          style={{ width: 120 }}
          allowClear
        >
          <Select.Option value={0}>{tx('正常')}</Select.Option>
          <Select.Option value={1}>{tx('禁用')}</Select.Option>
          <Select.Option value={2}>{tx('锁定')}</Select.Option>
        </Select>
        <Space>
          <Button type="primary" onClick={handleSearch}>
            {tx('搜索')}
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            {tx('重置')}
          </Button>
        </Space>
        <div style={{ flex: 1 }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          {tx('新增用户')}
        </Button>
      </div>

      {/* 表格 */}
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1200 }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => tx('共 {{value0}} 条', { value0: total }),
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />

      {/* 新增/编辑弹窗 */}
      <Modal
        title={editingUser ? tx('编辑用户') : tx('新增用户')}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={500}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          {!editingUser && (
            <Form.Item
              name="username"
              label={tx('用户名')}
              rules={[
                { required: true, message: tx('请输入用户名') },
                { min: 3, max: 20, message: tx('用户名长度为3-20个字符') },
              ]}
            >
              <Input placeholder={tx('请输入用户名')} />
            </Form.Item>
          )}
          {!editingUser && (
            <Form.Item
              name="password"
              label={tx('初始密码')}
              rules={[
                { required: true, message: tx('请输入初始密码') },
                { min: 6, message: tx('密码至少6个字符') },
              ]}
            >
              <Input.Password placeholder={tx('请输入初始密码')} />
            </Form.Item>
          )}
          <Form.Item name="real_name" label={tx('真实姓名')} rules={[{ required: true }]}>
            <Input placeholder={tx('请输入真实姓名')} />
          </Form.Item>
          <Form.Item
            name="email"
            label={tx('邮箱')}
            rules={[{ type: 'email', message: tx('请输入有效邮箱') }]}
          >
            <Input placeholder={tx('请输入邮箱')} />
          </Form.Item>
          <Form.Item name="phone" label={tx('手机号')}>
            <Input placeholder={tx('请输入手机号')} />
          </Form.Item>
          <Form.Item name="gender" label={tx('性别')}>
            <Select>
              <Select.Option value={0}>{tx('未知')}</Select.Option>
              <Select.Option value={1}>{tx('男')}</Select.Option>
              <Select.Option value={2}>{tx('女')}</Select.Option>
            </Select>
          </Form.Item>
          {editingUser && (
            <Form.Item name="status" label={tx('状态')}>
              <Select>
                <Select.Option value={0}>{tx('正常')}</Select.Option>
                <Select.Option value={1}>{tx('禁用')}</Select.Option>
                <Select.Option value={2}>{tx('锁定')}</Select.Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>
    </Card>
  );
};

export default UserList;
