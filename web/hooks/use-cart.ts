import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { cartAPI, type AddToCartRequest } from "@/lib/api";
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
    onSuccess: (response, variables) => {
      // Invalidate and refetch cart data
      queryClient.invalidateQueries({ queryKey: cartKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: cartKeys.check(variables.audiobook_id),
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

// Checkout hooks
import { checkoutAPI } from "@/lib/api";

// Query keys for checkout
export const checkoutKeys = {
  all: ["checkout"] as const,
  purchaseHistory: () => [...checkoutKeys.all, "purchaseHistory"] as const,
  purchaseHistoryList: (filters: string) => [...checkoutKeys.purchaseHistory(), { filters }] as const,
  isPurchased: (audiobookId: string) => [...checkoutKeys.all, "isPurchased", audiobookId] as const,
};

// Hook to checkout cart items
export const useCheckout = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (cartItemIds: string[]) => {
      const response = await checkoutAPI.checkout({ cart_item_ids: cartItemIds });
      return response;
    },
    onSuccess: (response) => {
      // Invalidate and refetch cart data
      queryClient.invalidateQueries({ queryKey: cartKeys.lists() });
      // Invalidate purchase history
      queryClient.invalidateQueries({ queryKey: checkoutKeys.purchaseHistory() });
      
      if (response.data) {
        // Store checkout data in localStorage for the success page
        localStorage.setItem("checkout_success_data", JSON.stringify(response.data));
        
        // Redirect to success page
        window.location.href = `/checkout/success?order_id=${response.data.order_id}&transaction_id=${response.data.transaction_id}&total_amount=${response.data.total_amount}`;
      }
    },
    onError: (error: any) => {
      console.error("Failed to checkout:", error);
      toast.error(error.response?.data?.error || "Failed to complete checkout");
    },
  });
};

// Hook to get purchase history
export const usePurchaseHistory = (limit?: number, offset?: number) => {
  return useQuery({
    queryKey: checkoutKeys.purchaseHistoryList(`limit=${limit}&offset=${offset}`),
    queryFn: async () => {
      const response = await checkoutAPI.getPurchaseHistory(limit, offset);
      return response.data;
    },
  });
};

// Hook to check if an audiobook is purchased
export const useIsAudioBookPurchased = (audiobookId: string) => {
  return useQuery({
    queryKey: checkoutKeys.isPurchased(audiobookId),
    queryFn: async () => {
      const response = await checkoutAPI.isAudioBookPurchased(audiobookId);
      return response.data?.is_purchased || false;
    },
    enabled: !!audiobookId,
  });
};
