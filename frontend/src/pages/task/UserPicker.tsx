import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Select } from 'antd';
import { getTaskCandidates } from '@/api/task';
interface Props {
  id?: string;
  kind: 'create' | 'assign' | 'transfer';
  value?: string;
  onChange?: (value?: string) => void;
  disabled?: boolean;
}
export default function UserPicker({ id, kind, value, onChange, disabled }: Props) {
  useLocale();
  const [keyword, setKeyword] = useState('');
  const [users, setUsers] = useState<Array<{ id: string; name: string }>>([]);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const seq = useRef(0);
  const invalidate = useCallback(() => {
    seq.current++;
  }, []);
  useEffect(() => {
    const current = ++seq.current;
    const timer = window.setTimeout(() => {
      setLoading(true);
      void getTaskCandidates(kind, keyword)
        .then((rows) => {
          if (seq.current === current) {
            setUsers(rows);
            setFailed(false);
          }
        })
        .catch(() => {
          if (seq.current === current) setFailed(true);
        })
        .finally(() => {
          if (seq.current === current) setLoading(false);
        });
    }, 300);
    return () => {
      window.clearTimeout(timer);
      invalidate();
    };
  }, [kind, keyword, invalidate]);
  return (
    <>
      <Select
        id={id}
        style={{ width: '100%' }}
        value={value}
        onChange={onChange}
        disabled={disabled}
        allowClear
        showSearch
        filterOption={false}
        onSearch={setKeyword}
        loading={loading}
        placeholder={tx('搜索姓名或用户名')}
        options={users.map((u) => ({ value: u.id, label: u.name }))}
        notFoundContent={loading ? tx('正在查找…') : tx('没有匹配的可用账号')}
      />
      {failed && (
        <Alert
          type="warning"
          message={tx('人员加载失败，请重新输入姓名搜索')}
          showIcon
          style={{ marginTop: 8 }}
        />
      )}
    </>
  );
}
