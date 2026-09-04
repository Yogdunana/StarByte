import React, { useCallback, useEffect, useState } from 'react';
import { Card, Modal, Table, Upload, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import { deleteFile, getFileList, uploadFile } from '@/api/file';
import type { FileInfo, ListFileParams } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import FileFilterBar, { type FileFilterValue } from './FileFilterBar';
import { buildFileColumns } from './fileColumns';

const emptyFilter: FileFilterValue = { keyword: '', category: undefined };

const FileList: React.FC = () => {
  const canCreate = usePermission('file:create');
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<FileInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [filters, setFilters] = useState<FileFilterValue>(emptyFilter);

  const buildParams = useCallback(
    (p: number, ps: number): ListFileParams => ({
      page: p,
      page_size: ps,
      keyword: filters.keyword || undefined,
      category: filters.category,
    }),
    [filters],
  );

  const loadList = useCallback(async (params: ListFileParams) => {
    setLoading(true);
    try {
      const res = await getFileList(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载文件列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadList(buildParams(page, pageSize));
  }, [page, pageSize, buildParams, loadList]);

  const handleDownload = (record: FileInfo) => {
    if (record.url) {
      window.open(record.url, '_blank');
      return;
    }
    message.warning('暂无下载地址');
  };

  const handleDelete = (record: FileInfo) => {
    Modal.confirm({
      title: '删除文件',
      content: `确定删除「${record.original_name || record.filename}」？将同时删除存储对象。`,
      onOk: async () => {
        await deleteFile(record.id);
        message.success('删除成功');
        loadList(buildParams(page, pageSize));
      },
    });
  };

  const customRequest: UploadProps['customRequest'] = async (options) => {
    const formData = new FormData();
    formData.append('file', options.file as File);
    try {
      await uploadFile(formData, (evt) => {
        const percent = Math.round((evt.loaded * 100) / (evt.total || 1));
        options.onProgress?.({ percent });
      });
      options.onSuccess?.({});
      message.success('上传成功');
      loadList(buildParams(page, pageSize));
    } catch (err) {
      options.onError?.(err as Error);
    }
  };

  return (
    <Card title="文件管理">
      <div style={{ display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap' }}>
        <FileFilterBar
          value={filters}
          onChange={setFilters}
          onSearch={() => {
            setPage(1);
            loadList(buildParams(1, pageSize));
          }}
          onReset={() => {
            setFilters(emptyFilter);
            setPage(1);
            loadList({ page: 1, page_size: pageSize });
          }}
        />
        {canCreate && (
          <Upload customRequest={customRequest} showUploadList={false}>
            <button type="button" style={{ border: '1px dashed #d9d9d9', padding: '4px 15px', borderRadius: 6, background: '#fff', cursor: 'pointer' }}>
              <UploadOutlined /> 上传文件
            </button>
          </Upload>
        )}
      </div>
      <Table
        columns={buildFileColumns(handleDownload, handleDelete)}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1000 }}
        size="middle"
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />
    </Card>
  );
};

export default FileList;
