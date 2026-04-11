import { createLink, deleteLink, permanentlyDeleteLink, restoreLink } from '@/services';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { linksQueryOptions } from './useLinks';
import logger from '@/lib/logger';

interface LinkMutationInput {
  urlOrCode: string;
  method: 'create' | 'delete' | 'restore' | 'permanentDelete';
}

const useMutateLink = () => {
  const queryClient = useQueryClient();

  const mutationFn = async ({ urlOrCode, method }: LinkMutationInput) => {
    const trimmed = urlOrCode.trim();

    if (!trimmed) throw new Error('URL is required');

    switch (method) {
      case 'create':
        return createLink(trimmed);
      case 'delete':
        return deleteLink(trimmed);
      case 'restore':
        return restoreLink(trimmed);
      case 'permanentDelete':
        return permanentlyDeleteLink(trimmed);
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

  const addLink = async (url: string) => {
    await mutation.mutateAsync({ urlOrCode: url, method: 'create' });
  };

  const removeLink = async (code: string) => {
    await mutation.mutateAsync({ urlOrCode: code, method: 'delete' });
  };

  const restoreDeleted = async (code: string) => {
    await mutation.mutateAsync({ urlOrCode: code, method: 'restore' });
  };

  const permanentlyDelete = async (code: string) => {
    await mutation.mutateAsync({ urlOrCode: code, method: 'permanentDelete' });
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
