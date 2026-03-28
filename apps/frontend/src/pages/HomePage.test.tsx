import { expect, test } from 'vitest';
import { render, screen } from '@testing-library/react';
import HomePage from './HomePage';

test('renders content', () => {
  render(<HomePage />);

  const element = screen.getByText('Home');
  expect(element).toBeDefined();
});
