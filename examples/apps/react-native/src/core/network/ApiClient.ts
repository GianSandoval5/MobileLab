import { API_BASE_URL } from '../config/apiConfig';
import { AppError } from '../errors/AppError';

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export class ApiClient {
  private token?: string;

  setToken(token?: string) {
    this.token = token;
  }

  get<T>(path: string) {
    return this.request<T>('GET', path);
  }

  post<T>(path: string, body?: unknown) {
    return this.request<T>('POST', path, body);
  }

  put<T>(path: string, body?: unknown) {
    return this.request<T>('PUT', path, body);
  }

  patch<T>(path: string, body?: unknown) {
    return this.request<T>('PATCH', path, body);
  }

  delete<T>(path: string, body?: unknown) {
    return this.request<T>('DELETE', path, body);
  }

  private async request<T>(method: Method, path: string, body?: unknown) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 12000);
    try {
      const response = await fetch(`${API_BASE_URL}${path}`, {
        method,
        signal: controller.signal,
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
          ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
        },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
      const text = await response.text();
      const data = text ? JSON.parse(text) : undefined;
      if (!response.ok) {
        throw new AppError(
          data?.message ?? `La API respondió con HTTP ${response.status}.`,
          response.status,
        );
      }
      return data as T;
    } catch (error) {
      if (error instanceof AppError) {
        throw error;
      }
      if (error instanceof Error && error.name === 'AbortError') {
        throw new AppError('La API tardó demasiado en responder.');
      }
      throw new AppError('No se pudo conectar con MobileLab.');
    } finally {
      clearTimeout(timeout);
    }
  }
}
