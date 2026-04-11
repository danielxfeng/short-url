import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';

import useAddLinkForm from '@/hooks/useAddLinkForm';
import { Copy, Send } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

const handleError = (error: unknown): { message: string } | undefined => {
  if (error instanceof Object && 'message' in error && typeof error.message === 'string')
    return { message: error.message };

  if (typeof error === 'string') return { message: error };

  if (error instanceof Error) return { message: error.message };

  return { message: 'An unknown error occurred' };
};

interface CreateLinkFormCompProps {
  form: ReturnType<typeof useAddLinkForm>['form'];
  isPending: boolean;
  addedLink: ReturnType<typeof useAddLinkForm>['addedLink'];
  shortLink: string;
  copied: boolean;
  handleCopy: () => Promise<void>;
}

export const CreateLinkFormComp = ({
  form,
  isPending,
  addedLink,
  shortLink,
  copied,
  handleCopy,
}: CreateLinkFormCompProps) => {
  return (
    <Card className='w-full'>
      <CardHeader>
        <CardTitle>Add a link</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          id='new-link-form'
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
          className='flex gap-4'
        >
          <form.Field
            name='original_url'
            children={(field) => {
              const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor={field.name} className='sr-only'>
                    Original URL
                  </FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    placeholder='https://danielslab.dev'
                    autoComplete='off'
                  />
                  {isInvalid && (
                    <FieldError errors={field.state.meta.errors.map((e) => handleError(e))} />
                  )}
                </Field>
              );
            }}
          />

          <Button
            type='submit'
            form='new-link-form'
            disabled={isPending || form.state.isSubmitting}
          >
            <Send size={16} />
          </Button>
        </form>
      </CardContent>

      {/* Added link block */}
      {addedLink && (
        <CardFooter className='border-t'>
          <div className='flex w-full items-center justify-between gap-4 rounded-lg border bg-muted/40 p-3'>
            <div>
              <p className='text-xs uppercase tracking-wide text-muted-foreground'>Added link</p>
              <a
                href={shortLink}
                target='_blank'
                rel='noopener noreferrer'
                className='block truncate font-medium text-foreground underline-offset-4 hover:underline'
              >
                {shortLink}
              </a>
            </div>
            <Button type='button' variant='outline' size='sm' onClick={() => void handleCopy()}>
              <Copy size={16} />
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>
        </CardFooter>
      )}
    </Card>
  );
};

const CreateLinkForm = () => {
  const { form, isPending, addedLink } = useAddLinkForm();
  const [copied, setCopied] = useState(false);
  const shortLink = addedLink ? `${window.location.origin}/${addedLink.code}` : '';

  const handleCopy = async () => {
    if (!addedLink) return;

    try {
      await navigator.clipboard.writeText(shortLink);
      setCopied(true);
      toast.success('Copied short link to clipboard.');
    } catch {
      toast.error('Failed to copy short link. Please try again.');
    }
  };

  return (
    <CreateLinkFormComp
      form={form}
      isPending={isPending}
      addedLink={addedLink}
      shortLink={shortLink}
      copied={copied}
      handleCopy={handleCopy}
    />
  );
};

export default CreateLinkForm;
