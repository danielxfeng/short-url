import * as z from 'zod';

export const OauthProviderValues = ['GOOGLE', 'GITHUB'] as const;
export const OauthProviderEnum = z.enum(OauthProviderValues);
export type OauthProvider = z.infer<typeof OauthProviderEnum>;

export const UserResSchema = z.object({
  id: z.int32(),
  provider: OauthProviderEnum,
  providerId: z.string(),
  displayName: z.string().nullish(),
  profilePic: z.string().nullish(),
});
export type UserRes = z.infer<typeof UserResSchema>;

const isAlphaNumericDash = (str: string) => /^[a-zA-Z0-9-]+$/.test(str);

const blockedHostnameSuffixes = ['.localhost', '.local', '.internal', '.home', '.lan'];
const blockedIPv4Cidrs = [
  { base: '100.64.0.0', prefix: 10 },
  { base: '198.18.0.0', prefix: 15 },
];

const ipv4ToInt = (ip: string) =>
  ip.split('.').reduce((acc, octet) => (acc << 8) + Number.parseInt(octet, 10), 0);

const isIPv4 = (host: string) => {
  const parts = host.split('.');
  if (parts.length !== 4) return false;
  return parts.every((part) => /^\d+$/.test(part) && Number(part) >= 0 && Number(part) <= 255);
};

const isIPv6 = (host: string) => host.includes(':');

const isBlockedHostname = (host: string) => {
  if (host === 'localhost' || host === 'local') return true;
  return blockedHostnameSuffixes.some((suffix) => host.endsWith(suffix));
};

const isBlockedIPv4 = (host: string) => {
  const value = ipv4ToInt(host);

  const firstOctet = Number(host.split('.')[0]);
  const secondOctet = Number(host.split('.')[1]);

  if (firstOctet === 0 || firstOctet === 10 || firstOctet === 127) return true;
  if (firstOctet === 169 && secondOctet === 254) return true;
  if (firstOctet === 172 && secondOctet >= 16 && secondOctet <= 31) return true;
  if (firstOctet === 192 && secondOctet === 168) return true;
  if (firstOctet >= 224) return true;

  return blockedIPv4Cidrs.some(({ base, prefix }) => {
    const mask = (0xffffffff << (32 - prefix)) >>> 0;
    return (value & mask) === (ipv4ToInt(base) & mask);
  });
};

const isBlockedIPv6 = (host: string) => {
  const normalized = host.toLowerCase().replace(/^\[/, '').replace(/\]$/, '');
  return (
    normalized === '::' ||
    normalized === '::1' ||
    normalized.startsWith('fe80:') ||
    normalized.startsWith('fc') ||
    normalized.startsWith('fd') ||
    normalized.startsWith('ff')
  );
};

export const isSafeTargetUrl = (value: string) => {
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    return false;
  }

  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return false;
  if (parsed.username || parsed.password) return false;

  const host = parsed.hostname.trim().toLowerCase().replace(/\.+$/, '');
  if (!host) return false;
  if (isBlockedHostname(host)) return false;
  if (isIPv4(host)) return !isBlockedIPv4(host);
  if (isIPv6(host)) return !isBlockedIPv6(host);

  return true;
};

const SafeTargetUrlSchema = z
  .string()
  .trim()
  .refine(isSafeTargetUrl, { message: 'Enter a public http(s) URL' });

export const CodeSchema = z
  .union([z.string().trim().min(1).max(255), z.null(), z.undefined()])
  .refine((val) => val === null || val === undefined || isAlphaNumericDash(val), {
    message: 'Code can only contain letters, numbers, and dashes',
  });

export const CreateLinkReqSchema = z.object({
  originalUrl: SafeTargetUrlSchema,
  code: CodeSchema,
  note: z.union([z.string().trim().max(255), z.null(), z.undefined()]),
});
export type CreateLinkReq = z.infer<typeof CreateLinkReqSchema>;

export const LinkResSchema = z.object({
  id: z.int32(),
  code: z.string(),
  originalUrl: z.url(),
  clicks: z.int32(),
  note: z.string().nullish(),
  createdAt: z.iso.datetime({ offset: true }),
  isDeleted: z.boolean(),
});
export type LinkRes = z.infer<typeof LinkResSchema>;

export const LinksResSchema = z.object({
  links: z.array(LinkResSchema),
  hasMore: z.boolean(),
  cursor: z.int32().nullish(),
});
export type LinksRes = z.infer<typeof LinksResSchema>;
