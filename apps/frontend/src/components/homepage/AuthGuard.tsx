import config from '@/config/config';
import { useUser } from '@/hooks/useUser';
import { OauthProviderValues, type OauthProvider } from '@/schemas/schemas';
import { Tooltip, TooltipContent, TooltipTrigger } from '../../components/ui/tooltip';
import { FaGithub, FaGoogle } from 'react-icons/fa';

interface LinkCompProps {
  provider: OauthProvider;
  icon: React.ReactNode;
}

const getAuthUrl = (provider: OauthProvider) =>
  new URL(`/user/auth/${provider.toLowerCase()}`, config.apiBaseUrl).toString();

export const LinkComp = ({ provider, icon }: LinkCompProps) => {
  const label = provider.charAt(0) + provider.slice(1).toLowerCase();

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <a
          href={getAuthUrl(provider)}
          className='inline-flex size-12 items-center justify-center rounded-md transition-colors hover:bg-accent'
          aria-label={`Log in with ${label}`}
          rel='noopener noreferrer'
        >
          {icon}
        </a>
      </TooltipTrigger>
      <TooltipContent>
        <p>Log in with {label}</p>
      </TooltipContent>
    </Tooltip>
  );
};

const getProviderIcon = (provider: OauthProvider) => {
  switch (provider) {
    case 'GOOGLE':
      return <FaGoogle className='size-8' />;
    case 'GITHUB':
      return <FaGithub className='size-8' />;
    default:
      return null;
  }
};

export const LoginComp = () => {
  return (
    <section className='flex flex-col gap-6 mt-6'>
      <h2>Please log in to continue:</h2>
      <div className='flex w-full gap-6'>
        {OauthProviderValues.map((provider) => (
          <LinkComp key={provider} provider={provider} icon={getProviderIcon(provider)} />
        ))}
      </div>
    </section>
  );
};

const AuthGuard = ({ children }: React.PropsWithChildren) => {
  const token = useUser((s) => s.token);
  if (!token) {
    return <LoginComp />;
  }

  return children;
};

export default AuthGuard;
