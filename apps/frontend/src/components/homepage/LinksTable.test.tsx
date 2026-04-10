import { fireEvent, render, screen } from '@testing-library/react';

import type { LinkRes } from '@/schemas/schemas';
import { LinkTableComp } from './LinksTable';

describe('LinkTableComp', () => {
  const renderLinkTable = (props: React.ComponentProps<typeof LinkTableComp>) =>
    render(<LinkTableComp {...props} />);

  const links: LinkRes[] = [
    {
      id: 1,
      code: 'alive123',
      original_url: 'https://example.com/very/long/url',
      clicks: 3,
      created_at: '2026-04-10T22:10:05.91425+03:00',
      is_deleted: false,
    },
    {
      id: 2,
      code: 'dead456',
      original_url: 'https://example.com/deleted',
      clicks: 7,
      created_at: '2026-04-10T22:10:05.91425+03:00',
      is_deleted: true,
    },
  ];

  it('renders rows and disables delete for soft-deleted links', () => {
    const fetchNext = vi.fn();
    const removeLink = vi.fn().mockResolvedValue(undefined);

    renderLinkTable({
      data: links,
      hasNext: false,
      fetchNext,
      isFetching: false,
      isFetchingNext: false,
      removeLink,
      isPending: false,
    });

    expect(screen.getByText('alive123')).toBeInTheDocument();
    expect(screen.getByText('dead456')).toBeInTheDocument();
    expect(screen.getByText('https://example.com/very/long/url')).toHaveClass('truncate');

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    expect(deleteButtons[0]).toBeEnabled();
    expect(deleteButtons[1]).toBeDisabled();

    fireEvent.click(deleteButtons[0]);
    expect(removeLink).toHaveBeenCalledWith('alive123');
  });

  it('shows a loading spinner while fetching with no data', () => {
    renderLinkTable({
      data: undefined,
      hasNext: false,
      fetchNext: vi.fn(),
      isFetching: true,
      isFetchingNext: false,
      removeLink: vi.fn().mockResolvedValue(undefined),
      isPending: false,
    });

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });

  it('renders and triggers the load-more action', () => {
    const fetchNext = vi.fn();

    renderLinkTable({
      data: links,
      hasNext: true,
      fetchNext,
      isFetching: false,
      isFetchingNext: false,
      removeLink: vi.fn().mockResolvedValue(undefined),
      isPending: false,
    });

    const loadMoreButton = screen.getByRole('button', { name: 'Load more' });
    expect(loadMoreButton).toBeEnabled();

    fireEvent.click(loadMoreButton);
    expect(fetchNext).toHaveBeenCalledTimes(1);
  });
});
