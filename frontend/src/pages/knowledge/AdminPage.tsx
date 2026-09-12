import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button, Drawer, Form, Input, Modal, Select, Space, Tree, Typography, message,
} from 'antd';
import type { DataNode } from 'antd/es/tree';
import { useTranslation } from 'react-i18next';
import PageContainer from '@/components/PageContainer';
import { usePermission } from '@/hooks/usePermission';
import { renderMarkdown } from '@/utils/markdown';
import {
  createKnowledgeDoc,
  getKnowledgeDoc,
  getKnowledgeHistory,
  getKnowledgeTree,
  getKnowledgeVersion,
  listKnowledgeCategories,
  publishKnowledgeDoc,
  rollbackKnowledgeDoc,
  searchKnowledge,
  updateKnowledgeDoc,
} from '@/api/knowledge';
import type { KnowledgeCategory, KnowledgeDoc, KnowledgeTreeNode, KnowledgeVersion } from '@/api/knowledge';
import './knowledge.css';

function toTree(nodes: KnowledgeTreeNode[]): DataNode[] {
  return nodes.map((n) => ({
    key: n.id,
    title: n.title,
    isLeaf: n.type === 'doc',
    children: n.children ? toTree(n.children) : undefined,
  }));
}

function flattenCats(rows: KnowledgeCategory[], prefix = ''): { value: string; label: string }[] {
  const out: { value: string; label: string }[] = [];
  for (const row of rows) {
    const label = prefix ? `${prefix} / ${row.name}` : row.name;
    out.push({ value: row.id, label });
    if (row.children) out.push(...flattenCats(row.children, label));
  }
  return out;
}

function collectExpandKeys(nodes: KnowledgeTreeNode[]): React.Key[] {
  const keys: React.Key[] = [];
  for (const node of nodes) {
    if (node.children?.length) {
      keys.push(node.id);
      keys.push(...collectExpandKeys(node.children));
    }
  }
  return keys;
}

