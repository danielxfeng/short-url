import { CreateLinkReqSchema, type LinkRes } from '@/schemas/schemas';
import { useForm } from '@tanstack/react-form';
import useMutateLink from './useMutateLink';
import { toast } from 'sonner';
import { useState } from 'react';

const useAddLinkForm = () => {
  const [addedLink, setAddedLink] = useState<LinkRes | undefined | null>(null);
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
          const result = await mutation.addLink(value.original_url);
          setAddedLink(result);
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
    addedLink,
  };
};

export default useAddLinkForm;
