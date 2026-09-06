import request from './request';

export interface DictType {
  id: string;
  code: string;
  name: string;
  description: string;
  sort_order: number;
  status: 0 | 1;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface DictItem {
  id: string;
  type_id: string;
  type_code: string;
  item_value: string;
  item_label: string;
  sort_order: number;
  status: 0 | 1;
  created_at: string;
  updated_at: string;
}

export interface CreateDictTypeParams {
  code: string;
  name: string;
  description?: string;
  sort_order?: number;
  status?: 0 | 1;
}

export interface UpdateDictTypeParams {
  name?: string;
  description?: string;
  sort_order?: number;
  status?: 0 | 1;
}

export interface CreateDictItemParams {
  type_code: string;
  item_value: string;
  item_label: string;
  sort_order?: number;
  status?: 0 | 1;
}

export interface UpdateDictItemParams {
  item_label?: string;
  sort_order?: number;
  status?: 0 | 1;
}

export function getDictTypes(): Promise<DictType[]> {
  return request.get('/system/dicts/types');
}

export function createDictType(data: CreateDictTypeParams): Promise<DictType> {
  return request.post('/system/dicts/types', data);
}

export function updateDictType(id: string, data: UpdateDictTypeParams): Promise<DictType> {
  return request.put(`/system/dicts/types/${id}`, data);
}

export function deleteDictType(id: string): Promise<void> {
  return request.delete(`/system/dicts/types/${id}`);
}

export function getDictItems(typeCode: string, all = false): Promise<DictItem[]> {
  return request.get(`/system/dicts/${typeCode}`, { params: all ? { all: 1 } : undefined });
}

export function createDictItem(data: CreateDictItemParams): Promise<DictItem> {
  return request.post('/system/dicts', data);
}

export function updateDictItem(id: string, data: UpdateDictItemParams): Promise<DictItem> {
  return request.put(`/system/dicts/${id}`, data);
}

export function deleteDictItem(id: string): Promise<void> {
  return request.delete(`/system/dicts/${id}`);
}
