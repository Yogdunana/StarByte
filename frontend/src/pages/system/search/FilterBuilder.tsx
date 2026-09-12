import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Button, Input, InputNumber, Select, Space } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import type { SearchCondition, SearchField, SearchGroup } from '@/types/api';

const logicOptions = [
  { value: 'and', label: 'AND' },
  { value: 'or', label: 'OR' },
];

function emptyCond(fields: SearchField[]): SearchCondition {
  const f = fields.find((x) => x.filterable);
  return { field: f?.name || '', operator: f?.operators[0] || 'eq', value: '' };
}

function parseIn(raw: string): Array<string | number> {
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .map((s) => {
      const n = Number(s);
      return Number.isFinite(n) && s !== '' && !Number.isNaN(n) && /^-?\d+(\.\d+)?$/.test(s)
        ? n
        : s;
    });
}

interface Props {
  group: SearchGroup;
  fields: SearchField[];
  onChange: (g: SearchGroup) => void;
  nested?: boolean;
}

const FilterBuilder: React.FC<Props> = ({ group, fields, onChange, nested }) => {
  useLocale();
  const filterable = fields.filter((f) => f.filterable);
  const setCond = (i: number, next: SearchCondition) => {
    const conditions = group.conditions.slice();
    conditions[i] = next;
    onChange({ ...group, conditions });
  };
  const fieldOf = (name: string) => filterable.find((f) => f.name === name);

  const valueInput = (c: SearchCondition, i: number) => {
    if (c.operator === 'is_null' || c.operator === 'not_null') return null;
    const field = fieldOf(c.field);
    if (c.operator === 'in' || c.operator === 'between') {
      return (
        <Input
          placeholder={c.operator === 'in' ? tx('逗号分隔') : 'min,max'}
          value={Array.isArray(c.value) ? c.value.join(',') : String(c.value ?? '')}
          onChange={(e) => setCond(i, { ...c, value: parseIn(e.target.value) })}
        />
      );
    }
    if (field?.type === 'number') {
      return (
        <InputNumber
          style={{ width: '100%' }}
          value={typeof c.value === 'number' ? c.value : undefined}
          onChange={(v) => setCond(i, { ...c, value: v ?? 0 })}
        />
      );
    }
    return (
      <Input
        value={typeof c.value === 'string' || typeof c.value === 'number' ? String(c.value) : ''}
        onChange={(e) => setCond(i, { ...c, value: e.target.value })}
      />
    );
  };

  return (
    <div className={nested ? 'search-nested' : undefined}>
      <Space style={{ marginBottom: 8 }}>
        <span>{tx('组合')}</span>
        <Select
          style={{ width: 90 }}
          value={group.logic}
          options={logicOptions}
          onChange={(logic) => onChange({ ...group, logic: logic as SearchGroup['logic'] })}
        />
        <Button
          size="small"
          icon={<PlusOutlined />}
          onClick={() =>
            onChange({
              ...group,
              conditions: [...group.conditions, emptyCond(filterable)],
            })
          }
        >
          {tx('条件')}
        </Button>
        <Button
          size="small"
          onClick={() =>
            onChange({
              ...group,
              groups: [...(group.groups || []), { logic: 'or', conditions: [] }],
            })
          }
        >
          {tx('子组')}
        </Button>
      </Space>
      {group.conditions.map((c, i) => {
        const field = fieldOf(c.field);
        const ops = (field?.operators || ['eq']).map((op) => ({ value: op, label: op }));
        return (
          <div className="search-filter-row" key={`${c.field}-${i}`}>
            <Select
              value={c.field}
              options={filterable.map((f) => ({ value: f.name, label: f.label }))}
              onChange={(name) => {
                const nf = fieldOf(name);
                setCond(i, { field: name, operator: nf?.operators[0] || 'eq', value: '' });
              }}
            />
            <Select
              value={c.operator}
              options={ops}
              onChange={(operator) => setCond(i, { ...c, operator, value: '' })}
            />
            {valueInput(c, i)}
            <Button
              type="text"
              icon={<MinusCircleOutlined />}
              onClick={() =>
                onChange({
                  ...group,
                  conditions: group.conditions.filter((_, j) => j !== i),
                })
              }
            />
          </div>
        );
      })}
      {(group.groups || []).map((g, i) => (
        <FilterBuilder
          key={`g-${i}`}
          nested
          group={g}
          fields={fields}
          onChange={(next) => {
            const groups = (group.groups || []).slice();
            groups[i] = next;
            onChange({ ...group, groups });
          }}
        />
      ))}
    </div>
  );
};

export function pruneGroup(g: SearchGroup): SearchGroup | undefined {
  const conditions = g.conditions.filter((c) => {
    if (!c.field) return false;
    if (c.operator === 'is_null' || c.operator === 'not_null') return true;
    if (c.value === '' || c.value == null) return false;
    if (Array.isArray(c.value) && c.value.length === 0) return false;
    return true;
  });
  const groups = (g.groups || [])
    .map((child) => pruneGroup(child))
    .filter((child): child is SearchGroup => Boolean(child));
  if (!conditions.length && !groups.length) return undefined;
  return { logic: g.logic, conditions, groups };
}

export { emptyCond };
export default FilterBuilder;
