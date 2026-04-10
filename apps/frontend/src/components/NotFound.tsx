import { useLocation } from 'react-router';

interface NotFoundCompProps {
  pathName: string;
}

export const NotFoundComp = ({ pathName }: NotFoundCompProps) => {
  return (
    <div className='w-full min-h-dvh flex flex-col items-center justify-center'>
      <h1 className='text-2xl font-bold mb-4'>404 - Not Found</h1>
      <p className='text-muted-foreground'>No page found for: {pathName}</p>
    </div>
  );
};

const NotFound = () => {
  const { pathname } = useLocation();

  return <NotFoundComp pathName={pathname} />;
};

export default NotFound;
