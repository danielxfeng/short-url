import {
  LinkResSchema,
  LinksResSchema,
  type CreateLinkReq,
  type LinkRes,
  type LinksRes,
} from '@/schemas/schemas';
import { fetchApi } from './service';

const createLink = async (url: string): Promise<LinkRes> => {
  const body: CreateLinkReq = { original_url: url };

  return fetchApi<CreateLinkReq, LinkRes>({
    path: 'short-urls',
    method: 'POST',
    isAuthRequired: true,
    body,
    schema: LinkResSchema,
  });
};

const getLinks = async (cursor?: number): Promise<LinksRes> => {
  return fetchApi<undefined, LinksRes>({
    path: 'short-urls',
    isAuthRequired: true,
    searchParams: { cursor: cursor },
    schema: LinksResSchema,
  });
};

const deleteLink = async (code: string) => {
  return fetchApi<undefined, undefined>({
    path: `short-urls/${code}`,
    method: 'DELETE',
    isAuthRequired: true,
  });
};

export { createLink, getLinks, deleteLink };
