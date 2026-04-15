import { CreateLinkReqSchema, type LinkRes } from '@/schemas/schemas';
import { useForm } from '@tanstack/react-form';
import useMutateLink from './useMutateLink';
import { toast } from 'sonner';
import { useState } from 'react';
import { ApiError } from '@/services/service';

const useAddLinkForm = () => {
  const [addedLink, setAddedLink] = useState<LinkRes | undefined | null>(null);
  const mutation = useMutateLink();

  const form = useForm({
    defaultValues: {
      originalUrl: '',
      code: undefined as string | undefined | null,
      note: undefined as string | undefined | null,
    },
    validators: {
      onSubmit: CreateLinkReqSchema,

      // I put the business logic here, until I find a better way
      // to handle the errors coming from the server.
      // ref: https://www.answeroverflow.com/m/1192055132851032125
      onSubmitAsync: async ({ value, formApi }) => {
        try {
          const parsed = CreateLinkReqSchema.parse(value);

          formApi.setFieldValue('originalUrl', parsed.originalUrl);
          formApi.setFieldValue('code', parsed.code ?? undefined);
          formApi.setFieldValue('note', parsed.note ?? undefined);

          const result = await mutation.addLink(parsed);
          setAddedLink(result);
        } catch (error: unknown) {
          if (error instanceof ApiError && error.status === 409) {
            return {
              fields: {
                code: error.message || 'Code already exists. Please choose another one.',
              },
            };
          }

          let msg = 'Failed to add link. Please try again.';
          if (error instanceof Error && error.message) msg = error.message;

          return { fields: { originalUrl: msg } };
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
