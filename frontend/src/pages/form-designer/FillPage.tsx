import React, { useEffect, useState } from 'react';
import { Card, Result, Spin, message } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { FormEngine } from '@/components/FormEngine';
import { getForm, submitForm, type FormDetail } from '@/api/forms';

const FillPage: React.FC = () => {
  const { id } = useParams();
  const nav = useNavigate();
  const [form, setForm] = useState<FormDetail>();
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    void getForm(id).then((f) => setForm(f)).finally(() => setLoading(false));
  }, [id]);

  if (loading) return <Card><Spin /></Card>;
  if (!form) return <Result status="404" title="表单不存在" />;
  if (form.status !== 1) return <Result status="warning" title="表单未发布" extra={<a onClick={() => nav('/forms')}>返回列表</a>} />;

  return (
    <Card title={form.name} extra={<a onClick={() => nav('/forms')}>返回</a>}>
      {form.description ? <p style={{ color: '#666' }}>{form.description}</p> : null}
      <FormEngine
        schema={{ fields: form.fields || [] }}
        loading={submitting}
        onSubmit={async (values) => {
          if (!id) return;
          setSubmitting(true);
          try {
            await submitForm(id, values);
            message.success('提交成功');
            nav('/forms');
          } finally {
            setSubmitting(false);
          }
        }}
      />
    </Card>
  );
};

export default FillPage;
