import { CreateLinkReqSchema, isSafeTargetUrl } from './schemas';

describe('isSafeTargetUrl', () => {
  it.each(['https://example.com', 'http://example.com/path?q=1', 'https://subdomain.example.com'])(
    'accepts public http(s) url %s',
    (value) => {
      expect(isSafeTargetUrl(value)).toBe(true);
    },
  );

  it.each([
    'not-a-url',
    'ftp://example.com/file',
    'http://localhost:3000',
    'http://app.internal/dashboard',
    'http://127.0.0.1/admin',
    'http://10.0.0.5/service',
    'http://169.254.169.254/latest/meta-data',
    'http://[::1]/admin',
    'https://user:pass@example.com/path',
  ])('rejects unsafe target %s', (value) => {
    expect(isSafeTargetUrl(value)).toBe(false);
  });
});

describe('CreateLinkReqSchema', () => {
  it('trims and accepts a public target url', () => {
    const parsed = CreateLinkReqSchema.parse({
      original_url: '  https://example.com/path  ',
    });

    expect(parsed.original_url).toBe('https://example.com/path');
  });

  it('rejects private targets with the frontend validation message', () => {
    const result = CreateLinkReqSchema.safeParse({
      original_url: 'http://127.0.0.1:8080/admin',
    });

    expect(result.success).toBe(false);
    if (result.success) return;

    expect(result.error.issues[0]?.message).toBe('Enter a public http(s) URL');
  });
});
