import { Outlet } from 'react-router';

const AuthGuard = () => {
  // Implement your authentication logic here
  const isAuthenticated = true; // Replace with actual authentication check

  if (!isAuthenticated) {
    return <div>Please log in to access this page.</div>;
  }

  return <Outlet />;
};

export default AuthGuard;
