import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Input, Select, Space, Statistic, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { getSearchResources, runSearchQuery } from '@/api/search';
import type {
  SearchAggRow, SearchField, SearchGroup, SearchResource, SearchResult, SearchSort,
} from '@/types/api';
import FilterBuilder, { emptyCond } from './FilterBuilder';
import './search.css';

function safeHeadline(html: string): string {
  return html
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/&lt;b&gt;/g, '<b>')
    .replace(/&lt;\/b&gt;/g, '</b>');
}

const emptyGroup = (fields: SearchField[]): SearchGroup => ({
  logic: 'and',
  conditions: fields.some((f) => f.filterable) ? [emptyCond(fields)] : [],
});

const SearchPage: React.FC = () => {
  const [resources, setResources] = useState<SearchResource[]>([]);
  const [code, setCode] = useState('tasks');
  const [keyword, setKeyword] = useState('');
  const [filters, setFilters] = useState<SearchGroup>({ logic: 'and', conditions: [] });
  const [sortField, setSortField] = useState<string>();
  const [aggField, setAggField] = useState<string>();
  const [timeField, setTimeField] = useState<string>();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SearchResult | null>(null);

  const current = resources.find((r) => r.code === code);
  const fields = useMemo(
    () => resources.find((r) => r.code === code)?.fields || [],
    [resources, code],
  );
  const sortable = fields.filter((f) => f.sortable);
  const aggFields = fields.filter((f) => f.agg);
  const timeFields = fields.filter((f) => f.type === 'time' && f.agg);

  const loadMeta = useCallback(async () => {
    const list = await getSearchResources();
    setResources(list || []);
  }, []);

  useEffect(() => { void loadMeta(); }, [loadMeta]);
  useEffect(() => {
    const cur = resources.find((r) => r.code === code);
    if (!cur) return;
    setFilters(emptyGroup(cur.fields));
    setSortField(cur.fields.find((f) => f.sortable && f.name === 'created_at')?.name
      || cur.fields.find((f) => f.sortable)?.name);
    setAggField(cur.fields.find((f) => f.agg && f.name === 'status')?.name);
    setTimeField(cur.fields.find((f) => f.type === 'time' && f.agg)?.name);
    setResult(null);
  }, [code, resources]);

  const run = async (cursor?: string) => {
    if (!current) return;
    setLoading(true);
    try {
      const sorts: SearchSort[] = sortField ? [{ field: sortField, desc: true }] : [];
      const aggregations = [];
      if (aggField) aggregations.push({ name: 'by_field', field: aggField, fn: 'count' });
      if (timeField) aggregations.push({ name: 'by_day', field: timeField, fn: 'count', interval: 'day' });
      const res = await runSearchQuery({
        resource: code,
        keyword: keyword.trim() || undefined,
        filters,
        sorts,
        page: 1,
        page_size: 10,
        cursor,
        aggregations,
      });
      setResult(res);
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<Record<string, unknown>> = useMemo(() => {
    const cols: ColumnsType<Record<string, unknown>> = fields.slice(0, 8).map((f) => ({
      title: f.label,
      dataIndex: f.name,
      ellipsis: true,
      render: (v: unknown) => <span className="search-mono">{v == null ? '—' : String(v)}</span>,
    }));
    cols.push({
      title: '高亮',
      dataIndex: '_headline',
      render: (v: unknown) => (typeof v === 'string' && v
        ? <span className="search-highlight" dangerouslySetInnerHTML={{ __html: safeHeadline(v) }} />
        : '—'),
    });
    return cols;
  }, [fields]);

  const buckets = (name: string): SearchAggRow[] => result?.aggregations?.[name] || [];

  return (
    <div>
      <div className="search-hero">
        <div>
          <h2>统一搜索</h2>
          <p>全文检索、组合筛选、游标分页与时间聚合。权限 search:read。</p>
        </div>
        <Button icon={<ReloadOutlined />} onClick={() => void loadMeta()}>刷新资源</Button>
      </div>
      <div className="search-stats">
        <Card><Statistic title="命中" value={result?.total ?? 0} /></Card>
        <Card><Statistic title="本页" value={result?.list.length ?? 0} /></Card>
        <Card><Statistic title="耗时" value={result?.elapsed_ms ?? 0} suffix="ms" /></Card>
        <Card><Statistic title="资源" value={resources.length} /></Card>
      </div>
      <Card className="search-shell" title="查询条件">
        <Space wrap style={{ marginBottom: 12 }}>
          <Select
            style={{ width: 180 }}
            value={code}
            options={resources.map((r) => ({ value: r.code, label: r.name }))}
            onChange={setCode}
          />
          <Input
            style={{ width: 280 }}
            allowClear
            placeholder="关键词（中文分词 / 英文）"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={() => void run()}
          />
          <Select
            allowClear
            placeholder="排序"
            style={{ width: 160 }}
            value={sortField}
            options={sortable.map((f) => ({ value: f.name, label: f.label }))}
            onChange={setSortField}
          />
          <Select
            allowClear
            placeholder="分组统计"
            style={{ width: 160 }}
            value={aggField}
            options={aggFields.map((f) => ({ value: f.name, label: f.label }))}
            onChange={setAggField}
          />
          <Select
            allowClear
            placeholder="按日聚合"
            style={{ width: 160 }}
            value={timeField}
            options={timeFields.map((f) => ({ value: f.name, label: f.label }))}
            onChange={setTimeField}
          />
          <Button type="primary" icon={<SearchOutlined />} loading={loading} onClick={() => void run()}>
            搜索
          </Button>
        </Space>
        {current ? <FilterBuilder group={filters} fields={current.fields} onChange={setFilters} /> : null}
      </Card>
      <Card className="search-shell" title="结果">
        <Table
          rowKey={(row) => String(row.id ?? JSON.stringify(row))}
          loading={loading}
          columns={columns}
          dataSource={result?.list || []}
          pagination={false}
          size="small"
        />
        <Space style={{ marginTop: 12 }}>
          <Tag color={result?.has_more ? 'blue' : 'default'}>{result?.has_more ? '还有下一页' : '没有更多'}</Tag>
          <Button
            disabled={!result?.has_more || !result.next_cursor}
            onClick={() => void run(result?.next_cursor)}
          >
            游标下一页
          </Button>
        </Space>
      </Card>
      {buckets('by_field').length > 0 || buckets('by_day').length > 0 ? (
        <Card className="search-shell" title="聚合">
          <Space align="start" wrap>
            <Table
              size="small"
              pagination={false}
              rowKey={(_, i) => `f-${i}`}
              dataSource={buckets('by_field')}
              columns={[
                { title: '分组', dataIndex: 'key', render: (v: unknown) => String(v ?? '') },
                { title: 'COUNT', dataIndex: 'value' },
              ]}
            />
            <Table
              size="small"
              pagination={false}
              rowKey={(_, i) => `d-${i}`}
              dataSource={buckets('by_day')}
              columns={[
                { title: '日', dataIndex: 'key', render: (v: unknown) => String(v ?? '') },
                { title: 'COUNT', dataIndex: 'value' },
              ]}
            />
          </Space>
        </Card>
      ) : null}
    </div>
  );
};

export default SearchPage;
