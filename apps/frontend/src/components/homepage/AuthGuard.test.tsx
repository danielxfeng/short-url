import { render, screen } from '@testing-library/react';

import config from '@/config/config';
import { TooltipProvider } from '@/components/ui/tooltip';
import { LinkComp, LoginComp } from './AuthGuard';

describe('AuthGuard UI', () => {
  beforeEach(() => {
    config.apiBaseUrl = 'http://localhost:8080/api/v1';
  });

  it('renders LinkComp with the provider auth URL', () => {
    render(
      <TooltipProvider>
        <LinkComp provider='GOOGLE' icon={<span>G</span>} />
      </TooltipProvider>,
    );

    const link = screen.getByRole('link', { name: 'Log in with Google' });

    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', 'http://localhost:8080/api/v1/user/auth/google');
  });

  it('renders LoginComp with links for all providers', () => {
    render(
      <TooltipProvider>
        <LoginComp />
      </TooltipProvider>,
    );

    expect(screen.getByText('Please log in to continue:')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Log in with Google' })).toHaveAttribute(
      'href',
      'http://localhost:8080/api/v1/user/auth/google',
    );
    expect(screen.getByRole('link', { name: 'Log in with Github' })).toHaveAttribute(
      'href',
      'http://localhost:8080/api/v1/user/auth/github',
    );
  });
});
