import { render, screen } from '@testing-library/react';
import Footer from './Footer';

describe('Footer', () => {
  it('renders the credit link', () => {
    render(<Footer />);

    expect(screen.getByText(/developped with/i)).toBeInTheDocument();

    const link = screen.getByRole('link', { name: "Daniel's Lab" });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', 'https://danielslab.dev');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });
});