const AdminPage: React.FC = () => {
  const { t } = useTranslation();
  const canCreate = usePermission('doc:create');
  const canUpdate = usePermission('doc:update');
  const canPublish = usePermission('doc:publish');
  const [form] = Form.useForm();
  const [tree, setTree] = useState<KnowledgeTreeNode[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [cats, setCats] = useState<KnowledgeCategory[]>([]);
  const [current, setCurrent] = useState<KnowledgeDoc | null>(null);
  const [content, setContent] = useState('');
  const [versions, setVersions] = useState<KnowledgeVersion[]>([]);
  const [histOpen, setHistOpen] = useState(false);
  const [diff, setDiff] = useState<{ left: string; right: string; title: string } | null>(null);

  const reloadTree = useCallback(async () => {
    const [nodes, categories] = await Promise.all([getKnowledgeTree(), listKnowledgeCategories()]);
    const next = nodes || [];
    setTree(next);
    setExpandedKeys(collectExpandKeys(next));
    setCats(categories || []);
  }, []);

  useEffect(() => { void reloadTree(); }, [reloadTree]);

  const loadDoc = async (id: string) => {
    const doc = await getKnowledgeDoc(id);
    setCurrent(doc);
    setContent(doc.content || '');
    form.setFieldsValue({
      title: doc.title,
      slug: doc.slug,
      kind: doc.kind,
      visibility: doc.visibility,
      permission_code: doc.permission_code,
      summary: doc.summary,
      category_id: doc.category_id,
    });
  };

  const onSelect = (keys: React.Key[]) => {
    const id = String(keys[0] || '');
    const isDoc = treeHasDoc(tree, id);
    if (isDoc) void loadDoc(id);
  };

  const save = async () => {
    const values = await form.validateFields();
    if (!current) {
      const created = await createKnowledgeDoc({ ...values, content });
      message.success(t('common.saved'));
      await reloadTree();
      await loadDoc(created.id);
      return;
    }
    await updateKnowledgeDoc(current.id, { ...values, content });
    message.success(t('common.saved'));
    await reloadTree();
    await loadDoc(current.id);
  };

  const openHistory = async () => {
    if (!current) return;
    setVersions(await getKnowledgeHistory(current.id));
    setHistOpen(true);
  };

  const preview = useMemo(() => renderMarkdown(content), [content]);

  return (
    <PageContainer title={t('knowledge.title')}>
      <div className="kb-workspace">
        <aside className="kb-side">
          <Input.Search
            placeholder={t('knowledge.search')}
            onSearch={async (q) => {
              if (!q.trim()) {
                await reloadTree();
                return;
              }
              const res = await searchKnowledge(q.trim());
              const hits = (res.list || []).map((d) => ({
                id: d.id, type: 'doc' as const, title: d.title, slug: d.slug, path: d.path, kind: d.kind,
              }));
              setTree(hits);
              setExpandedKeys([]);
            }}
          />
          <Tree
            style={{ marginTop: 12 }}
            treeData={toTree(tree)}
            expandedKeys={expandedKeys}
            onExpand={setExpandedKeys}
            onSelect={onSelect}
          />
          {canCreate && (
            <Button
              block
              style={{ marginTop: 16 }}
              onClick={() => {
                setCurrent(null);
                setContent('');
                form.resetFields();
                form.setFieldsValue({ kind: 'doc', visibility: 'authenticated' });
              }}
            >
              {t('knowledge.create')}
            </Button>
          )}
        </aside>
        <section className="kb-main">
          <Form form={form} layout="vertical">
            <Space wrap style={{ width: '100%' }}>
              <Form.Item name="title" label={t('knowledge.docTitle')} rules={[{ required: true }]} style={{ minWidth: 220 }}>
                <Input />
              </Form.Item>
              <Form.Item name="slug" label="Slug" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="kind" label={t('knowledge.kind')} rules={[{ required: true }]}>
                <Select options={[{ value: 'page', label: t('knowledge.kindPage') }, { value: 'doc', label: t('knowledge.kindDoc') }]} style={{ width: 140 }} />
              </Form.Item>
              <Form.Item name="visibility" label={t('knowledge.visibility')} rules={[{ required: true }]}>
                <Select
                  style={{ width: 160 }}
                  options={[
                    { value: 'public', label: t('knowledge.visPublic') },
                    { value: 'authenticated', label: t('knowledge.visAuth') },
                    { value: 'permission', label: t('knowledge.visPerm') },
                  ]}
                />
              </Form.Item>
            </Space>
            <Form.Item name="permission_code" label={t('knowledge.permCode')}>
              <Input placeholder="doc:publish" />
            </Form.Item>
            <Form.Item name="category_id" label={t('knowledge.category')}>
              <Select allowClear options={flattenCats(cats)} />
            </Form.Item>
            <Form.Item name="summary" label={t('knowledge.summary')}>
              <Input.TextArea rows={2} />
            </Form.Item>
          </Form>
          <div className="kb-editor">
            <Input.TextArea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={18}
              placeholder="Markdown"
            />
            <div className="kb-preview" dangerouslySetInnerHTML={{ __html: preview }} />
          </div>
          <Space style={{ marginTop: 16 }}>
            {canUpdate || canCreate ? <Button type="primary" onClick={() => void save()}>{t('knowledge.save')}</Button> : null}
            {canPublish && current && current.status === 0 && (
              <Button onClick={() => publishKnowledgeDoc(current.id).then(() => { message.success(t('knowledge.published')); void loadDoc(current.id); })}>
                {t('knowledge.publish')}
              </Button>
            )}
            {current && <Button onClick={() => void openHistory()}>{t('knowledge.versions')}</Button>}
            {current?.path && <Typography.Text type="secondary">{current.path} · v{current.version}</Typography.Text>}
          </Space>
        </section>
      </div>
      <Drawer title={t('knowledge.versions')} open={histOpen} onClose={() => setHistOpen(false)} width={480}>
        {versions.map((v) => (
          <div key={v.version} style={{ marginBottom: 12 }}>
            <Typography.Text strong>v{v.version}</Typography.Text>
            {v.is_current && <Typography.Text type="success"> {t('knowledge.current')}</Typography.Text>}
            <div>{v.title} · {v.editor?.name}</div>
            <Space>
              <Button size="small" onClick={async () => {
                if (!current) return;
                const full = await getKnowledgeVersion(current.id, v.version);
                setDiff({ title: `v${v.version}`, left: full.content || '', right: content });
              }}>{t('knowledge.diff')}</Button>
              {canUpdate && !v.is_current && (
                <Button size="small" onClick={() => {
                  if (!current) return;
                  rollbackKnowledgeDoc(current.id, v.version).then(() => {
                    message.success(t('knowledge.restored'));
                    setHistOpen(false);
                    void loadDoc(current.id);
                    void reloadTree();
                  });
                }}>{t('knowledge.restore')}</Button>
              )}
            </Space>
          </div>
        ))}
      </Drawer>
      <Modal title={diff?.title} open={Boolean(diff)} onCancel={() => setDiff(null)} footer={null} width={800}>
        {diff && (
          <div className="kb-diff">
            <pre>{diff.left}</pre>
            <pre>{diff.right}</pre>
          </div>
        )}
      </Modal>
    </PageContainer>
  );
};

function treeHasDoc(nodes: KnowledgeTreeNode[], id: string): boolean {
  for (const n of nodes) {
    if (n.id === id) return n.type === 'doc';
    if (n.children && treeHasDoc(n.children, id)) return true;
  }
  return false;
}

export default AdminPage;
