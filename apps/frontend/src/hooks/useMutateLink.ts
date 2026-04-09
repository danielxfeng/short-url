import { createLink, deleteLink } from '@/services';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { linksQueryOptions } from './useLinks';

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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linksQueryOptions().queryKey });
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
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error,
  };
};

export default useMutateLink;
