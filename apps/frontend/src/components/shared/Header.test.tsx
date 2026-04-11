import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import Header from './Header';

describe('Header', () => {
  it('renders a home link', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>,
    );

    const link = screen.getByRole('link', { name: 'shorturl' });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', '/');
  });
});
