import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { Dispatch, SetStateAction } from 'react';

import { TooltipProvider } from '@/components/ui/tooltip';
import type { LinkRes } from '@/schemas/schemas';
import { LinkRowComp, LinkTableComp } from './LinksTable';

const activeLink: LinkRes = {
  id: 1,
  code: 'alive123',
  originalUrl: 'https://example.com/very/long/url',
  clicks: 3,
  createdAt: '2026-04-10T22:10:05.91425+03:00',
  isDeleted: false,
};

const deletedLink: LinkRes = {
  id: 2,
  code: 'dead456',
  originalUrl: 'https://example.com/deleted',
  clicks: 7,
  createdAt: '2026-04-10T22:10:05.91425+03:00',
  isDeleted: true,
};

const renderRow = (props: React.ComponentProps<typeof LinkRowComp>) =>
  render(
    <TooltipProvider>
      <table>
        <tbody>
          <LinkRowComp {...props} />
        </tbody>
      </table>
    </TooltipProvider>,
  );

const renderTable = (props: React.ComponentProps<typeof LinkTableComp>) =>
  render(
    <TooltipProvider>
      <LinkTableComp {...props} />
    </TooltipProvider>,
  );

describe('LinkRowComp', () => {
  it('opens the short link in a new tab without suppressing referrer at the anchor level', () => {
    renderRow({
      link: activeLink,
      removeLink: async () => undefined,
      restoreDeleted: async () => undefined,
      permanentlyDelete: async () => undefined,
      isPending: false,
    });

    const link = screen.getByRole('link', { name: 'alive123' });
    expect(link).toHaveAttribute('href', '/alive123');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener');
  });

  it('soft deletes an active link from the expanded row actions', async () => {
    let removedCode: string | null = null;
    let restoredCode: string | null = null;
    let permanentlyDeletedCode: string | null = null;

    renderRow({
      link: activeLink,
      removeLink: async (code) => {
        removedCode = code;
      },
      restoreDeleted: async (code) => {
        restoredCode = code;
      },
      permanentlyDelete: async (code) => {
        permanentlyDeletedCode = code;
      },
      isPending: false,
    });

    const row = screen.getByText('alive123').closest('tr');
    if (!row) throw new Error('expected active row');

    fireEvent.click(row);
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));

    await waitFor(() => {
      expect(removedCode).toBe('alive123');
    });
    expect(restoredCode).toBeNull();
    expect(permanentlyDeletedCode).toBeNull();
  });

  it('restores a deleted link from the expanded row actions', async () => {
    let removedCode: string | null = null;
    let restoredCode: string | null = null;
    let permanentlyDeletedCode: string | null = null;

    renderRow({
      link: deletedLink,
      removeLink: async (code) => {
        removedCode = code;
      },
      restoreDeleted: async (code) => {
        restoredCode = code;
      },
      permanentlyDelete: async (code) => {
        permanentlyDeletedCode = code;
      },
      isPending: false,
    });

    const row = screen.getByText('dead456').closest('tr');
    if (!row) throw new Error('expected deleted row');

    fireEvent.click(row);
    fireEvent.click(screen.getByRole('button', { name: 'Restore' }));

    await waitFor(() => {
      expect(restoredCode).toBe('dead456');
    });
    expect(removedCode).toBeNull();
    expect(permanentlyDeletedCode).toBeNull();
  });

  it('permanently deletes a deleted link only after confirmation', async () => {
    let permanentlyDeletedCode: string | null = null;

    renderRow({
      link: deletedLink,
      removeLink: async () => undefined,
      restoreDeleted: async () => undefined,
      permanentlyDelete: async (code) => {
        permanentlyDeletedCode = code;
      },
      isPending: false,
    });

    const row = screen.getByText('dead456').closest('tr');
    if (!row) throw new Error('expected deleted row');

    fireEvent.click(row);
    fireEvent.click(screen.getByRole('button', { name: 'Permanently Delete' }));

    const dialog = await screen.findByRole('alertdialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Permanently Delete' }));

    await waitFor(() => {
      expect(permanentlyDeletedCode).toBe('dead456');
    });
  });
});

describe('LinkTableComp', () => {
  const loadMoreRef: Dispatch<SetStateAction<HTMLDivElement | null>> = () => undefined;

  it('shows a loading spinner while fetching with no data', () => {
    renderTable({
      data: undefined,
      hasNext: false,
      isFetching: true,
      isFetchingNext: false,
      removeLink: async () => undefined,
      restoreDeleted: async () => undefined,
      permanentlyDelete: async () => undefined,
      isPending: false,
      loadMoreRef,
    });

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });

  it('renders the infinite-scroll sentinel when there are more pages', () => {
    renderTable({
      data: [activeLink, deletedLink],
      hasNext: true,
      isFetching: false,
      isFetchingNext: false,
      removeLink: async () => undefined,
      restoreDeleted: async () => undefined,
      permanentlyDelete: async () => undefined,
      isPending: false,
      loadMoreRef,
    });

    const table = screen.getByRole('table', { name: 'The list of links.' });
    expect(table.parentElement?.parentElement?.lastElementChild).toBeInTheDocument();
    expect(screen.queryByRole('status', { name: 'Loading' })).not.toBeInTheDocument();
  });

  it('shows a spinner in the sentinel while fetching the next page', () => {
    renderTable({
      data: [activeLink],
      hasNext: true,
      isFetching: false,
      isFetchingNext: true,
      removeLink: async () => undefined,
      restoreDeleted: async () => undefined,
      permanentlyDelete: async () => undefined,
      isPending: false,
      loadMoreRef,
    });

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });
});
