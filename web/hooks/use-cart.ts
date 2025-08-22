import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { cartAPI, type CartResponse, type AddToCartRequest } from "@/lib/api";
import { toast } from "sonner";

// Query keys for cart
export const cartKeys = {
  all: ["cart"] as const,
  lists: () => [...cartKeys.all, "list"] as const,
  list: (filters: string) => [...cartKeys.lists(), { filters }] as const,
  details: () => [...cartKeys.all, "detail"] as const,
  detail: (id: string) => [...cartKeys.details(), id] as const,
  check: (audiobookId: string) =>
    [...cartKeys.all, "check", audiobookId] as const,
};

// Hook to get cart items
export const useCart = () => {
  return useQuery({
    queryKey: cartKeys.lists(),
    queryFn: async () => {
      const response = await cartAPI.getCart();
      return response.data;
    },
  });
};

// Hook to check if an audiobook is in cart
export const useIsInCart = (audiobookId: string) => {
  return useQuery({
    queryKey: cartKeys.check(audiobookId),
    queryFn: async () => {
      const response = await cartAPI.checkIfInCart(audiobookId);
      return response.data?.is_in_cart || false;
    },
    enabled: !!audiobookId,
  });
};

// Hook to add item to cart
export const useAddToCart = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: AddToCartRequest) => {
      const response = await cartAPI.addToCart(data);
      return response;
    },
    onSuccess: () => {
      // Invalidate and refetch cart data
      queryClient.invalidateQueries({ queryKey: cartKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: cartKeys.check(data.audiobook_id),
      });
      toast.success("Added to cart successfully");
    },
    onError: (error: any) => {
      console.error("Failed to add to cart:", error);
      toast.error(error.response?.data?.error || "Failed to add to cart");
    },
  });
};

// Hook to remove item from cart
export const useRemoveFromCart = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (audiobookId: string) => {
      const response = await cartAPI.removeFromCart(audiobookId);
      return response;
    },
    onSuccess: (_, audiobookId) => {
      // Invalidate and refetch cart data
      queryClient.invalidateQueries({ queryKey: cartKeys.lists() });
      queryClient.invalidateQueries({ queryKey: cartKeys.check(audiobookId) });
      toast.success("Removed from cart successfully");
    },
    onError: (error: any) => {
      console.error("Failed to remove from cart:", error);
      toast.error(error.response?.data?.error || "Failed to remove from cart");
    },
  });
};
