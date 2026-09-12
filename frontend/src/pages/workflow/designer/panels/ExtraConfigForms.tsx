import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Checkbox, Form, Input, InputNumber, Select } from 'antd';
import type {
  NotifyChannel,
  NotifyConfig,
  ParallelConfig,
  TimerConfig,
  TimerUnit,
} from '@/types/workflow';
import { NOTIFY_TEMPLATES } from '../constants';

interface ParallelFormProps {
  value: ParallelConfig;
  disabled?: boolean;
  onChange: (next: ParallelConfig) => void;
}

export const ParallelConfigForm: React.FC<ParallelFormProps> = ({ value, disabled, onChange }) => {
  useLocale();
  return (
    <>
      <Form.Item label={tx('分支数量')}>
        <InputNumber
          disabled={disabled}
          min={2}
          max={8}
          value={value.branchCount}
          onChange={(branchCount) => {
            const count = branchCount ?? 2;
            const labels = [...(value.branchLabels ?? [])];
            while (labels.length < count)
              labels.push(tx('分支{{value0}}', { value0: labels.length + 1 }));
            onChange({ ...value, branchCount: count, branchLabels: labels.slice(0, count) });
          }}
        />
      </Form.Item>
      {(value.branchLabels ?? []).map((label, index) => (
        <Form.Item
          key={`branch-label-${index}`}
          label={tx('分支{{value0}}说明', { value0: index + 1 })}
        >
          <Input
            disabled={disabled}
            value={label}
            onChange={(event) => {
              const branchLabels = [...value.branchLabels];
              branchLabels[index] = event.target.value;
              onChange({ ...value, branchLabels });
            }}
          />
        </Form.Item>
      ))}
    </>
  );
};

interface TimerFormProps {
  value: TimerConfig;
  disabled?: boolean;
  onChange: (next: TimerConfig) => void;
}

export const TimerConfigForm: React.FC<TimerFormProps> = ({ value, disabled, onChange }) => {
  useLocale();
  return (
    <>
      <Form.Item label={tx('持续时长')}>
        <InputNumber
          disabled={disabled}
          min={1}
          value={value.duration}
          onChange={(duration) => onChange({ ...value, duration: duration ?? 1 })}
        />
      </Form.Item>
      <Form.Item label={tx('单位')}>
        <Select
          disabled={disabled}
          value={value.unit}
          onChange={(unit: TimerUnit) => onChange({ ...value, unit })}
          options={[
            { label: tx('分钟'), value: 'minutes' },
            { label: tx('小时'), value: 'hours' },
            { label: tx('天'), value: 'days' },
          ]}
        />
      </Form.Item>
    </>
  );
};

interface NotifyFormProps {
  value: NotifyConfig;
  disabled?: boolean;
  onChange: (next: NotifyConfig) => void;
}

export const NotifyConfigForm: React.FC<NotifyFormProps> = ({ value, disabled, onChange }) => {
  useLocale();
  return (
    <>
      <Form.Item label={tx('通知渠道')}>
        <Checkbox.Group
          disabled={disabled}
          value={value.channels}
          onChange={(channels) => onChange({ ...value, channels: channels as NotifyChannel[] })}
          options={[
            { label: tx('站内信'), value: 'in_app' },
            { label: tx('邮件'), value: 'email' },
          ]}
        />
      </Form.Item>
      <Form.Item label={tx('通知模板')}>
        <Select
          disabled={disabled}
          value={value.notificationType}
          onChange={(notificationType: string) => onChange({ ...value, notificationType })}
          options={NOTIFY_TEMPLATES}
        />
      </Form.Item>
    </>
  );
};
