import React from 'react';
import { DatePicker, Radio, Space } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

dayjs.locale('zh-cn');

export type TimePreset = 'today' | 'week' | 'month' | 'semester' | 'custom';

export interface TimeRangeValue {
  preset: TimePreset;
  start?: string;
  end?: string;
}

interface TimeRangeSelectProps {
  value: TimeRangeValue;
  onChange: (next: TimeRangeValue) => void;
}

function semesterRange(now: Dayjs): [Dayjs, Dayjs] {
  const month = now.month();
  if (month >= 7) {
    return [now.month(7).date(1).startOf('day'), now.add(1, 'year').month(0).date(31).endOf('day')];
  }
  if (month === 0) {
    return [now.subtract(1, 'year').month(7).date(1).startOf('day'), now.month(0).date(31).endOf('day')];
  }
  return [now.month(1).date(1).startOf('day'), now.month(6).date(31).endOf('day')];
}

export function resolveTimeRange(value: TimeRangeValue): { start_date?: string; end_date?: string } {
  const now = dayjs();
  let start = value.start ? dayjs(value.start) : undefined;
  let end = value.end ? dayjs(value.end) : undefined;
  switch (value.preset) {
    case 'today':
      start = now.startOf('day');
      end = now.endOf('day');
      break;
    case 'week':
      start = now.startOf('week');
      end = now.endOf('week');
      break;
    case 'month':
      start = now.startOf('month');
      end = now.endOf('month');
      break;
    case 'semester': {
      const range = semesterRange(now);
      start = range[0];
      end = range[1];
      break;
    }
    default:
      break;
  }
  return {
    start_date: start ? start.format('YYYY-MM-DD') : undefined,
    end_date: end ? end.format('YYYY-MM-DD') : undefined,
  };
}

const TimeRangeSelect: React.FC<TimeRangeSelectProps> = ({ value, onChange }) => (
  <Space wrap>
    <Radio.Group
      value={value.preset}
      onChange={(e) => onChange({ ...value, preset: e.target.value as TimePreset })}
    >
      <Radio.Button value="today">今天</Radio.Button>
      <Radio.Button value="week">本周</Radio.Button>
      <Radio.Button value="month">本月</Radio.Button>
      <Radio.Button value="semester">本学期</Radio.Button>
      <Radio.Button value="custom">自定义</Radio.Button>
    </Radio.Group>
    {value.preset === 'custom' ? (
      <DatePicker.RangePicker
        value={value.start && value.end ? [dayjs(value.start), dayjs(value.end)] : null}
        onChange={(dates) => {
          onChange({
            preset: 'custom',
            start: dates?.[0]?.format('YYYY-MM-DD'),
            end: dates?.[1]?.format('YYYY-MM-DD'),
          });
        }}
      />
    ) : null}
  </Space>
);

export default TimeRangeSelect;
