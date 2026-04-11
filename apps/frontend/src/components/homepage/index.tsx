import AuthGuard from '@/components/homepage/AuthGuard';
import CreateLinkForm from './CreateLinkForm';
import LinksTable from './LinksTable';

const HomePage = () => {
  return (
    <div className='w-full max-w-2xl p-4 flex-1 flex flex-col mt-4 mb-8 gap-6'>
      <h1 className='text-2xl font-bold'>URL Shortener</h1>
      <p className='text-muted-foreground -mt-4'>Create and manage your short URLs</p>
      <AuthGuard>
        <CreateLinkForm />
        <LinksTable />
      </AuthGuard>
    </div>
  );
};

export default HomePage;
