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
  values?: Partial<Record<'originalUrl' | 'code' | 'note', string>>;
  errorsByField?: Partial<Record<'originalUrl' | 'code' | 'note', unknown[]>>;
  touchedByField?: Partial<Record<'originalUrl' | 'code' | 'note', boolean>>;
  validByField?: Partial<Record<'originalUrl' | 'code' | 'note', boolean>>;
}) => {
  const handleSubmit = vi.fn();
  const handleBlur = vi.fn();
  const handleChange = vi.fn();

  const form = {
    handleSubmit,
    state: {
      isSubmitting: overrides?.isSubmitting ?? false,
    },
    Field: ({
      name,
      children,
    }: {
      name: 'originalUrl' | 'code' | 'note';
      children: (field: MockField) => ReactNode;
    }) =>
      children({
        name,
        state: {
          value:
            overrides?.values?.[name] ?? (name === 'originalUrl' ? (overrides?.value ?? '') : ''),
          meta: {
            isTouched: overrides?.touchedByField?.[name] ?? overrides?.isTouched ?? false,
            isValid: overrides?.validByField?.[name] ?? overrides?.isValid ?? true,
            errors: overrides?.errorsByField?.[name] ?? overrides?.errors ?? [],
          },
        },
        handleBlur,
        handleChange,
      }),
  };

  return { form: form as never, handleSubmit, handleBlur, handleChange };
};

const addedLink: LinkRes = {
  id: 1,
  code: 'abc123',
  originalUrl: 'https://example.com',
  clicks: 0,
  createdAt: '2026-04-10T22:10:05.91425+03:00',
  isDeleted: false,
};

describe('CreateLinkFormComp', () => {
  const defaultHandleCopy = vi.fn().mockResolvedValue(undefined);
  const defaultSetCopied = vi.fn();

  beforeEach(() => {
    defaultHandleCopy.mockClear();
    defaultSetCopied.mockClear();
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
        setCopied={defaultSetCopied}
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
    expect(defaultSetCopied).toHaveBeenCalledWith(false);
    expect(handleSubmit).toHaveBeenCalledTimes(1);
  });

  it('shows validation errors when the field is touched and invalid', () => {
    const { form } = createFormMock({
      touchedByField: { originalUrl: true },
      validByField: { originalUrl: false, code: true, note: true },
      errorsByField: { originalUrl: ['Invalid URL'] },
    });

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={null}
        shortLink=''
        copied={false}
        setCopied={defaultSetCopied}
        handleCopy={defaultHandleCopy}
      />,
    );

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid URL');
  });

  it('normalizes empty optional fields to undefined', () => {
    const { form, handleChange } = createFormMock();

    render(
      <CreateLinkFormComp
        form={form}
        isPending={false}
        addedLink={null}
        shortLink=''
        copied={false}
        setCopied={defaultSetCopied}
        handleCopy={defaultHandleCopy}
      />,
    );

    fireEvent.change(screen.getByPlaceholderText('Custom code (optional)'), {
      target: { value: '' },
    });
    fireEvent.change(screen.getByPlaceholderText('Note (optional)'), {
      target: { value: '   ' },
    });

    expect(handleChange).toHaveBeenCalledWith(undefined);
    expect(handleChange).toHaveBeenLastCalledWith(undefined);
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
        setCopied={defaultSetCopied}
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
        setCopied={defaultSetCopied}
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
        setCopied={defaultSetCopied}
        handleCopy={handleCopy}
      />,
    );

    const resultLink = screen.getByRole('link', { name: 'http://localhost:3000/abc123' });
    expect(screen.getByText('Short link')).toBeInTheDocument();
    expect(resultLink).toHaveAttribute('href', 'http://localhost:3000/abc123');
    expect(resultLink).toHaveAttribute('rel', 'noopener');

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
        setCopied={defaultSetCopied}
        handleCopy={defaultHandleCopy}
      />,
    );

    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();
  });
});
