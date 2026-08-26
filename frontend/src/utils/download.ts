/**
 * 文件下载工具
 */

/**
 * 通过 Blob 下载文件
 * @param blob Blob 对象
 * @param filename 下载文件名
 */
export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * 通过 URL 下载文件
 * @param url 文件 URL
 * @param filename 下载文件名
 * @param token 鉴权 Token（可选）
 */
export async function downloadFile(
  url: string,
  filename?: string,
  token?: string,
): Promise<void> {
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, { headers });
  if (!response.ok) {
    throw new Error(`下载失败: ${response.status}`);
  }

  const blob = await response.blob();
  const name = filename || getFilenameFromResponse(response) || 'download';
  downloadBlob(blob, name);
}

/**
 * 从响应头获取文件名
 */
function getFilenameFromResponse(response: Response): string | null {
  const disposition = response.headers.get('Content-Disposition');
  if (!disposition) return null;
  const match = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';\s]+)["']?/i);
  return match ? decodeURIComponent(match[1]) : null;
}

/**
 * 导出数据为 CSV 文件
 * @param data 二维数组（第一行为表头）
 * @param filename 文件名
 */
export function exportCSV(data: string[][], filename: string): void {
  const csvContent = data
    .map((row) => row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(','))
    .join('\n');
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
  downloadBlob(blob, filename.endsWith('.csv') ? filename : `${filename}.csv`);
}
