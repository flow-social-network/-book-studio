import {firebaseAuth} from '@/lib/firebase/client';

const baseURL = process.env.NEXT_PUBLIC_BOOK_STUDIO_API_URL?.replace(/\/$/, '');

export class APIError extends Error {
  constructor(message: string, public readonly status: number, public readonly code: string) {
    super(message);
    this.name = 'APIError';
  }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  if (!baseURL) throw new Error('NEXT_PUBLIC_BOOK_STUDIO_API_URL is not configured.');
  const user = firebaseAuth().currentUser;
  if (!user) throw new APIError('Entre na sua conta para continuar.', 401, 'UNAUTHENTICATED');
  const token = await user.getIdToken();
  const response = await fetch(`${baseURL}${path.startsWith('/') ? path : `/${path}`}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...init.headers,
    },
    cache: 'no-store',
  });
  if (!response.ok) {
    let code = 'API_ERROR';
    let message = `A API respondeu com HTTP ${response.status}.`;
    try {
      const body = await response.json();
      code = body?.error?.code ?? code;
      message = body?.error?.message ?? message;
    } catch { /* The HTTP status remains the source of truth. */ }
    throw new APIError(message, response.status, code);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export type BookContent = {type:'doc';content:Array<Record<string, unknown>>};
export type Project = {
  id:string;title:string;author:string;subtitle:string;genre:string;language:string;
  format:string;description:string;status:string;content:BookContent;
  createdAt:string;updatedAt:string;
};
export type Chapter = {
  id:string;projectId:string;title:string;position:number;content:BookContent;
  createdAt:string;updatedAt:string;
};
export type ProjectInput = Omit<Project,'id'|'createdAt'|'updatedAt'>;

export async function downloadAPIFile(path: string, filename: string): Promise<void> {
  if (!baseURL) throw new Error('NEXT_PUBLIC_BOOK_STUDIO_API_URL is not configured.');
  const user = firebaseAuth().currentUser;
  if (!user) throw new APIError('Entre na sua conta para continuar.', 401, 'UNAUTHENTICATED');
  const response = await fetch(`${baseURL}${path}`, {headers:{Authorization:`Bearer ${await user.getIdToken()}`},cache:'no-store'});
  if (!response.ok) {
    let message = `A API respondeu com HTTP ${response.status}.`;
    try { const body = await response.json(); message = body?.error?.message ?? message; } catch { /* Keep HTTP error. */ }
    throw new APIError(message,response.status,'EXPORT_ERROR');
  }
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}

export async function uploadAPIFile(file: File, projectId?: string): Promise<unknown> {
  if (!baseURL) throw new Error('NEXT_PUBLIC_BOOK_STUDIO_API_URL is not configured.');
  const user = firebaseAuth().currentUser;
  if (!user) throw new APIError('Entre na sua conta para continuar.', 401, 'UNAUTHENTICATED');
  const body = new FormData();
  body.append('file', file);
  if (projectId) body.append('projectId', projectId);
  const response = await fetch(`${baseURL}/api/v1/assets`, {method:'POST',headers:{Authorization:`Bearer ${await user.getIdToken()}`},body,cache:'no-store'});
  if (!response.ok) {
    let message = `A API respondeu com HTTP ${response.status}.`;
    try { const data = await response.json(); message = data?.error?.message ?? message; } catch { /* Keep HTTP error. */ }
    throw new APIError(message,response.status,'UPLOAD_ERROR');
  }
  return response.json();
}
export async function deleteAPIAsset(id: string): Promise<void> {
  return api<void>(`/api/v1/assets/${id}`,{method:'DELETE'});
}
export async function downloadAPIAsset(id: string, filename: string): Promise<void> {
  return downloadAPIFile(`/api/v1/assets/${id}/download`,filename);
}
