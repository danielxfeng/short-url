import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';

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

describe('CreateLinkFormComp', () => {
  it('renders the input and submits the form', () => {
    const { form, handleSubmit, handleChange, handleBlur } = createFormMock({
      value: 'https://example.com',
    });

    render(<CreateLinkFormComp form={form} isPending={false} />);

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

    render(<CreateLinkFormComp form={form} isPending={false} />);

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid URL');
  });

  it('disables the submit button while pending or submitting', () => {
    const pendingForm = createFormMock().form;
    const submittingForm = createFormMock({ isSubmitting: true }).form;

    const { rerender } = render(<CreateLinkFormComp form={pendingForm} isPending={true} />);
    expect(screen.getByRole('button')).toBeDisabled();

    rerender(<CreateLinkFormComp form={submittingForm} isPending={false} />);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
