import config from '@/config/config';
import { useUser } from '@/hooks/useUser';
import * as z from 'zod';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const method = ['GET', 'POST', 'PUT', 'DELETE'] as const;
export type HttpMethod = (typeof method)[number];

const QueryParamValue = z.union([z.string(), z.number(), z.boolean(), z.null(), z.undefined()]);
export type QueryParamValue = z.infer<typeof QueryParamValue>;

export class ApiError extends Error {
  status: number;
  body?: unknown;

  constructor(message: string, status: number, body?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

interface FetchApiParams<Tbody, Tresponse> {
  path: string;
  isAuthRequired?: boolean;
  method?: HttpMethod;
  schema?: z.ZodSchema<Tresponse>;
  searchParams?: Record<string, QueryParamValue>;
  body?: Tbody;
  injectedToken?: string;
}

export const fetchApi = async <Tbody, Tresponse>(
  params: FetchApiParams<Tbody, Tresponse>,
): Promise<Tresponse> => {
  const { path, isAuthRequired = false, method = 'GET', schema, searchParams, body } = params;

  const url = new URL(path, config.apiBaseUrl);

  if (searchParams) {
    Object.entries(searchParams).forEach(([key, value]) => {
      if (value !== null && value !== undefined) {
        url.searchParams.set(key, String(value));
      }
    });
  }

  const options: RequestInit = {
    method: method,
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (isAuthRequired) {
    const token = params.injectedToken ? params.injectedToken : useUser.getState().token;
    if (!token) {
      window.location.href = '/';
      throw new Error('Unauthorized');
    }

    options.headers = {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    };
  }

  if (body !== undefined) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(url.toString(), options);

  if (!response.ok) {
    if (response.status === 401) {
      useUser.getState().logout();
      window.location.href = '/';
      throw new Error('Unauthorized');
    }

    let errorBody: unknown;
    let errorMessage = `Request failed for ${method} ${url.toString()}: ${response.statusText}`;

    try {
      errorBody = await response.clone().json();
      if (
        errorBody &&
        typeof errorBody === 'object' &&
        'error' in errorBody &&
        typeof errorBody.error === 'string'
      ) {
        errorMessage = errorBody.error;
      }
    } catch {
      // Ignore non-JSON error bodies and fall back to the default message.
    }

    throw new ApiError(errorMessage, response.status, errorBody);
  }

  if (!schema) return undefined as unknown as Tresponse;

  return schema.parse(await response.json());
};
