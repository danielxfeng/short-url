import { fireEvent, render, screen } from '@testing-library/react';

import type { UserRes } from '@/schemas/schemas';
import { ProfileComp } from './Profile';

describe('ProfileComp', () => {
  const user: UserRes = {
    id: 1,
    provider: 'GOOGLE',
    provider_id: 'provider-id',
    display_name: 'Alice',
    profile_pic: null,
  };

  it('renders an accessible profile trigger and menu actions', async () => {
    const handleLogout = vi.fn();
    const handleDeleteAccount = vi.fn().mockResolvedValue(undefined);

    render(
      <ProfileComp
        user={user}
        handleLogout={handleLogout}
        handleDeleteAccount={handleDeleteAccount}
        isDeleting={false}
      />,
    );

    const trigger = screen.getByRole('button', { name: 'Open account menu' });
    expect(trigger).toBeInTheDocument();

    fireEvent.keyDown(trigger, { key: 'Enter' });

    const logoutItem = await screen.findByRole('menuitem', { name: 'Logout' });
    fireEvent.click(logoutItem);

    fireEvent.keyDown(trigger, { key: 'Enter' });

    const deleteItem = await screen.findByRole('menuitem', { name: 'Delete Account' });
    fireEvent.click(deleteItem);

    const confirmDeleteButton = await screen.findByRole('button', { name: 'Delete' });
    fireEvent.click(confirmDeleteButton);

    expect(handleLogout).toHaveBeenCalledTimes(1);
    expect(handleDeleteAccount).toHaveBeenCalledTimes(1);
  });
});
