import config from '@/config/config';
import { CodeSchema } from '@/schemas/schemas';
import { useEffect } from 'react';
import { useLocation, useSearchParams } from 'react-router';

interface NotFoundCompProps {
  pathName?: string;
}

export const NotFoundComp = ({ pathName }: NotFoundCompProps) => {
  return (
    <div className='w-full min-h-dvh flex flex-col items-center justify-center'>
      <h1 className='text-2xl font-bold mb-4'>404 - Not Found</h1>
      {pathName ? <p className='text-muted-foreground'>No page found for: {pathName}</p> : null}
    </div>
  );
};

const NotFound = () => {
  const [searchParams] = useSearchParams();
  const { pathname } = useLocation();
  const invalidUrl = searchParams.get('invalid-url');
  const code = pathname.slice(1); // Remove leading slash
  const isReservedPath = pathname === '/not-found' || pathname.startsWith('/auth/');
  const isCode = !isReservedPath && CodeSchema.safeParse(code).success;
  const shouldRedirect = !invalidUrl && isCode;

  useEffect(() => {
    if (!shouldRedirect) return;

    const backendRedirectUrl = new URL(
      `short-urls/${encodeURIComponent(code)}`,
      `${config.apiBaseUrl.replace(/\/+$/, '')}/`,
    );

    window.location.replace(backendRedirectUrl.toString());
  }, [code, shouldRedirect]);

  if (invalidUrl) return <NotFoundComp pathName={invalidUrl} />;
  if (!isCode) return <NotFoundComp pathName={isReservedPath ? undefined : pathname} />;

  return null;
};

export default NotFound;
