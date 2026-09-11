import request from './request';

export interface Equipment {
  id: string;
  name: string;
  model: string;
  category: string;
  category_text: string;
  total_quantity: number;
  available_quantity: number;
  location: string;
  manager_id: string;
  manager_name: string;
  photo_file_id: string;
  photo_url: string;
  status: number;
  status_text: string;
  qr_code: string;
  description: string;
  purchase_date: string;
  purchase_price: number;
  created_at: string;
  updated_at: string;
}

export interface BorrowRecord {
  id: string;
  equipment_id: string;
  equipment_name: string;
  borrower_id: string;
  borrower_name: string;
  quantity: number;
  borrow_at: string;
  expected_return_at: string;
  actual_return_at: string;
  status: number;
  status_text: string;
  approver_id: string;
  approver_name: string;
  approved_at: string;
  approved_remark: string;
  return_remark: string;
  return_checker_id: string;
  remark: string;
  created_at: string;
  updated_at: string;
}

export interface MaintenanceRecord {
  id: string;
  equipment_id: string;
  equipment_name: string;
  type: number;
  type_text: string;
  description: string;
  cost: number;
  start_date: string;
  end_date: string;
  status: number;
  status_text: string;
  operator_id: string;
  remark: string;
  created_at: string;
}

export interface InventoryRecord {
  id: string;
  equipment_id: string;
  equipment_name: string;
  expected_quantity: number;
  actual_quantity: number;
  difference: number;
  checker_id: string;
  check_date: string;
  remark: string;
  created_at: string;
}

export function getEquipmentList(params: {
  page?: number;
  page_size?: number;
  keyword?: string;
  category?: string;
  status?: number;
}) {
  return request.get('/equipment', { params });
}

export function createEquipment(data: {
  name: string;
  model?: string;
  category: string;
  total_quantity: number;
  location?: string;
  manager_id?: string;
  photo_file_id?: string;
  qr_code?: string;
  description?: string;
  purchase_date?: string;
  purchase_price?: number;
}) {
  return request.post('/equipment', data);
}

export function updateEquipment(id: string, data: Partial<{
  name: string;
  model: string;
  category: string;
  total_quantity: number;
  location: string;
  manager_id: string;
  status: number;
  qr_code: string;
  description: string;
}>) {
  return request.put(`/equipment/${id}`, data);
}

export function deleteEquipment(id: string) {
  return request.delete(`/equipment/${id}`);
}

export function createBorrow(data: {
  equipment_id: string;
  quantity: number;
  expected_return_at: string;
  remark?: string;
}) {
  return request.post('/equipment/borrow', data);
}

export function approveBorrow(id: string, data: { status: number; remark?: string }) {
  return request.put(`/equipment/borrow/${id}/approve`, data);
}

export function returnBorrow(id: string, data: { remark?: string }) {
  return request.put(`/equipment/borrow/${id}/return`, data);
}

export function getBorrowList(params: {
  page?: number;
  page_size?: number;
  equipment_id?: string;
  borrower_id?: string;
  status?: number;
}) {
  return request.get('/equipment/borrow', { params });
}

export function getEquipmentRecords(id: string, params: { page?: number; page_size?: number }) {
  return request.get(`/equipment/${id}/records`, { params });
}

export function createMaintenance(data: {
  equipment_id: string;
  type: number;
  description: string;
  cost?: number;
  start_date: string;
  end_date?: string;
  remark?: string;
}) {
  return request.post('/equipment/maintenance', data);
}

export function getMaintenanceList(params: {
  page?: number;
  page_size?: number;
  equipment_id?: string;
  status?: number;
}) {
  return request.get('/equipment/maintenance', { params });
}

export function createInventory(data: {
  equipment_id: string;
  actual_quantity: number;
  remark?: string;
}) {
  return request.post('/equipment/inventory', data);
}

export function getInventoryList(params: {
  page?: number;
  page_size?: number;
  equipment_id?: string;
}) {
  return request.get('/equipment/inventory', { params });
}
