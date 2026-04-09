import { CreateLinkReqSchema } from '@/schemas/schemas';
import { useForm } from '@tanstack/react-form';
import useMutateLink from './useMutateLink';

const useAddLinkForm = () => {
  const mutation = useMutateLink();

  const form = useForm({
    defaultValues: {
      original_url: '',
    },
    validators: {
      onChange: CreateLinkReqSchema,
    },
    onSubmit: async ({ value, formApi }) => {
      try {
        await mutation.addLink(value.original_url);
        formApi.reset();
      } catch {
        // Error handling is done in the mutation's onError callback
      }
    },
  });

  return {
    form,
    isPending: mutation.isPending,
  };
};

export default useAddLinkForm;
