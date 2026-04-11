import { createLink, deleteLink, permanentlyDeleteLink, restoreLink } from '@/services';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { linksQueryOptions } from './useLinks';
import logger from '@/lib/logger';
import type { CreateLinkReq } from '@/schemas/schemas';

interface LinkMutationInput {
  code?: string;
  req?: CreateLinkReq;
  method: 'create' | 'delete' | 'restore' | 'permanentDelete';
}

const useMutateLink = () => {
  const queryClient = useQueryClient();

  const mutationFn = async ({ code, req, method }: LinkMutationInput) => {
    const trimmed = code?.trim();

    if (method !== 'create' && !trimmed) throw new Error('URL is required');
    if (method === 'create' && !req) throw new Error('Request body is required');

    switch (method) {
      case 'create':
        return createLink(req!);
      case 'delete':
        return deleteLink(trimmed!);
      case 'restore':
        return restoreLink(trimmed!);
      case 'permanentDelete':
        return permanentlyDeleteLink(trimmed!);
      default:
        throw new Error('Invalid method');
    }
  };

  const mutation = useMutation({
    mutationFn,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: linksQueryOptions().queryKey });
    },
    onError: (error) => {
      logger.error('Link mutation error', error);
    },
  });

  const addLink = async (req: CreateLinkReq) => {
    return await mutation.mutateAsync({ req, method: 'create' });
  };

  const removeLink = async (code: string) => {
    await mutation.mutateAsync({ code, method: 'delete' });
  };

  const restoreDeleted = async (code: string) => {
    await mutation.mutateAsync({ code, method: 'restore' });
  };

  const permanentlyDelete = async (code: string) => {
    await mutation.mutateAsync({ code, method: 'permanentDelete' });
  };

  const clearLinks = () => {
    queryClient.removeQueries({ queryKey: linksQueryOptions().queryKey });
  };

  return {
    addLink,
    removeLink,
    restoreDeleted,
    permanentlyDelete,
    clearLinks,
    isPending: mutation.isPending,
  };
};

export default useMutateLink;
