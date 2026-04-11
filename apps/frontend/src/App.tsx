import Header from './components/shared/Header';
import Footer from './components/shared/Footer';
import HomePage from './components/homepage';
import { Outlet, Route, Routes } from 'react-router';
import AuthCallback from './components/AuthCallback';
import NotFound from './components/NotFound';

const Layout = () => {
  return (
    <div className='min-h-dvh w-full flex flex-col bg-[radial-gradient(#d4d4d4_1px,transparent_1px)] bg-size-[20px_20px] dark:bg-[radial-gradient(#404040_1px,transparent_1px)]'>
      <header className='h-16 flex shrink-0 items-center justify-center border-b backdrop:blur-sm bg-background/80 sticky top-0 z-10'>
        <Header />
      </header>
      <main className='flex-1 flex flex-col items-center'>
        <Outlet />
      </main>
      <footer className='h-12 flex shrink-0 items-center justify-center border-t'>
        <Footer />
      </footer>
    </div>
  );
};

const App = () => {
  return (
    <Routes>
      <Route path='/auth/callback' element={<AuthCallback />} />

      <Route element={<Layout />}>
        <Route path='/' element={<HomePage />} />
      </Route>

      <Route path='*' element={<NotFound />} />
    </Routes>
  );
};

export default App;
