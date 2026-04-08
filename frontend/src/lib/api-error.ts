import { isAxiosError } from 'axios';
import { notifyError } from '@/lib/notify';

export function extractApiErrorMessage(error: unknown): string | null {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? error.message ?? null;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return null;
}

export function notifyApiError(error: unknown, fallbackMessage: string): void {
  notifyError(extractApiErrorMessage(error) ?? fallbackMessage);
}
