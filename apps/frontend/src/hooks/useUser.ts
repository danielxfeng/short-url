import type { UserRes } from '@/schemas/schemas';
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface UserStore {
  user: UserRes | null;
  token: string | null;
  login: (user: UserRes, token: string) => void;
  logout: () => void;
}

export const useUser = create<UserStore>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      login: (user, token) => set({ user, token }),
      logout: () => set({ user: null, token: null }),
    }),
    {
      name: 'user-storage',
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
