import { CreateLinkReqSchema } from '@/schemas/schemas';
import { useForm } from '@tanstack/react-form';
import useMutateLink from './useMutateLink';
import { toast } from 'sonner';

const useAddLinkForm = () => {
  const mutation = useMutateLink();

  const form = useForm({
    defaultValues: {
      original_url: '',
    },
    validators: {
      onSubmit: CreateLinkReqSchema,

      // I put the business logic here, until I find a better way
      // to handle the errors coming from the server.
      // ref: https://www.answeroverflow.com/m/1192055132851032125
      onSubmitAsync: async ({ value }) => {
        try {
          await mutation.addLink(value.original_url);
        } catch {
          return {
            fields: {
              original_url: 'Failed to add link. Please try again.',
            },
          };
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
