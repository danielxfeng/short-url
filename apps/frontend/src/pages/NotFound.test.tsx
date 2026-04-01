import { render, screen } from '@testing-library/react';
import { NotFoundComp } from './NotFound';

describe('NotFoundComp', () => {
  it('renders the provided pathname', () => {
    render(<NotFoundComp pathName='/missing-page' />);

    expect(screen.getByText('404 - Not Found')).toBeInTheDocument();
    expect(screen.getByText('No page found for: /missing-page')).toBeInTheDocument();
  });
});
