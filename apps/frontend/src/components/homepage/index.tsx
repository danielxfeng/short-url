import AuthGuard from '@/components/homepage/AuthGuard';
import CreateLinkForm from './CreateLinkForm';
import LinksTable from './LinksTable';

const HomePage = () => {
  return (
    <div className='w-full max-w-2xl p-4 flex-1 flex flex-col mt-4 mb-8 gap-4'>
      <h1 className='text-2xl font-bold'>Welcome to the Short URL Service</h1>
      <p className='text-muted-foreground'>Create and manage your short URLs with ease.</p>
      <AuthGuard>
        <CreateLinkForm />
        <LinksTable />
      </AuthGuard>
    </div>
  );
};

export default HomePage;
