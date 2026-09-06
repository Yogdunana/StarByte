import request from './request';
import type { SearchQueryBody, SearchResource, SearchResult } from '@/types/api';

export function getSearchResources(): Promise<SearchResource[]> {
  return request.get('/system/search/resources');
}

export function runSearchQuery(body: SearchQueryBody): Promise<SearchResult> {
  return request.post('/system/search/query', body);
}
