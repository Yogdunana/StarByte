import React, { useEffect, useState } from 'react';
import { Button, Descriptions, Drawer, Input, List, Space, Tabs, Tag, Upload, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { useSelector } from 'react-redux';
import { selectCurrentUser } from '@/store/slices/userSlice';
import {
  addTaskComment, deleteTaskAttachment, deleteTaskComment, downloadTaskAttachment,
  getTaskAttachments, getTaskComments, getTaskDetail, getTaskLogs, updateTaskStatus,
  uploadTaskAttachment, urgeTask,
} from '@/api/task';
import type { Task, TaskAttachment, TaskComment, TaskLog } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap } from './meta';

interface Props {
  taskId: string | null;
  open: boolean;
  onClose: () => void;
  onChanged: () => void;
}

const DetailDrawer: React.FC<Props> = ({ taskId, open, onClose, onChanged }) => {
  const canUpdate = usePermission('task:update');
  const canComment = usePermission('task:comment');
  const canCreate = usePermission('task:create');
  const me = useSelector(selectCurrentUser);
  const [task, setTask] = useState<Task | null>(null);
  const [comments, setComments] = useState<TaskComment[]>([]);
  const [logs, setLogs] = useState<TaskLog[]>([]);
  const [files, setFiles] = useState<TaskAttachment[]>([]);
  const [content, setContent] = useState('');

  const load = async (id: string) => {
    const [t, c, l, a] = await Promise.all([
      getTaskDetail(id),
      getTaskComments(id),
      getTaskLogs(id),
      getTaskAttachments(id),
    ]);
    setTask(t);
    setComments(c || []);
    setLogs(l || []);
    setFiles(a || []);
  };

  useEffect(() => {
    if (open && taskId) {
      void load(taskId);
    }
  }, [open, taskId]);

  const changeStatus = async (status: number) => {
    if (!taskId) return;
    await updateTaskStatus(taskId, status);
    message.success('状态已更新');
    await load(taskId);
    onChanged();
  };

  return (
    <Drawer title={task?.title || '任务详情'} width={640} open={open} onClose={onClose}>
      {task && (
        <>
          <Descriptions column={2} size="small" style={{ marginBottom: 16 }}>
            <Descriptions.Item label="状态"><StatusTag status={task.status} mapping={TaskStatusMap} /></Descriptions.Item>
            <Descriptions.Item label="优先级"><StatusTag status={task.priority} mapping={TaskPriorityMap} /></Descriptions.Item>
            <Descriptions.Item label="负责人">{task.assignee?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="创建人">{task.creator?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="部门">{task.department?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="截止">{task.due_date?.replace('T', ' ').slice(0, 16) || '-'}</Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>
              {(task.tags || []).map((tag) => <Tag key={tag}>{tag}</Tag>)}
            </Descriptions.Item>
            <Descriptions.Item label="说明" span={2}>{task.description || '-'}</Descriptions.Item>
          </Descriptions>
          <Space wrap style={{ marginBottom: 16 }}>
            {canUpdate && task.status === 0 && <Button onClick={() => changeStatus(1)}>开始</Button>}
            {canUpdate && task.status === 1 && <Button onClick={() => changeStatus(2)}>完成</Button>}
            {canUpdate && task.status === 1 && <Button onClick={() => changeStatus(4)}>挂起</Button>}
            {canUpdate && task.status === 4 && <Button onClick={() => changeStatus(1)}>恢复</Button>}
            {canUpdate && (task.status === 0 || task.status === 1) && <Button danger onClick={() => changeStatus(3)}>取消</Button>}
            {canCreate && task.creator?.id === me?.id && task.assignee && (
              <Button onClick={() => urgeTask(task.id, '请尽快处理').then(() => message.success('已催办'))}>催办</Button>
            )}
          </Space>
          <Tabs
            items={[
              {
                key: 'comments',
                label: `评论 (${task.comment_count})`,
                children: (
                  <>
                    <List
                      dataSource={comments}
                      locale={{ emptyText: '暂无评论' }}
                      renderItem={(item) => (
                        <List.Item
                          actions={item.user_id === me?.id ? [
                            <Button key="d" type="link" size="small" danger onClick={() => deleteTaskComment(task.id, item.id).then(() => load(task.id))}>删除</Button>,
                          ] : undefined}
                        >
                          <List.Item.Meta title={item.user?.name} description={item.content} />
                        </List.Item>
                      )}
                    />
                    {canComment && (
                      <Space.Compact style={{ width: '100%', marginTop: 12 }}>
                        <Input.TextArea
                          rows={2}
                          value={content}
                          onChange={(e) => setContent(e.target.value)}
                          placeholder="评论，可用 @username 提及"
                        />
                        <Button
                          type="primary"
                          onClick={async () => {
                            if (!content.trim()) return;
                            await addTaskComment(task.id, content.trim());
                            setContent('');
                            await load(task.id);
                            onChanged();
                          }}
                        >
                          发送
                        </Button>
                      </Space.Compact>
                    )}
                  </>
                ),
              },
              {
                key: 'files',
                label: `附件 (${task.attachment_count})`,
                children: (
                  <>
                    <List
                      dataSource={files}
                      locale={{ emptyText: '暂无附件' }}
                      renderItem={(item) => (
                        <List.Item
                          actions={[
                            <Button key="dl" type="link" size="small" onClick={() => downloadTaskAttachment(task.id, item.id, item.file_name)}>下载</Button>,
                            canUpdate ? <Button key="rm" type="link" size="small" danger onClick={() => deleteTaskAttachment(task.id, item.id).then(() => load(task.id))}>删除</Button> : null,
                          ]}
                        >
                          {item.file_name} ({Math.round(item.file_size / 1024)} KB)
                        </List.Item>
                      )}
                    />
                    {canUpdate && (
                      <Upload
                        showUploadList={false}
                        beforeUpload={async (file) => {
                          await uploadTaskAttachment(task.id, file);
                          message.success('已上传');
                          await load(task.id);
                          onChanged();
                          return false;
                        }}
                      >
                        <Button icon={<UploadOutlined />} style={{ marginTop: 8 }}>上传附件</Button>
                      </Upload>
                    )}
                  </>
                ),
              },
              {
                key: 'logs',
                label: '流转历史',
                children: (
                  <List
                    dataSource={logs}
                    locale={{ emptyText: '暂无记录' }}
                    renderItem={(item) => (
                      <List.Item>
                        <List.Item.Meta
                          title={`${item.operator?.name} · ${item.action_type}`}
                          description={`${item.old_value || '-'} → ${item.new_value || '-'} ${item.comment || ''} ${item.created_at?.replace('T', ' ').slice(0, 16)}`}
                        />
                      </List.Item>
                    )}
                  />
                ),
              },
            ]}
          />
        </>
      )}
    </Drawer>
  );
};

export default DetailDrawer;
