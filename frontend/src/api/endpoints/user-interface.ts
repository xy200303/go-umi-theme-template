import { http } from '@/api/client/http';
import type { ApiEnvelope } from '@/types/auth';

export type InterfaceType = 'openai_api' | 'openai_response' | 'claude' | 'gemini';

export interface UserInterfaceItem {
  id: number;
  name: string;
  interface_type: InterfaceType;
  target_base_url: string;
  target_api_key_mask: string;
  default_model: string;
  enabled: boolean;
  gateway_key: string;
  gateway_key_prefix: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserInterfaceReq {
  name: string;
  interface_type: InterfaceType;
  target_base_url: string;
  target_api_key: string;
  default_model: string;
  enabled: boolean;
}

export interface UpdateUserInterfaceReq {
  name: string;
  interface_type: InterfaceType;
  target_base_url: string;
  target_api_key?: string;
  default_model: string;
  enabled: boolean;
}

export interface CreateUserInterfaceResp {
  interface: UserInterfaceItem;
  gateway_key: string;
}

export interface RegenerateUserInterfaceGatewayKeyResp {
  interface: UserInterfaceItem;
  gateway_key: string;
}

export interface TestUserInterfaceResp {
  ok: boolean;
  status_code: number;
  message: string;
}

export interface GatewayRequestLogItem {
  id: number;
  user_id: number;
  user_interface_id: number;
  interface_type: InterfaceType;
  request_method: string;
  request_path: string;
  upstream_url: string;
  model: string;
  status_code: number;
  duration_ms: number;
  error_message: string;
  created_at: string;
}

export interface GatewayRequestLogListResp {
  list: GatewayRequestLogItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface ChatRecordItem {
  id: number;
  user_id: number;
  user_interface_id: number;
  interface_type: InterfaceType;
  request_method: string;
  request_path: string;
  model: string;
  user_input: string;
  model_output: string;
  status_code: number;
  duration_ms: number;
  error_message: string;
  created_at: string;
}

export interface ChatRecordListResp {
  list: ChatRecordItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface GatewayDisplayConfigResp {
  gateway_base_url: string;
}

export async function listUserInterfaces(): Promise<UserInterfaceItem[]> {
  const { data } = await http.get<ApiEnvelope<UserInterfaceItem[]>>('/user/interfaces');
  return data.data;
}

export async function createUserInterface(payload: CreateUserInterfaceReq): Promise<CreateUserInterfaceResp> {
  const { data } = await http.post<ApiEnvelope<CreateUserInterfaceResp>>('/user/interfaces', payload);
  return data.data;
}

export async function updateUserInterface(id: number, payload: UpdateUserInterfaceReq): Promise<UserInterfaceItem> {
  const { data } = await http.put<ApiEnvelope<UserInterfaceItem>>(`/user/interfaces/${id}`, payload);
  return data.data;
}

export async function deleteUserInterface(id: number): Promise<void> {
  await http.delete(`/user/interfaces/${id}`);
}

export async function regenerateUserInterfaceGatewayKey(id: number): Promise<RegenerateUserInterfaceGatewayKeyResp> {
  const { data } = await http.post<ApiEnvelope<RegenerateUserInterfaceGatewayKeyResp>>(`/user/interfaces/${id}/regenerate-key`);
  return data.data;
}

export async function testUserInterfaceConnection(id: number): Promise<TestUserInterfaceResp> {
  const { data } = await http.post<ApiEnvelope<TestUserInterfaceResp>>(`/user/interfaces/${id}/test`);
  return data.data;
}

export async function listGatewayRequestLogs(params?: {
  keyword?: string;
  interface_id?: number;
  status_code?: number;
  page?: number;
  page_size?: number;
}): Promise<GatewayRequestLogListResp> {
  const { data } = await http.get<ApiEnvelope<GatewayRequestLogListResp>>('/user/gateway-logs', {
    params: {
      ...params,
      _ts: Date.now()
    },
    headers: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache'
    }
  });
  const payload = data?.data;
  return {
    list: Array.isArray(payload?.list) ? payload.list : [],
    total: typeof payload?.total === 'number' ? payload.total : 0,
    page: typeof payload?.page === 'number' ? payload.page : (params?.page ?? 1),
    page_size: typeof payload?.page_size === 'number' ? payload.page_size : (params?.page_size ?? 20)
  };
}

export async function getGatewayDisplayConfig(): Promise<GatewayDisplayConfigResp> {
  const { data } = await http.get<ApiEnvelope<GatewayDisplayConfigResp>>('/user/gateway-display-config');
  return {
    gateway_base_url: data?.data?.gateway_base_url ?? ''
  };
}

export async function listChatRecords(params?: {
  keyword?: string;
  interface_id?: number;
  status_code?: number;
  page?: number;
  page_size?: number;
}): Promise<ChatRecordListResp> {
  const { data } = await http.get<ApiEnvelope<ChatRecordListResp>>('/user/chat-records', {
    params: {
      ...params,
      _ts: Date.now()
    },
    headers: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache'
    }
  });
  const payload = data?.data;
  return {
    list: Array.isArray(payload?.list) ? payload.list : [],
    total: typeof payload?.total === 'number' ? payload.total : 0,
    page: typeof payload?.page === 'number' ? payload.page : (params?.page ?? 1),
    page_size: typeof payload?.page_size === 'number' ? payload.page_size : (params?.page_size ?? 20)
  };
}
