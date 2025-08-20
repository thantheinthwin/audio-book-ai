import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { createClient } from "@/lib/supabase/client";
import { testAuth, testAdminAuth } from "@/lib/api";
import { authKeys } from "@/queryKeys";

// Hook to get current session
export function useSession() {
  return useQuery({
    queryKey: authKeys.session(),
    queryFn: async () => {
      const supabase = createClient();
      const {
        data: { session },
        error,
      } = await supabase.auth.getSession();

      if (error) {
        throw error;
      }

      return session;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to get current user
export function useUser() {
  return useQuery({
    queryKey: authKeys.user(),
    queryFn: async () => {
      const supabase = createClient();
      const {
        data: { user },
        error,
      } = await supabase.auth.getUser();

      if (error) {
        throw error;
      }

      return user;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to test authentication
export function useTestAuth() {
  return useMutation({
    mutationFn: testAuth,
    onSuccess: (data) => {
      console.log("Auth test successful:", data);
    },
    onError: (error) => {
      console.error("Auth test failed:", error);
    },
  });
}

// Hook to test admin authentication
export function useTestAdminAuth() {
  return useMutation({
    mutationFn: testAdminAuth,
    onSuccess: (data) => {
      console.log("Admin auth test successful:", data);
    },
    onError: (error) => {
      console.error("Admin auth test failed:", error);
    },
  });
}

// Hook to sign out
export function useSignOut() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const supabase = createClient();
      const { error } = await supabase.auth.signOut();
      if (error) throw error;
    },
    onSuccess: () => {
      // Invalidate all auth-related queries
      queryClient.invalidateQueries({ queryKey: authKeys.all });
    },
  });
}
