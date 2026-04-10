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

    expect(
      screen.getByRole('heading', { name: 'Welcome to the Short URL Service' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Create and manage your short URLs with ease.')).toBeInTheDocument();
  });
});
