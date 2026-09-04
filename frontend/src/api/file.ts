import request from './request';
import type { FileInfo, UploadResult, ListFileParams, PageResponse } from '@/types/api';

function multipartHeaders(): { 'Content-Type': undefined } {
  // 去掉默认 application/json，让浏览器写入 multipart boundary
  return { 'Content-Type': undefined };
}

export function getFileList(params: ListFileParams): Promise<PageResponse<FileInfo>> {
  return request.get('/files', { params });
}

export function getFileDetail(id: string): Promise<FileInfo> {
  return request.get(`/files/${id}`);
}

export function deleteFile(id: string): Promise<void> {
  return request.delete(`/files/${id}`);
}

export function downloadFile(id: string): Promise<FileInfo> {
  return getFileDetail(id);
}

export function uploadFile(
  formData: FormData,
  onUploadProgress?: (progressEvent: { loaded: number; total?: number }) => void,
): Promise<UploadResult> {
  return request.post('/files/upload', formData, {
    onUploadProgress,
    headers: multipartHeaders(),
  });
}

export function uploadFileBatch(
  formData: FormData,
  onUploadProgress?: (progressEvent: { loaded: number; total?: number }) => void,
): Promise<UploadResult[]> {
  return request.post('/files/upload-batch', formData, {
    onUploadProgress,
    headers: multipartHeaders(),
  });
}
