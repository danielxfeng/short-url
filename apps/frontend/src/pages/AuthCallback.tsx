import { useNavigate, useParams } from 'react-router';
import { toast } from 'sonner';

const AuthCallback = () => {
  const { auth, error } = useParams();
  const navigate = useNavigate();

  if (error) toast.error(`Authentication failed: ${error}`);

  if (auth) {
    localStorage.setItem('auth', auth);
  }

  navigate('/');
  return null;
};

export default AuthCallback;
