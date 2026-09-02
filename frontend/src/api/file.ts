import request from './request';
import type { FileInfo, UploadResult, ListFileParams, PageResponse } from '@/types/api';

// 获取文件列表
export function getFileList(params: ListFileParams): Promise<PageResponse<FileInfo>> {
  return request.get('/files', { params });
}

// 获取文件详情
export function getFileDetail(id: string): Promise<FileInfo> {
  return request.get(`/files/${id}`);
}

// 删除文件
export function deleteFile(id: string): Promise<void> {
  return request.delete(`/files/${id}`);
}

/**
 * 获取上传地址（上传文件时使用，支持上传进度）
 *
 * @example
 * ```ts
 * const formData = new FormData();
 * formData.append('file', file);
 * const result = await uploadFile(formData, (progressEvent) => {
 *   const percent = Math.round((progressEvent.loaded * 100) / (progressEvent.total || 0));
 *   console.log(percent);
 * });
 * ```
 */
export function uploadFile(
  formData: FormData,
  onUploadProgress?: (progressEvent: { loaded: number; total?: number }) => void,
): Promise<UploadResult> {
  return request.post('/files/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress,
  });
}
