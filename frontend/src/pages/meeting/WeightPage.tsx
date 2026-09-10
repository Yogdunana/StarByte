import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, InputNumber, Space, message } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import PageIntro from '@/components/PageIntro/PageIntro';
import { usePermission } from '@/hooks/usePermission';
import { getVoteWeightConfig, updateVoteWeightConfig } from '@/api/meeting';
interface Fields { default_weight: number; weights: Array<{ code: string; weight: number }> }
const names: Record<string, string> = { president: '会长', vice_president: '副会长', minister: '部长', vice_minister: '副部长', officer: '干事', member: '会员', center_director: '中心主任' };
export default function WeightPage() {
  const canEdit = usePermission('system:config');
  const [form] = Form.useForm<Fields>();
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const cfg = await getVoteWeightConfig();
      const weights = { ...cfg.weights };
      if (weights.vice_minister === undefined && weights.deputy !== undefined) weights.vice_minister = weights.deputy;
      delete weights.deputy;
      form.setFieldsValue({ default_weight: cfg.default_weight, weights: Object.entries(weights).map(([code, weight]) => ({ code, weight })) });
      setFailed(false);
    } catch { setFailed(true); } finally { setLoading(false); }
  }, [form]);
  useEffect(() => { void load(); }, [load]);
  const save = async (values: Fields) => {
    const weights: Record<string, number> = {};
    for (const item of values.weights) {
      const code = item.code.trim() === 'deputy' ? 'vice_minister' : item.code.trim();
      if (code in weights) { message.error('同一职务或角色只能配置一次'); return; }
      weights[code] = item.weight;
    }
    if (weights.vice_minister !== undefined) weights.deputy = weights.vice_minister;
    setBusy(true);
    try { await updateVoteWeightConfig({ weights, default_weight: values.default_weight }); message.success('权重配置已保存，新投票开始时生效'); } catch { /* Preserve edits. */ } finally { setBusy(false); }
  };
  return <>
    <PageIntro eyebrow="VOTING / 计票设置" title="明确每一票的分量" description="新投票开始时固定名单与权重。多个已配置职务或角色取最高权重，每人仍只投一票。" />
    <Card loading={loading}>
      {failed ? <Alert type="error" showIcon message="权重配置暂不可用" action={<Button onClick={() => void load()}>重试读取</Button>} /> : <Form form={form} layout="vertical" disabled={!canEdit || busy} onFinish={values => void save(values)}>
        <Alert type="info" showIcon message="修改仅影响之后发起的投票" description="普通议题中，未绑定职务的参会人使用默认权重；已绑定但尚未配置权重的职务会阻止发起。历史投票如没有保存快照，需单独核对。" style={{ marginBottom: 24 }} />
        <Form.List name="weights">{(fields, { add, remove }) => <>
          {fields.map(field => <Space key={field.key} align="start" wrap style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item name={[field.name, 'code']} label="职务或角色编码" rules={[{ required: true, whitespace: true, message: '请填写组织设置中的编码' }]}><Input placeholder="例如 minister（部长）" style={{ width: 'min(240px,65vw)' }} suffix={names[form.getFieldValue(['weights', field.name, 'code']) as string]} /></Form.Item>
            <Form.Item name={[field.name, 'weight']} label="权重" rules={[{ required: true }, { type: 'number', min: 0.01, max: 999999.99 }]}><InputNumber min={0.01} max={999999.99} precision={2} style={{ width: 130 }} /></Form.Item>
            {canEdit && <Button aria-label={`移除第${field.name + 1}项权重`} icon={<DeleteOutlined />} onClick={() => remove(field.name)} style={{ marginTop: 30 }} />}
          </Space>)}
          {canEdit && <Button type="dashed" icon={<PlusOutlined />} onClick={() => add({ code: '', weight: 1 })} style={{ marginBottom: 24 }}>添加职务或角色</Button>}
        </>}</Form.List>
        <Form.Item name="default_weight" label="普通议题默认权重" rules={[{ required: true }, { type: 'number', min: 0.01, max: 999999.99 }]}><InputNumber min={0.01} max={999999.99} precision={2} /></Form.Item>
        {canEdit && <Button type="primary" htmlType="submit" loading={busy}>保存配置</Button>}
      </Form>}
    </Card>
  </>;
}
