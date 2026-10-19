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
