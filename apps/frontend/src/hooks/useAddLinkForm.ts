import { CreateLinkReqSchema } from '@/schemas/schemas';
import { useForm } from '@tanstack/react-form';
import useMutateLink from './useMutateLink';
import logger from '@/lib/logger';
import { toast } from 'sonner';

const useAddLinkForm = () => {
  const mutation = useMutateLink();

  const form = useForm({
    defaultValues: {
      original_url: '',
    },
    validators: {
      onSubmit: CreateLinkReqSchema,
      onSubmitAsync: async ({ value }) => {
        try {
          await mutation.addLink(value.original_url);
          return undefined;
        } catch (error) {
          logger.error('Failed to add link', error);
          if (error instanceof Error) {
            return {
              fields: {
                original_url: error.message,
              },
            };
          }
        }
      },
    },
    onSubmit: ({ formApi }) => {
      formApi.reset();
      toast.success('Link added successfully!');
    },
  });

  return {
    form,
    isPending: mutation.isPending,
  };
};

export default useAddLinkForm;
