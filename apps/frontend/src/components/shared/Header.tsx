import { Link } from 'react-router';
import Profile from './Profile';

const Header = () => (
  <div className='w-full max-w-2xl flex items-center justify-between px-4 py-2'>
    <Link to='/' className='inline-flex items-center gap-2'>
      <span className='font-semibold tracking-tight text-lg'>
        <span className='text-foreground'>short</span>
        <span className='text-sky-600'>url</span>
      </span>
    </Link>
    <Profile />
  </div>
);

export default Header;
