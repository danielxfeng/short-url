import { UserResSchema, type UserRes } from '@/schemas/schemas';
import { fetchApi } from './service';

const getUserInfo = async (token: string) => {
  return fetchApi<undefined, UserRes>({
    path: 'user/me',
    method: 'GET',
    isAuthRequired: true,
    injectedToken: token,
    schema: UserResSchema,
  });
};

const deleteUser = async () => {
  return fetchApi<undefined, undefined>({
    path: 'user/me',
    method: 'DELETE',
    isAuthRequired: true,
  });
};

export { getUserInfo, deleteUser };
