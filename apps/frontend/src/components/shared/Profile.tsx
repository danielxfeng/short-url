import { useState } from 'react';
import { useNavigate } from 'react-router';

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';

import { useUser } from '@/hooks/useUser';
import type { UserRes } from '@/schemas/schemas';
import { deleteUser } from '@/services';
import logger from '@/lib/logger';
import { toast } from 'sonner';

interface ProfileCompProps {
  user: UserRes;
  handleLogout: () => void;
  handleDeleteAccount: () => Promise<void>;
  isDeleting: boolean;
}

export const ProfileComp = ({
  user,
  handleLogout,
  handleDeleteAccount,
  isDeleting,
}: ProfileCompProps) => (
  <DropdownMenu>
    <DropdownMenuTrigger asChild>
      <Button variant='ghost' size='icon' className='rounded-full' aria-label='Open account menu'>
        {/* User avatar */}
        <Avatar>
          <AvatarImage src={user.profile_pic ?? undefined} />
          <AvatarFallback>{user.display_name?.charAt(0) || 'U'}</AvatarFallback>
        </Avatar>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent className='w-32'>
      <DropdownMenuGroup>
        {/* Logout button */}
        <DropdownMenuItem onSelect={handleLogout}>Logout</DropdownMenuItem>

        {/* Delete account button */}
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <DropdownMenuItem onSelect={(event) => event.preventDefault()} variant='destructive'>
              Delete Account
            </DropdownMenuItem>
          </AlertDialogTrigger>
          <AlertDialogContent size='sm'>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete account?</AlertDialogTitle>
              <AlertDialogDescription>
                This will permanently delete your account and all associated data. This action
                cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel variant='outline'>Cancel</AlertDialogCancel>
              <AlertDialogAction
                variant='destructive'
                onClick={handleDeleteAccount}
                disabled={isDeleting}
              >
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </DropdownMenuGroup>
    </DropdownMenuContent>
  </DropdownMenu>
);

const Profile = () => {
  const [isDeleting, setIsDeleting] = useState(false);
  const navigate = useNavigate();
  const user = useUser((s) => s.user);
  const logout = useUser((s) => s.logout);

  const handleLogout = () => {
    logout();
    toast.success('Logged out successfully, redirecting to homepage');
    navigate('/');
  };

  const handleDeleteAccount = async () => {
    setIsDeleting(true);

    try {
      await deleteUser();
      logout();
      toast.success('Account deleted successfully, redirecting to homepage');
      navigate('/');
    } catch (error) {
      logger.error('Failed to delete account', error);
      toast.error('Failed to delete account, please try again');
    } finally {
      setIsDeleting(false);
    }
  };

  if (!user) return null;

  return (
    <ProfileComp
      user={user}
      handleLogout={handleLogout}
      handleDeleteAccount={handleDeleteAccount}
      isDeleting={isDeleting}
    />
  );
};

export default Profile;
