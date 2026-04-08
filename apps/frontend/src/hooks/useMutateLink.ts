import { createLink, deleteLink } from '@/services';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { linksQueryOptions } from './useLinks';

interface LinkMutationInput {
  url: string;
  method: 'create' | 'delete';
}

const useMutateLink = () => {
  const queryClient = useQueryClient();

  const mutationFn = async ({ url, method }: LinkMutationInput) => {
    const trimmed = url.trim();

    if (!trimmed) throw new Error('URL is required');

    return method === 'create' ? createLink(trimmed) : deleteLink(trimmed);
  };

  const mutation = useMutation({
    mutationFn,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linksQueryOptions().queryKey });
    },
  });

  const addLink = (url: string) => {
    mutation.mutate({ url, method: 'create' });
  };

  const removeLink = (url: string) => {
    mutation.mutate({ url, method: 'delete' });
  };

  return {
    addLink,
    removeLink,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error,
  };
};

export default useMutateLink;
