import AuthGuard from '@/components/homepage/AuthGuard';
import CreateLinkForm from './CreateLinkForm';
import LinksTable from './LinksTable';

const HomePage = () => {
  return (
    <div className='w-full max-w-2xl p-4 flex-1 flex flex-col mt-4 mb-8 gap-6'>
      <div className='flex flex-col gap-1'>
        <h1 className='text-2xl font-bold tracking-tight'>Short links</h1>
        <p className='text-muted-foreground'>Create, manage, restore, and remove links.</p>
      </div>
      <AuthGuard>
        <CreateLinkForm />
        <LinksTable />
      </AuthGuard>
    </div>
  );
};

export default HomePage;
