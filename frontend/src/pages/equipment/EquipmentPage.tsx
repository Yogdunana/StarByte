import { useState, useEffect, useCallback } from 'react';
import { Card, Table, Button, Modal, Form, Input, InputNumber, Select, Tag, Space, message, Tabs } from 'antd';
import { PlusOutlined, ToolOutlined, AuditOutlined } from '@ant-design/icons';
import {
  getEquipmentList, createEquipment, deleteEquipment,
  createBorrow, getBorrowList, approveBorrow, returnBorrow,
  createMaintenance, getMaintenanceList,
  createInventory, getInventoryList,
  type Equipment, type BorrowRecord,
} from '@/api/equipment';

const equipStatusColors: Record<number, string> = { 0: 'success', 1: 'warning', 2: 'error', 3: 'processing', 4: 'default' };
const borrowStatusColors: Record<number, string> = { 0: 'default', 1: 'processing', 2: 'warning', 3: 'success', 4: 'error', 5: 'error' };

export default function EquipmentPage() {
  const [activeTab, setActiveTab] = useState('list');
  const [equipment, setEquipment] = useState<Equipment[]>([]);
  const [equipTotal, setEquipTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [borrowModalOpen, setBorrowModalOpen] = useState(false);
  const [selectedEquip, setSelectedEquip] = useState<Equipment | null>(null);
  const [createForm] = Form.useForm();
  const [borrowForm] = Form.useForm();

  // 借用记录
  const [borrows, setBorrows] = useState<BorrowRecord[]>([]);
  const [borrowTotal, setBorrowTotal] = useState(0);
  const [borrowPage, setBorrowPage] = useState(1);

  const fetchEquipment = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getEquipmentList({ page, page_size: pageSize });
      setEquipment(res.data.list || []);
      setEquipTotal(res.data.total || 0);
    } catch {
      message.error('加载物资列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  const fetchBorrows = useCallback(async () => {
    try {
      const res = await getBorrowList({ page: borrowPage, page_size: pageSize });
      setBorrows(res.data.list || []);
      setBorrowTotal(res.data.total || 0);
    } catch {
      message.error('加载借用记录失败');
    }
  }, [borrowPage, pageSize]);

  useEffect(() => {
    if (activeTab === 'list') fetchEquipment();
    if (activeTab === 'borrows') fetchBorrows();
  }, [activeTab, fetchEquipment, fetchBorrows]);

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields();
      await createEquipment(values);
      message.success('物资入库成功');
      setCreateModalOpen(false);
      createForm.resetFields();
      fetchEquipment();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.message || '入库失败');
    }
  };

  const handleBorrow = async () => {
    try {
      const values = await borrowForm.validateFields();
      if (!selectedEquip) return;
      await createBorrow({
        equipment_id: selectedEquip.id,
        quantity: values.quantity,
        expected_return_at: values.expected_return_at.toISOString(),
        remark: values.remark,
      });
      message.success('借用申请已提交');
      setBorrowModalOpen(false);
      borrowForm.resetFields();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.message || '申请失败');
    }
  };

  const handleApprove = async (id: string, approved: boolean) => {
    try {
      await approveBorrow(id, { status: approved ? 1 : 4 });
      message.success(approved ? '已批准' : '已拒绝');
      fetchBorrows();
    } catch {
      message.error('操作失败');
    }
  };

  const handleReturn = async (id: string) => {
    try {
      await returnBorrow(id, {});
      message.success('归还确认成功');
      fetchBorrows();
    } catch {
      message.error('归还失败');
    }
  };

  const equipColumns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '型号', dataIndex: 'model', key: 'model' },
    { title: '分类', dataIndex: 'category_text', key: 'category_text' },
    { title: '可用/总量', key: 'qty', render: (_: unknown, r: Equipment) => `${r.available_quantity}/${r.total_quantity}` },
    { title: '存放位置', dataIndex: 'location', key: 'location' },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (s: number, r: Equipment) => <Tag color={equipStatusColors[s]}>{r.status_text}</Tag>,
    },
    {
      title: '操作', key: 'action',
      render: (_: unknown, record: Equipment) => (
        <Space>
          {record.status === 0 && record.available_quantity > 0 && (
            <Button size="small" onClick={() => { setSelectedEquip(record); setBorrowModalOpen(true); }}>
              借用
            </Button>
          )}
          <Button size="small" danger onClick={async () => {
            try { await deleteEquipment(record.id); message.success('删除成功'); fetchEquipment(); }
            catch { message.error('删除失败'); }
          }}>删除</Button>
        </Space>
      ),
    },
  ];

  const borrowColumns = [
    { title: '物资', dataIndex: 'equipment_name', key: 'equipment_name' },
    { title: '数量', dataIndex: 'quantity', key: 'quantity' },
    { title: '借用时间', dataIndex: 'borrow_at', key: 'borrow_at' },
    { title: '预计归还', dataIndex: 'expected_return_at', key: 'expected_return_at' },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (s: number, r: BorrowRecord) => <Tag color={borrowStatusColors[s]}>{r.status_text}</Tag>,
    },
    {
      title: '操作', key: 'action',
      render: (_: unknown, record: BorrowRecord) => (
        <Space>
          {record.status === 0 && (
            <>
              <Button size="small" type="primary" onClick={() => handleApprove(record.id, true)}>批准</Button>
              <Button size="small" danger onClick={() => handleApprove(record.id, false)}>拒绝</Button>
            </>
          )}
          {(record.status === 2 || record.status === 5) && (
            <Button size="small" onClick={() => handleReturn(record.id)}>归还确认</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card title="物资/设备管理">
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        { key: 'list', label: '物资列表', children: (
          <>
            <Button icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => setCreateModalOpen(true)}>入库</Button>
            <Table dataSource={equipment} columns={equipColumns} rowKey="id" loading={loading}
              pagination={{ current: page, pageSize, total: equipTotal, onChange: (p, ps) => { setPage(p); setPageSize(ps); } }} />
          </>
        )},
        { key: 'borrows', label: '借用记录', children: (
          <Table dataSource={borrows} columns={borrowColumns} rowKey="id"
            pagination={{ current: borrowPage, pageSize, total: borrowTotal, onChange: setBorrowPage }} />
        )},
      ]} />
    </Card>

    // 入库 Modal
    // 借用 Modal
  );
}
