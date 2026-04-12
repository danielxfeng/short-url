import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { toast } from 'sonner';
import { getUserInfo } from '@/services';
import { useUser } from '@/hooks/useUser';
import useMutateLink from '@/hooks/useMutateLink';

const AuthCallback = () => {
  const [searchParams] = useSearchParams();
  const auth = searchParams.get('auth');
  const error = searchParams.get('error');
  const login = useUser((s) => s.login);
  const logout = useUser((s) => s.logout);
  const navigate = useNavigate();
  const { clearLinks } = useMutateLink();

  useEffect(() => {
    let isCancelled = false;

    const handleCallback = async () => {
      if (error) {
        toast.error(`Authentication failed: ${error}`);
        logout();
        clearLinks();
        navigate('/', { replace: true });
        return;
      }

      if (!auth) {
        navigate('/', { replace: true });
        return;
      }

      try {
        const user = await getUserInfo(auth);
        if (!user) throw new Error('Failed to retrieve user information after authentication.');

        if (isCancelled) return;

        login(user, auth);
        toast.success('Authentication successful!');
      } catch (err) {
        if (isCancelled) return;

        logout();
        console.error('Error fetching user info:', err);
        toast.error('An error occurred while fetching user information.');
      } finally {
        clearLinks();
      }

      if (!isCancelled) navigate('/', { replace: true });
    };

    void handleCallback();

    return () => {
      isCancelled = true;
    };
  }, [auth, clearLinks, error, login, logout, navigate]);

  return null;
};

export default AuthCallback;
