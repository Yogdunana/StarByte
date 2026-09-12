import type { TFunction } from 'i18next';
import type { StatusMap } from '@/types/common';
import type { AnnouncementCategory } from '@/api/announcement';

export function announcementStatusMap(t: TFunction): StatusMap {
  return {
    0: { color: 'default', text: t('announcement.status.0') },
    1: { color: 'success', text: t('announcement.status.1') },
    2: { color: 'warning', text: t('announcement.status.2') },
  };
}

export const AnnouncementCategories: { value: AnnouncementCategory; labelKey: string }[] = [
  { value: 'association', labelKey: 'announcement.category.association' },
  { value: 'activity', labelKey: 'announcement.category.activity' },
  { value: 'system', labelKey: 'announcement.category.system' },
  { value: 'personnel', labelKey: 'announcement.category.personnel' },
];

export function sanitizeAnnouncementHTML(html: string): string {
  return html
    .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
    .replace(/<iframe[\s\S]*?>[\s\S]*?<\/iframe>/gi, '')
    .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, '')
    .replace(/\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '');
}
