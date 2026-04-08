import MD5 from 'md5.js';
import { http } from '@/api/client/http';
import type { ApiEnvelope } from '@/types/auth';

type UploadMode = 'direct' | 'proxy';

export interface UploadedFile {
  id: string;
  storage_driver: string;
  original_name: string;
  ext: string;
  mime_type: string;
  size: number;
  upload_status: string;
  file_url: string;
  created_at: string;
  updated_at: string;
}

interface InitDirectUploadReq {
  file_md5: string;
  file_name: string;
  file_size: number;
  mime_type?: string;
}

interface CompleteDirectUploadReq {
  file_id: string;
  file_name: string;
  file_size: number;
  mime_type?: string;
}

interface DirectUploadInitResp {
  mode: UploadMode;
  already_exists: boolean;
  file_id: string;
  storage_driver: string;
  upload_method?: string;
  upload_url?: string;
  upload_headers?: Record<string, string>;
  expires_at?: string;
  file?: UploadedFile;
}

export async function uploadUserFile(file: File): Promise<UploadedFile> {
  const directUpload = await initDirectUpload(file);
  if (directUpload.mode === 'direct') {
    if (directUpload.already_exists && directUpload.file) {
      return directUpload.file;
    }

    if (!directUpload.file_id || !directUpload.upload_url) {
      throw new Error('Invalid direct upload response');
    }

    await uploadToSignedURL(file, directUpload);
    return completeDirectUpload({
      file_id: directUpload.file_id,
      file_name: file.name,
      file_size: file.size,
      mime_type: normalizeMimeType(file)
    });
  }

  return uploadUserFileByProxy(file);
}

async function uploadUserFileByProxy(file: File): Promise<UploadedFile> {
  const formData = new FormData();
  formData.append('file', file);

  const { data } = await http.post<ApiEnvelope<UploadedFile>>('/user/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data.data;
}

async function initDirectUpload(file: File): Promise<DirectUploadInitResp> {
  const payload: InitDirectUploadReq = {
    file_md5: await buildFileFingerprint(file),
    file_name: file.name,
    file_size: file.size,
    mime_type: normalizeMimeType(file)
  };

  const { data } = await http.post<ApiEnvelope<DirectUploadInitResp>>('/user/files/direct/init', payload);
  return data.data;
}

async function completeDirectUpload(payload: CompleteDirectUploadReq): Promise<UploadedFile> {
  const { data } = await http.post<ApiEnvelope<UploadedFile>>('/user/files/direct/complete', payload);
  return data.data;
}

async function uploadToSignedURL(file: File, ticket: DirectUploadInitResp): Promise<void> {
  const headers = new Headers(ticket.upload_headers ?? {});
  const mimeType = normalizeMimeType(file);
  if (mimeType && !headers.has('Content-Type')) {
    headers.set('Content-Type', mimeType);
  }

  const response = await fetch(ticket.upload_url ?? '', {
    method: ticket.upload_method || 'PUT',
    headers,
    body: file
  });

  if (!response.ok) {
    throw new Error(`Direct upload failed with status ${response.status}`);
  }
}

async function buildFileFingerprint(file: File): Promise<string> {
  const buffer = await file.arrayBuffer();
  return new MD5().update(new Uint8Array(buffer)).digest('hex');
}

function normalizeMimeType(file: File): string | undefined {
  const mimeType = file.type.trim();
  return mimeType || undefined;
}
