import { z } from 'zod';

import { fetchApi } from './service';

const userState = {
  user: null,
  token: null as string | null,
  login: vi.fn(),
  logout: vi.fn(),
};

vi.mock('@/hooks/useUser', () => ({
  useUser: {
    getState: () => userState,
  },
}));

describe('fetchApi', () => {
  const originalFetch = globalThis.fetch;
  const storage = new Map<string, string>();

  beforeEach(() => {
    storage.clear();
    globalThis.fetch = vi.fn();
    userState.user = null;
    userState.token = null;
    userState.login = vi.fn();
    userState.logout = vi.fn();
    vi.stubGlobal('localStorage', {
      getItem: vi.fn((key: string) => storage.get(key) ?? null),
      setItem: vi.fn((key: string, value: string) => {
        storage.set(key, value);
      }),
      removeItem: vi.fn((key: string) => {
        storage.delete(key);
      }),
      clear: vi.fn(() => {
        storage.clear();
      }),
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('assembles the request URL with search params and parses the response', async () => {
    const schema = z.object({
      ok: z.boolean(),
    });

    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    const result = await fetchApi<undefined, { ok: boolean }>({
      path: '/links',
      searchParams: {
        cursor: 123,
        active: true,
        empty: undefined,
        removed: null,
      },
      schema,
    });

    expect(globalThis.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/links?cursor=123&active=true',
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      },
    );
    expect(result).toEqual({ ok: true });
  });

  it('adds auth and serializes the request body', async () => {
    const schema = z.object({
      id: z.number(),
    });

    userState.token = 'token-123';
    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(JSON.stringify({ id: 1 }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    const body = { original_url: 'https://example.com' };

    const result = await fetchApi<typeof body, { id: number }>({
      path: '/links',
      method: 'POST',
      isAuthRequired: true,
      body,
      schema,
    });

    expect(globalThis.fetch).toHaveBeenCalledWith('http://localhost:8080/links', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer token-123',
      },
      body: JSON.stringify(body),
    });
    expect(result).toEqual({ id: 1 });
  });

  it('uses an injected token for authenticated requests', async () => {
    const schema = z.object({
      id: z.number(),
    });

    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(JSON.stringify({ id: 1 }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    const result = await fetchApi<undefined, { id: number }>({
      path: '/user/me',
      isAuthRequired: true,
      injectedToken: 'callback-token',
      schema,
    });

    expect(globalThis.fetch).toHaveBeenCalledWith('http://localhost:8080/user/me', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer callback-token',
      },
    });
    expect(result).toEqual({ id: 1 });
  });

  it('returns undefined when no schema is provided', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(null, {
        status: 204,
      }),
    );

    const result = await fetchApi<undefined, undefined>({
      path: '/links/test',
      method: 'DELETE',
    });

    expect(result).toBeUndefined();
  });

  it('throws on non-OK responses', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: 'boom' }), {
        status: 500,
        statusText: 'Internal Server Error',
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    await expect(
      fetchApi<undefined, { ok: boolean }>({
        path: '/links',
        schema: z.object({ ok: z.boolean() }),
      }),
    ).rejects.toThrow('Request failed for GET http://localhost:8080/links: Internal Server Error');
  });

  it('rethrows when fetch itself rejects', async () => {
    vi.mocked(globalThis.fetch).mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(
      fetchApi<undefined, { ok: boolean }>({
        path: '/links',
        schema: z.object({ ok: z.boolean() }),
      }),
    ).rejects.toBeInstanceOf(TypeError);
  });

  it('throws a SyntaxError when the response body is not valid JSON', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response('not-json', {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    await expect(
      fetchApi<undefined, { ok: boolean }>({
        path: '/links',
        schema: z.object({ ok: z.boolean() }),
      }),
    ).rejects.toBeInstanceOf(SyntaxError);
  });

  it('throws a ZodError when schema parsing fails', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValue(
      new Response(JSON.stringify({ ok: 'yes' }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );

    await expect(
      fetchApi<undefined, { ok: boolean }>({
        path: '/links',
        schema: z.object({ ok: z.boolean() }),
      }),
    ).rejects.toBeInstanceOf(z.ZodError);
  });
});
