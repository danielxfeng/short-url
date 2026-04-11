import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';

import type { LinkRes } from '@/schemas/schemas';
import { CreateLinkFormComp } from './CreateLinkForm';

type MockField = {
  name: string;
  state: {
    value: string;
    meta: {
      isTouched: boolean;
      isValid: boolean;
      errors: unknown[];
    };
  };
  handleBlur: ReturnType<typeof vi.fn>;
  handleChange: ReturnType<typeof vi.fn>;
};

const createFormMock = (overrides?: {
  value?: string;
  isTouched?: boolean;
  isValid?: boolean;
  errors?: unknown[];
  isSubmitting?: boolean;
}) => {
  const handleSubmit = vi.fn();
  const handleBlur = vi.fn();
  const handleChange = vi.fn();

  const field: MockField = {
    name: 'original_url',
    state: {
      value: overrides?.value ?? '',
      meta: {
        isTouched: overrides?.isTouched ?? false,
        isValid: overrides?.isValid ?? true,
        errors: overrides?.errors ?? [],
      },
    },
    handleBlur,
    handleChange,
  };

  const form = {
    handleSubmit,
    state: {
      isSubmitting: overrides?.isSubmitting ?? false,
    },
    Field: ({ children }: { name: string; children: (field: MockField) => ReactNode }) =>
      children(field),
  };

  return { form: form as never, handleSubmit, handleBlur, handleChange };
};

const addedLink: LinkRes = {
  id: 1,
  code: 'abc123',
  original_url: 'https://example.com',
  clicks: 0,
  created_at: '2026-04-10T22:10:05.91425+03:00',
  is_deleted: false,
};

describe('CreateLinkFormComp', () => {
  const defaultHandleCopy = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    defaultHandleCopy.mockClear();
  });

  it('renders the input and submits the form', () => {
    const { form, handleSubmit, handleChange, handleBlur } = createFormMock({
      value: 'https://example.com',
    });

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={null}
        shortLink=''
        copied={false}
        handleCopy={defaultHandleCopy}
      />,
    );

    const input = screen.getByLabelText('Original URL');
    expect(input).toHaveValue('https://example.com');

    fireEvent.change(input, { target: { value: 'https://openai.com' } });
    fireEvent.blur(input);
    fireEvent.submit(input.closest('form')!);

    expect(handleChange).toHaveBeenCalledWith('https://openai.com');
    expect(handleBlur).toHaveBeenCalledTimes(1);
    expect(handleSubmit).toHaveBeenCalledTimes(1);
  });

  it('shows validation errors when the field is touched and invalid', () => {
    const { form } = createFormMock({
      isTouched: true,
      isValid: false,
      errors: ['Invalid URL'],
    });

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={null}
        shortLink=''
        copied={false}
        handleCopy={defaultHandleCopy}
      />,
    );

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid URL');
  });

  it('disables the submit button while pending or submitting', () => {
    const pendingForm = createFormMock().form;
    const submittingForm = createFormMock({ isSubmitting: true }).form;

    const { rerender } = render(
      <CreateLinkFormComp
        form={pendingForm}
        isPending={true}
        addedLink={null}
        shortLink=''
        copied={false}
        handleCopy={defaultHandleCopy}
      />,
    );
    expect(screen.getByRole('button')).toBeDisabled();

    rerender(
      <CreateLinkFormComp
        form={submittingForm}
        isPending={false}
        addedLink={null}
        shortLink=''
        copied={false}
        handleCopy={defaultHandleCopy}
      />,
    );
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('shows the created link block and copies the short link', async () => {
    const { form } = createFormMock();
    const handleCopy = vi.fn().mockResolvedValue(undefined);

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={addedLink}
        shortLink='http://localhost:3000/abc123'
        copied={false}
        handleCopy={handleCopy}
      />,
    );

    const resultLink = screen.getByRole('link', { name: 'http://localhost:3000/abc123' });
    expect(screen.getByText('Added link')).toBeInTheDocument();
    expect(resultLink).toHaveAttribute('href', 'http://localhost:3000/abc123');

    fireEvent.click(screen.getByRole('button', { name: 'Copy' }));

    await waitFor(() => {
      expect(handleCopy).toHaveBeenCalledTimes(1);
    });
  });

  it('shows copied state from props', () => {
    const { form } = createFormMock();

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={addedLink}
        shortLink='http://localhost:3000/abc123'
        copied
        handleCopy={defaultHandleCopy}
      />,
    );

    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();
  });
});
