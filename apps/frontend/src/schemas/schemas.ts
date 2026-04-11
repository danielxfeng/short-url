import * as z from 'zod';

export const OauthProviderValues = ['GOOGLE', 'GITHUB'] as const;
export const OauthProviderEnum = z.enum(OauthProviderValues);
export type OauthProvider = z.infer<typeof OauthProviderEnum>;

export const UserResSchema = z.object({
  id: z.int32(),
  provider: OauthProviderEnum,
  provider_id: z.string(),
  display_name: z.string().nullish(),
  profile_pic: z.string().nullish(),
});
export type UserRes = z.infer<typeof UserResSchema>;

const isAlphaNumericDash = (str: string) => /^[a-zA-Z0-9-]+$/.test(str);

export const CodeSchema = z
  .string()
  .trim()
  .min(1)
  .max(255)
  .nullish()
  .refine((val) => val === null || val === undefined || isAlphaNumericDash(val), {
    message: 'Code can only contain letters, numbers, and dashes',
  });

export const CreateLinkReqSchema = z.object({
  original_url: z.url(),
  code: CodeSchema,
  note: z.string().trim().max(255).nullish(),
});
export type CreateLinkReq = z.infer<typeof CreateLinkReqSchema>;

export const LinkResSchema = z.object({
  id: z.int32(),
  code: z.string(),
  original_url: z.url(),
  clicks: z.int32(),
  note: z.string().nullish(),
  created_at: z.iso.datetime({ offset: true }),
  is_deleted: z.boolean(),
});
export type LinkRes = z.infer<typeof LinkResSchema>;

export const LinksResSchema = z.object({
  links: z.array(LinkResSchema),
  has_more: z.boolean(),
  cursor: z.int32().nullish(),
});
export type LinksRes = z.infer<typeof LinksResSchema>;
