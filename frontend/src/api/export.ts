import request from './request';
import { downloadBlob } from '@/utils/download';
import type {
  ExportDownloadInfo,
  ExportTableRequest,
  ExportTask,
  ExportTemplateInfo,
  ExportTemplateRequest,
} from '@/types/api';

export function exportExcel(body: ExportTableRequest): Promise<ExportTask> {
  return request.post('/export/excel', body);
}

export function exportCsv(body: ExportTableRequest): Promise<ExportTask> {
  return request.post('/export/csv', body);
}

export function exportPdf(body: ExportTableRequest): Promise<ExportTask> {
  return request.post('/export/pdf', body);
}

export function exportJson(body: ExportTableRequest): Promise<ExportTask> {
  return request.post('/export/json', body);
}

export function exportTemplate(templateId: string, body: ExportTemplateRequest): Promise<ExportTask> {
  return request.post(`/export/template/${encodeURIComponent(templateId)}`, body);
}

export function getExportTask(taskId: string): Promise<ExportTask> {
  return request.get(`/export/tasks/${encodeURIComponent(taskId)}`);
}

export function getExportTemplates(): Promise<ExportTemplateInfo[]> {
  return request.get('/export/templates');
}

export function getExportDownload(fileId: string): Promise<ExportDownloadInfo> {
  return request.get(`/export/download/${encodeURIComponent(fileId)}`);
}

export async function downloadExportFile(fileId: string, filename: string): Promise<void> {
  try {
    const info = await getExportDownload(fileId);
    if (info.url) {
      window.open(info.url, '_blank');
      return;
    }
  } catch {
    // fall through to stream
  }
  const res = await request.get(`/export/download/${encodeURIComponent(fileId)}`, {
    params: { stream: 1 },
    responseType: 'blob',
  });
  const blob = (res as { data?: Blob }).data ?? (res as unknown as Blob);
  downloadBlob(blob, filename || 'export');
}
