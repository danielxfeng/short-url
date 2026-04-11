import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';

import type { LinkRes } from '@/schemas/schemas';
import { LinkRowComp, LinkTableComp } from './LinksTable';

const activeLink: LinkRes = {
  id: 1,
  code: 'alive123',
  original_url: 'https://example.com/very/long/url',
  clicks: 3,
  created_at: '2026-04-10T22:10:05.91425+03:00',
  is_deleted: false,
};

const deletedLink: LinkRes = {
  id: 2,
  code: 'dead456',
  original_url: 'https://example.com/deleted',
  clicks: 7,
  created_at: '2026-04-10T22:10:05.91425+03:00',
  is_deleted: true,
};

const renderRow = (props: React.ComponentProps<typeof LinkRowComp>) =>
  render(
    <table>
      <tbody>
        <LinkRowComp {...props} />
      </tbody>
    </table>,
  );

describe('LinkRowComp', () => {
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
  const loadMoreRef = { current: null };

  it('shows a loading spinner while fetching with no data', () => {
    render(
      <LinkTableComp
        data={undefined}
        hasNext={false}
        isFetching
        isFetchingNext={false}
        removeLink={async () => undefined}
        restoreDeleted={async () => undefined}
        permanentlyDelete={async () => undefined}
        isPending={false}
        loadMoreRef={loadMoreRef}
      />,
    );

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });

  it('renders the infinite-scroll sentinel when there are more pages', () => {
    render(
      <LinkTableComp
        data={[activeLink, deletedLink]}
        hasNext
        isFetching={false}
        isFetchingNext={false}
        removeLink={async () => undefined}
        restoreDeleted={async () => undefined}
        permanentlyDelete={async () => undefined}
        isPending={false}
        loadMoreRef={loadMoreRef}
      />,
    );

    const table = screen.getByRole('table', { name: 'The list of links.' });
    expect(table.parentElement?.parentElement?.lastElementChild).toBeInTheDocument();
    expect(screen.queryByRole('status', { name: 'Loading' })).not.toBeInTheDocument();
  });

  it('shows a spinner in the sentinel while fetching the next page', () => {
    render(
      <LinkTableComp
        data={[activeLink]}
        hasNext
        isFetching={false}
        isFetchingNext
        removeLink={async () => undefined}
        restoreDeleted={async () => undefined}
        permanentlyDelete={async () => undefined}
        isPending={false}
        loadMoreRef={loadMoreRef}
      />,
    );

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });
});
