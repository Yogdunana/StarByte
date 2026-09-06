import React, { useEffect, useMemo, useState } from 'react';
import { Card, Empty, Space, Tag, Tree, Typography } from 'antd';
import type { DataNode } from 'antd/es/tree';
import { ApartmentOutlined } from '@ant-design/icons';
import { getDepartmentTree } from '@/api/department';
import type { Department } from '@/types/api';
import './department.css';

const { Paragraph, Text } = Typography;

function isCenter(node: Department): boolean {
  return node.code.startsWith('center_');
}

function charterTree(nodes: Department[]): Department[] {
  return nodes.filter(isCenter).map((n) => ({
    ...n,
    children: (n.children || []).filter((c) => !isCenter(c)),
  }));
}

function flattenLeaves(nodes: Department[]): Department[] {
  const out: Department[] = [];
  const walk = (list: Department[]) => {
    list.forEach((n) => {
      if (n.children && n.children.length > 0) {
        walk(n.children);
      } else if (n.parent_id) {
        out.push(n);
      }
    });
  };
  walk(nodes);
  return out;
}

function toTreeData(nodes: Department[]): DataNode[] {
  return nodes.map((n) => ({
    key: n.id,
    title: (
      <Space size={8}>
        <span>{n.name}</span>
        <Tag color={isCenter(n) ? 'geekblue' : 'blue'}>{isCenter(n) ? '中心' : '部门'}</Tag>
      </Space>
    ),
    children: n.children?.length ? toTreeData(n.children) : undefined,
  }));
}

function findNode(nodes: Department[], id: string): Department | undefined {
  for (const n of nodes) {
    if (n.id === id) return n;
    if (n.children) {
      const hit = findNode(n.children, id);
      if (hit) return hit;
    }
  }
  return undefined;
}

function allKeys(nodes: Department[]): React.Key[] {
  const keys: React.Key[] = [];
  const walk = (list: Department[]) => {
    list.forEach((n) => {
      keys.push(n.id);
      if (n.children) walk(n.children);
    });
  };
  walk(nodes);
  return keys;
}

const DepartmentPage: React.FC = () => {
  const [tree, setTree] = useState<Department[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState<string>();
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    setLoading(true);
    void getDepartmentTree()
      .then((rows) => {
        const list = charterTree(rows || []);
        setTree(list);
        setExpandedKeys(allKeys(list));
        setSelectedId((cur) => cur ?? list[0]?.id);
      })
      .finally(() => setLoading(false));
  }, []);

  const selected = selectedId ? findNode(tree, selectedId) : undefined;
  const leaves = useMemo(() => flattenLeaves(tree), [tree]);
  const treeData = useMemo(() => toTreeData(tree), [tree]);

  return (
    <div>
      <div className="dept-hero">
        <div>
          <h2>组织架构</h2>
          <p>按计协章程铺三大中心、七大职能部门。入会、面试、任务下拉只出现职能部门，不选中心。</p>
        </div>
        <Tag color="gold">章程第十九条–二十二条</Tag>
      </div>
      <div className="dept-grid">
        <Card className="page-shell" title="中心与部门树" loading={loading}>
          {tree.length === 0 && !loading ? (
            <Empty description="暂无部门，请先执行 seed" />
          ) : (
            <Tree
              showLine
              expandedKeys={expandedKeys}
              onExpand={setExpandedKeys}
              selectedKeys={selectedId ? [selectedId] : []}
              onSelect={(keys) => {
                if (keys[0]) setSelectedId(String(keys[0]));
              }}
              treeData={treeData}
            />
          )}
        </Card>
        <Card className="page-shell" title="节点说明">
          {selected ? (
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <Text strong>{selected.name}</Text>
              <Space>
                <Tag>{selected.code}</Tag>
                <Tag color={isCenter(selected) ? 'geekblue' : 'blue'}>
                  {isCenter(selected) ? '三大中心' : '职能部门'}
                </Tag>
              </Space>
              <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                {selected.description || '暂无职能说明'}
              </Paragraph>
            </Space>
          ) : (
            <Empty description="点击左侧节点查看职能" />
          )}
        </Card>
      </div>
      <Card className="page-shell" title={`七大职能部门（${leaves.length}）`} style={{ marginTop: 16 }}>
        <div className="dept-leaf-list">
          {leaves.map((d) => (
            <button
              key={d.id}
              type="button"
              className={`dept-leaf${selectedId === d.id ? ' is-active' : ''}`}
              onClick={() => setSelectedId(d.id)}
            >
              <ApartmentOutlined />
              <span>{d.name}</span>
              <Text type="secondary">{d.code}</Text>
            </button>
          ))}
        </div>
      </Card>
    </div>
  );
};

export default DepartmentPage;
