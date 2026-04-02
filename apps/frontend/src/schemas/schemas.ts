import * as z from 'zod';

const oauthProviderValues = ['GOOGLE', 'GITHUB'] as const;
export const OauthProviderEnum = z.enum(oauthProviderValues);
export type OauthProvider = z.infer<typeof OauthProviderEnum>;

export const UserResSchema = z.object({
  id: z.int32(),
  provider: OauthProviderEnum,
  provider_id: z.string(),
  display_name: z.string().nullish(),
  profile_pic: z.string().nullish(),
});
export type UserRes = z.infer<typeof UserResSchema>;

export const CreateLinkReqSchema = z.object({
  original_url: z.url(),
});
export type CreateLinkReq = z.infer<typeof CreateLinkReqSchema>;

export const LinkResSchema = z.object({
  id: z.int32(),
  code: z.string(),
  original_url: z.url(),
  clicks: z.int32(),
  created_at: z.iso.datetime(),
  is_deleted: z.boolean(),
});
export type LinkRes = z.infer<typeof LinkResSchema>;

export const LinksResSchema = z.object({
  links: z.array(LinkResSchema),
  has_more: z.boolean(),
  cursor: z.int32().nullish(),
});
export type LinksRes = z.infer<typeof LinksResSchema>;
