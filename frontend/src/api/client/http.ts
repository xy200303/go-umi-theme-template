import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { getClientEnv } from '@/lib/env';
import { useAuthStore } from '@/stores';
import type { ApiEnvelope, AuthToken, AuthUser } from '@/types/auth';

const baseURL = getClientEnv('API_BASE_URL') ?? '/api/v1';

export const http = axios.create({
  baseURL,
  timeout: 15000
});

const refreshClient = axios.create({
  baseURL,
  timeout: 15000
});

async function fetchCurrentUser(accessToken: string): Promise<AuthUser> {
  const { data } = await refreshClient.get<ApiEnvelope<AuthUser>>('/user/profile', {
    headers: {
      Authorization: `Bearer ${accessToken}`
    }
  });
  return data.data;
}

async function refreshAuthToken(refreshToken: string): Promise<AuthToken> {
  const { data } = await refreshClient.post<ApiEnvelope<AuthToken>>('/auth/refresh', {
    refresh_token: refreshToken
  });
  return data.data;
}

export async function syncCurrentSession(): Promise<boolean> {
  const current = useAuthStore.getState();
  const refreshToken = current.token?.refresh_token;

  if (!refreshToken || !current.user) {
    current.clearLogin();
    return false;
  }

  try {
    const nextToken = await refreshAuthToken(refreshToken);
    let nextUser = current.user;

    try {
      nextUser = await fetchCurrentUser(nextToken.access_token);
    } catch {
      // Keep the session usable even if profile sync fails transiently.
    }

    useAuthStore.getState().setLogin(nextToken, nextUser);
    return true;
  } catch {
    useAuthStore.getState().clearLogin();
    return false;
  }
}

function attachAuth(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const token = useAuthStore.getState().token?.access_token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}

http.interceptors.request.use(attachAuth);

let isRefreshing = false;
let pendingQueue: Array<(token: string | null) => void> = [];

function resolveQueue(token: string | null): void {
  pendingQueue.forEach((cb) => cb(token));
  pendingQueue = [];
}

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    const status = error.response?.status;

    if (status !== 401 || originalRequest?._retry) {
      return Promise.reject(error);
    }

    const state = useAuthStore.getState();
    const refreshToken = state.token?.refresh_token;
    if (!refreshToken) {
      state.clearLogin();
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        pendingQueue.push((token) => {
          if (!token) {
            reject(error);
            return;
          }
          originalRequest.headers.Authorization = `Bearer ${token}`;
          resolve(http(originalRequest));
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      const sessionSynced = await syncCurrentSession();
      const current = useAuthStore.getState();
      if (!sessionSynced || !current.token) {
        return Promise.reject(error);
      }
      resolveQueue(current.token.access_token);
      originalRequest.headers.Authorization = `Bearer ${current.token.access_token}`;
      return http(originalRequest);
    } catch (refreshErr) {
      resolveQueue(null);
      useAuthStore.getState().clearLogin();
      return Promise.reject(refreshErr);
    } finally {
      isRefreshing = false;
    }
  }
);
