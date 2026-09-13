import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, InputNumber, Space, message } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import PageIntro from '@/components/PageIntro/PageIntro';
import { usePermission } from '@/hooks/usePermission';
import { getVoteWeightConfig, updateVoteWeightConfig } from '@/api/meeting';
interface Fields {
  default_weight: number;
  weights: Array<{ code: string; weight: number }>;
}
const names: Record<string, string> = {
  get president() {
    return tx('会长');
  },
  get vice_president() {
    return tx('副会长');
  },
  get minister() {
    return tx('部长');
  },
  get vice_minister() {
    return tx('副部长');
  },
  get officer() {
    return tx('干事');
  },
  get member() {
    return tx('会员');
  },
  get center_director() {
    return tx('中心主任');
  },
};
export default function WeightPage() {
  useLocale();
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
      if (weights.vice_minister === undefined && weights.deputy !== undefined)
        weights.vice_minister = weights.deputy;
      delete weights.deputy;
      form.setFieldsValue({
        default_weight: cfg.default_weight,
        weights: Object.entries(weights).map(([code, weight]) => ({ code, weight })),
      });
      setFailed(false);
    } catch {
      setFailed(true);
    } finally {
      setLoading(false);
    }
  }, [form]);
  useEffect(() => {
    void load();
  }, [load]);
  const save = async (values: Fields) => {
    const weights: Record<string, number> = {};
    for (const item of values.weights) {
      const code = item.code.trim() === 'deputy' ? 'vice_minister' : item.code.trim();
      if (code in weights) {
        message.error(tx('同一职务或角色只能配置一次'));
        return;
      }
      weights[code] = item.weight;
    }
    if (weights.vice_minister !== undefined) weights.deputy = weights.vice_minister;
    setBusy(true);
    try {
      await updateVoteWeightConfig({ weights, default_weight: values.default_weight });
      message.success(tx('权重配置已保存，新投票开始时生效'));
    } catch {
      /* Preserve edits. */
    } finally {
      setBusy(false);
    }
  };
  return (
    <>
      <PageIntro
        eyebrow={tx('VOTING / 计票设置')}
        title={tx('明确每一票的分量')}
        description={tx(
          '新投票开始时固定名单与权重。多个已配置职务或角色取最高权重，每人仍只投一票。',
        )}
      />
      <Card loading={loading}>
        {failed ? (
          <Alert
            type="error"
            showIcon
            message={tx('权重配置暂不可用')}
            action={<Button onClick={() => void load()}>{tx('重试读取')}</Button>}
          />
        ) : (
          <Form
            form={form}
            layout="vertical"
            disabled={!canEdit || busy}
            onFinish={(values) => void save(values)}
          >
            <Alert
              type="info"
              showIcon
              message={tx('修改仅影响之后发起的投票')}
              description={tx(
                '普通议题中，未绑定职务的参会人使用默认权重；已绑定但尚未配置权重的职务会阻止发起。历史投票如没有保存快照，需单独核对。',
              )}
              style={{ marginBottom: 24 }}
            />
            <Form.List name="weights">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <Space
                      key={field.key}
                      align="start"
                      wrap
                      style={{ display: 'flex', marginBottom: 8 }}
                    >
                      <Form.Item
                        name={[field.name, 'code']}
                        label={tx('职务或角色编码')}
                        rules={[
                          {
                            required: true,
                            whitespace: true,
                            message: tx('请填写组织设置中的编码'),
                          },
                        ]}
                      >
                        <Input
                          placeholder={tx('例如 minister（部长）')}
                          style={{ width: 'min(240px,65vw)' }}
                          suffix={
                            names[form.getFieldValue(['weights', field.name, 'code']) as string]
                          }
                        />
                      </Form.Item>
                      <Form.Item
                        name={[field.name, 'weight']}
                        label={tx('权重')}
                        rules={[{ required: true }, { type: 'number', min: 0, max: 999999.99 }]}
                      >
                        <InputNumber min={0} max={999999.99} precision={2} style={{ width: 130 }} />
                      </Form.Item>
                      {canEdit && (
                        <Button
                          aria-label={tx('移除第{{value0}}项权重', { value0: field.name + 1 })}
                          icon={<DeleteOutlined />}
                          onClick={() => remove(field.name)}
                          style={{ marginTop: 30 }}
                        />
                      )}
                    </Space>
                  ))}
                  {canEdit && (
                    <Button
                      type="dashed"
                      icon={<PlusOutlined />}
                      onClick={() => add({ code: '', weight: 1 })}
                      style={{ marginBottom: 24 }}
                    >
                      {tx('添加职务或角色')}
                    </Button>
                  )}
                </>
              )}
            </Form.List>
            <Form.Item
              name="default_weight"
              label={tx('普通议题默认权重')}
              rules={[{ required: true }, { type: 'number', min: 0, max: 999999.99 }]}
            >
              <InputNumber min={0} max={999999.99} precision={2} />
            </Form.Item>
            {canEdit && (
              <Button type="primary" htmlType="submit" loading={busy}>
                {tx('保存配置')}
              </Button>
            )}
          </Form>
        )}
      </Card>
    </>
  );
}
