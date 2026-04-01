import { Link } from 'react-router';

const Header = () => (
  <div className='w-full max-w-2xl flex items-center justify-between px-4 py-2 sticky top-0 z-10'>
    <Link to='/'>Short URL</Link>
  </div>
);

export default Header;
