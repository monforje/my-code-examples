import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  usersProfileMeAvatarDelete,
  usersProfileMeAvatarUpdate,
  usersProfileMeGet,
  usersProfileMeSettingsUpdate,
  type ProfileResponse,
  type UpdateProfileSettingsRequest,
} from "@shared/api";
import { getApiErrorMessage } from "@shared/lib/api-error";

export const profileQueryKey = ["profile", "me"] as const;

export function useProfile(enabled = true) {
  return useQuery({
    queryKey: profileQueryKey,
    enabled,
    retry: false,
    staleTime: 10 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
    queryFn: async () => {
      const res = await usersProfileMeGet();
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить профиль."));
    },
  });
}

export function useUpdateProfileSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateProfileSettingsRequest) => {
      const res = await usersProfileMeSettingsUpdate(input);
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось сохранить профиль."));
    },
    onSuccess: (profile) => {
      queryClient.setQueryData<ProfileResponse>(profileQueryKey, profile);
    },
  });
}

export function useUpdateAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (avatar: File) => {
      const res = await usersProfileMeAvatarUpdate({ avatar });
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить аватар."));
    },
    onSuccess: (avatar) => {
      queryClient.setQueryData<ProfileResponse>(profileQueryKey, (profile) => {
        if (!profile) return profile;
        return {
          ...profile,
          avatar_url: avatar.avatar_url,
          updated_at: avatar.updated_at,
        };
      });
    },
  });
}

export function useDeleteAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const res = await usersProfileMeAvatarDelete();
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось удалить аватар."));
    },
    onSuccess: (avatar) => {
      queryClient.setQueryData<ProfileResponse>(profileQueryKey, (profile) => {
        if (!profile) return profile;
        return {
          ...profile,
          avatar_url: avatar.avatar_url,
          updated_at: avatar.updated_at,
        };
      });
    },
  });
}
