import { createLink, deleteLink } from '@/services';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { linksQueryOptions } from './useLinks';
import logger from '@/lib/logger';

interface LinkMutationInput {
  urlOrCode: string;
  method: 'create' | 'delete';
}

const useMutateLink = () => {
  const queryClient = useQueryClient();

  const mutationFn = async ({ urlOrCode, method }: LinkMutationInput) => {
    const trimmed = urlOrCode.trim();

    if (!trimmed) throw new Error('URL is required');

    return method === 'create' ? createLink(trimmed) : deleteLink(trimmed);
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

  const clearLinks = () => {
    queryClient.removeQueries({ queryKey: linksQueryOptions().queryKey });
  };

  return {
    addLink,
    removeLink,
    clearLinks,
    isPending: mutation.isPending,
  };
};

export default useMutateLink;
