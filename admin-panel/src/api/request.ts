/* ============================================================
 * Unified HTTP Request Layer
 *
 * - Reads baseURL from VITE_API_BASE_URL (falls back to '' / same-origin)
 * - Injects Authorization header automatically from localStorage
 * - Unified timeout via AbortController (default 15s)
 * - 401 → clear token + redirect to /login
 * - Parses backend response: { data: <T>, error?: string, errCode?: number }
 *   (not the generic { code, data, message } format — that's a different API)
 * ============================================================ */

export class ApiError extends Error {
  declare status: number;
  constructor(
    status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  params?: Record<string, string | number | boolean | undefined>;
  body?: unknown;
  timeout?: number;
}

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';
const DEFAULT_TIMEOUT = 15_000;

function getToken(): string | null {
  try {
    return localStorage.getItem('nd.access');
  } catch {
    return null;
  }
}

function buildUrl(
  path: string,
  params?: Record<string, string | number | boolean | undefined>,
): string {
  const url = new URL(path, BASE_URL);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '' && value !== null) {
        url.searchParams.set(key, String(value));
      }
    });
  }
  return url.toString();
}

function handleUnauthorized() {
  localStorage.removeItem('nd.access');
  localStorage.removeItem('nd.user');
  window.location.href = '/login';
}

async function runRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { params, body, timeout = DEFAULT_TIMEOUT, headers, ...rest } = options;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeout);

  try {
    const token = getToken();
    const response = await fetch(buildUrl(path, params), {
      ...rest,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(headers as Record<string, string>),
      },
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
    });

    if (response.status === 401) {
      handleUnauthorized();
      throw new ApiError(401, '登录已过期');
    }

    if (!response.ok) {
      const errBody = await response.json().catch(() => ({}));
      throw new ApiError(response.status, errBody.error || `HTTP ${response.status}`);
    }

    // 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    const json = await response.json();
    // Backend format: { data: T, error?: string, errCode?: number }
    if (json.error) {
      throw new ApiError(json.errCode ?? 500, json.error);
    }
    return json.data as T;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError(408, '请求超时，请稍后重试');
    }
    throw new ApiError(0, '网络异常，请稍后重试');
  } finally {
    clearTimeout(timer);
  }
}

export const request = {
  get: <T>(path: string, options?: RequestOptions) =>
    runRequest<T>(path, { ...options, method: 'GET' }),

  post: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    runRequest<T>(path, { ...options, method: 'POST', body }),

  put: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    runRequest<T>(path, { ...options, method: 'PUT', body }),

  delete: <T>(path: string, options?: RequestOptions) =>
    runRequest<T>(path, { ...options, method: 'DELETE' }),

  patch: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    runRequest<T>(path, { ...options, method: 'PATCH', body }),
};