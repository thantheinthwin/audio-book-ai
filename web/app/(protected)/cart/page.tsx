"use client";

import { useCart, useRemoveFromCart, useCheckout } from "@/hooks/use-cart";
import { useUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Trash2, ShoppingCart, DollarSign, CreditCard } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { redirect, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

export default function CartPage() {
  const { data: user, isLoading: userLoading } = useUser();
  const { data: cart, isLoading, error } = useCart();
  const removeFromCartMutation = useRemoveFromCart();
  const checkoutMutation = useCheckout();
  const router = useRouter();
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [itemToRemove, setItemToRemove] = useState<{
    id: string;
    title: string;
  } | null>(null);

  // Check if user is a normal user (not admin)
  useEffect(() => {
    if (!userLoading && !user) {
      redirect("/auth/login");
    }

    if (!userLoading && user) {
      const userRole = user.user_metadata?.role || "user";
      if (userRole !== "user") {
        redirect("/");
      }
    }
  }, [user, userLoading]);

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <ShoppingCart className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  const handleCheckout = () => {
    if (selectedItems.length === 0) {
      // If no items selected, checkout all items
      const allItemIds = cart?.items.map((item) => item.id) || [];
      checkoutMutation.mutate(allItemIds);
    } else {
      checkoutMutation.mutate(selectedItems);
    }
  };

  const handleItemSelect = (itemId: string) => {
    setSelectedItems((prev) =>
      prev.includes(itemId)
        ? prev.filter((id) => id !== itemId)
        : [...prev, itemId]
    );
  };

  const isAllSelected =
    cart?.items &&
    cart.items.length > 0 &&
    selectedItems.length === cart.items.length;
  const isSomeSelected = selectedItems.length > 0;

  const handleSelectAll = () => {
    if (isAllSelected) {
      setSelectedItems([]);
    } else {
      const allItemIds = cart?.items?.map((item) => item.id) || [];
      setSelectedItems(allItemIds);
    }
  };

  // Event delegation for remove buttons
  const handleRemoveClick = (e: React.MouseEvent) => {
    const target = e.target as HTMLElement;
    const removeButton = target.closest("[data-remove-item]") as HTMLElement;

    if (removeButton) {
      e.preventDefault();
      e.stopPropagation();

      const itemId = removeButton.dataset.removeItem;
      const itemTitle = removeButton.dataset.itemTitle;

      if (itemId && itemTitle) {
        setItemToRemove({ id: itemId, title: itemTitle });
        setShowConfirmDialog(true);
      }
    }
  };

  const handleConfirmRemove = () => {
    if (itemToRemove) {
      removeFromCartMutation.mutate(itemToRemove.id);
      setShowConfirmDialog(false);
      setItemToRemove(null);
    }
  };

  const handleCancelRemove = () => {
    setShowConfirmDialog(false);
    setItemToRemove(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <ShoppingCart className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Loading cart...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <ShoppingCart className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Failed to load cart</p>
        </div>
      </div>
    );
  }

  if (!cart || cart.total_items === 0) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Shopping Cart</h1>
          <p className="text-muted-foreground">Your cart is empty</p>
        </div>

        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <ShoppingCart className="h-16 w-16 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold mb-2">Your cart is empty</h3>
            <p className="text-muted-foreground mb-4">
              Add some audiobooks to get started
            </p>
            <Link href="/library">
              <Button>Browse Audiobooks</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Shopping Cart</h1>
        <p className="text-muted-foreground">
          {cart.total_items} item{cart.total_items !== 1 ? "s" : ""} in your
          cart
        </p>
      </div>

      <div className="grid gap-6">
        {/* Cart Items */}
        <div className="space-y-4" onClick={handleRemoveClick}>
          {cart.items.map((item) => (
            <Card key={item.id}>
              <CardContent className="p-6">
                <div className="flex gap-4">
                  <div className="flex items-start pt-1">
                    <input
                      type="checkbox"
                      checked={selectedItems.includes(item.id)}
                      onChange={() => handleItemSelect(item.id)}
                      className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                  </div>
                  <div className="flex-shrink-0">
                    <Image
                      src={
                        item.audiobook.cover_image_url ||
                        "/images/placeholder.png"
                      }
                      alt={item.audiobook.title}
                      width={80}
                      height={80}
                      className="w-20 h-20 object-cover rounded-md"
                    />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-start">
                      <div className="flex-1 min-w-0">
                        <h3 className="font-semibold text-lg truncate">
                          {item.audiobook.title}
                        </h3>
                        <p className="text-muted-foreground">
                          by {item.audiobook.author}
                        </p>
                        <div className="flex items-center gap-2 mt-2">
                          <Badge variant="secondary">
                            {item.audiobook.language.toUpperCase()}
                          </Badge>
                          {item.audiobook.duration_seconds && (
                            <Badge variant="outline">
                              {Math.floor(item.audiobook.duration_seconds / 60)}
                              m
                            </Badge>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <div className="flex items-center gap-1">
                            <DollarSign className="h-4 w-4" />
                            <span className="font-semibold text-lg">
                              {item.audiobook.price.toFixed(2)}
                            </span>
                          </div>
                        </div>
                        <Button
                          variant="outline"
                          size="icon"
                          disabled={removeFromCartMutation.isPending}
                          data-remove-item={item.audiobook_id}
                          data-item-title={item.audiobook.title}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Cart Summary */}
        <Card>
          <CardHeader>
            <CardTitle>Order Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={isAllSelected}
                onChange={handleSelectAll}
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label className="text-sm font-medium">
                Select All ({cart.total_items} items)
              </label>
            </div>
            <div className="flex justify-between">
              <span>Subtotal ({cart.total_items} items)</span>
              <span className="font-semibold">
                ${cart.total_price.toFixed(2)}
              </span>
            </div>
            <div className="border-t pt-4">
              <div className="flex justify-between text-lg font-semibold">
                <span>Total</span>
                <span>${cart.total_price.toFixed(2)}</span>
              </div>
            </div>
            <Button
              className="w-full"
              size="lg"
              onClick={handleCheckout}
              disabled={checkoutMutation.isPending || cart.total_items === 0}
            >
              {checkoutMutation.isPending ? (
                "Processing..."
              ) : (
                <>
                  <CreditCard className="h-4 w-4 mr-2" />
                  {isSomeSelected
                    ? `Checkout Selected (${selectedItems.length})`
                    : "Checkout All Items"}
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Confirmation Dialog */}
      <Dialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Removal</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove &quot;{itemToRemove?.title}&quot;?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={handleCancelRemove}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleConfirmRemove}>
              Remove
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
