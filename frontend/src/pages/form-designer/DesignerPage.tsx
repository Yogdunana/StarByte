import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Col, Drawer, Input, Row, Space, message } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { FormEngine, FIELD_TYPE_LABELS, type FormField, type FormFieldType } from '@/components/FormEngine';
import { createForm, getForm, updateForm } from '@/api/forms';
import Palette from './Palette';
import FieldList from './FieldList';
import PropPanel from './PropPanel';

function slug(type: FormFieldType, used: Set<string>): string {
  const base = type === 'text' ? 'field' : type;
  if (!used.has(base)) {
    return base;
  }
  let i = 2;
  while (used.has(`${base}_${i}`)) {
    i += 1;
  }
  return `${base}_${i}`;
}

const DesignerPage: React.FC = () => {
  const { id } = useParams();
  const nav = useNavigate();
  const [name, setName] = useState('未命名表单');
  const [description, setDescription] = useState('');
  const [fields, setFields] = useState<FormField[]>([]);
  const [selected, setSelected] = useState<string>();
  const [formId, setFormId] = useState<string | undefined>(id);
  const [saving, setSaving] = useState(false);
  const [preview, setPreview] = useState(false);

  useEffect(() => {
    if (!id) return;
    void getForm(id).then((f) => {
      setFormId(f.id);
      setName(f.name);
      setDescription(f.description);
      setFields(f.fields || []);
    });
  }, [id]);

  const selectedField = fields.find((f) => f.name === selected);

  const addField = useCallback((type: FormFieldType) => {
    const names = new Set(fields.map((f) => f.name));
    const nameKey = slug(type, names);
    const next: FormField = { name: nameKey, label: FIELD_TYPE_LABELS[type], type };
    setFields((prev) => [...prev, next]);
    setSelected(nameKey);
  }, [fields]);

  const patchField = (patch: Partial<FormField>) => {
    if (!selected) return;
    setFields((prev) => prev.map((f) => {
      if (f.name !== selected) return f;
      const next = { ...f, ...patch };
      if (patch.name && patch.name !== f.name) setSelected(patch.name);
      return next;
    }));
  };

  const move = (fname: string, dir: -1 | 1) => {
    setFields((prev) => {
      const idx = prev.findIndex((f) => f.name === fname);
      const j = idx + dir;
      if (idx < 0 || j < 0 || j >= prev.length) return prev;
      const copy = [...prev];
      const tmp = copy[idx];
      copy[idx] = copy[j];
      copy[j] = tmp;
      return copy;
    });
  };

  const save = async (publish?: boolean) => {
    setSaving(true);
    try {
      const body = { name, description, fields, status: publish ? 1 : undefined };
      if (formId) {
        const out = await updateForm(formId, body);
        setFormId(out.id);
      } else {
        const out = await createForm(body);
        setFormId(out.id);
        nav(`/forms/designer/${out.id}`, { replace: true });
      }
      message.success(publish ? '已保存并发布' : '已保存');
    } finally {
      setSaving(false);
    }
  };

  const exportJSON = () => {
    const blob = new Blob([JSON.stringify({ name, description, fields }, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name || 'form'}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Card
      title="表单设计器"
      extra={(
        <Space>
          <Button onClick={() => nav('/forms')}>返回列表</Button>
          <Button onClick={() => setPreview(true)}>预览</Button>
          <Button onClick={exportJSON}>导出 JSON</Button>
          <Button loading={saving} onClick={() => { void save(false); }}>保存草稿</Button>
          <Button type="primary" loading={saving} onClick={() => { void save(true); }}>保存并发布</Button>
        </Space>
      )}
    >
      <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="表单名称" maxLength={100} />
        <Input.TextArea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="说明" rows={2} maxLength={500} />
      </Space>
      <Row gutter={16}>
        <Col xs={24} md={5}><Palette onAdd={addField} /></Col>
        <Col xs={24} md={11}>
          <FieldList
            fields={fields}
            selected={selected}
            onSelect={setSelected}
            onRemove={(n) => setFields((prev) => prev.filter((f) => f.name !== n))}
            onMove={move}
            onDropType={(t) => addField(t as FormFieldType)}
          />
        </Col>
        <Col xs={24} md={8}>
          <PropPanel field={selectedField} allNames={fields.map((f) => f.name)} onChange={patchField} />
        </Col>
      </Row>
      <Drawer title="实时预览" open={preview} onClose={() => setPreview(false)} width={480}>
        <FormEngine schema={{ fields }} showSubmit={false} />
      </Drawer>
    </Card>
  );
};

export default DesignerPage;
