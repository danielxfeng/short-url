import Header from './components/shared/Header';
import Footer from './components/shared/Footer';
import HomePage from './pages/homepage';
import AuthGuard from './components/shared/AuthGuard';
import { Outlet, Route, Routes } from 'react-router';
import AuthCallback from './pages/AuthCallback';
import NotFound from './pages/NotFound';

const Layout = () => {
  return (
    <div className='min-h-dvh w-full flex flex-col'>
      <header className='h-16 flex shrink-0 items-center justify-center border-b'>
        <Header />
      </header>
      <main className='flex-1 flex flex-col'>
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
      <Route element={<Layout />}>
        <Route path='/auth/callback' element={<AuthCallback />} />
        <Route element={<AuthGuard />}>
          <Route path='/' element={<HomePage />} />
        </Route>
        <Route path='*' element={<NotFound />} />
      </Route>
    </Routes>
  );
};

export default App;
