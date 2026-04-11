import { render, screen } from '@testing-library/react';

import HomePage from './index';

vi.mock('@/components/homepage/AuthGuard', () => ({
  default: ({ children }: React.PropsWithChildren) => <div>{children}</div>,
}));

vi.mock('./CreateLinkForm', () => ({
  default: () => <div>CreateLinkForm</div>,
}));

vi.mock('./LinksTable', () => ({
  default: () => <div>LinksTable</div>,
}));

describe('HomePage', () => {
  it('renders the page heading and supporting text', () => {
    render(<HomePage />);

    expect(screen.getByRole('heading', { name: 'Short links' })).toBeInTheDocument();
    expect(screen.getByText('Create, manage, restore, and remove links.')).toBeInTheDocument();
  });
});
