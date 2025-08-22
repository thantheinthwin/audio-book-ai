"use client";

import { useParams } from "next/navigation";
import { useAudioBook } from "@/hooks/use-audiobooks";
import { useAddToCart, useIsInCart, useRemoveFromCart } from "@/hooks/use-cart";
import { notFound } from "next/navigation";
import { Trash2, Loader2, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import Image from "next/image";

export default function AudioBookDetailPage() {
  const params = useParams();
  const audiobookId = params.id as string;

  const {
    data: audioBookResponse,
    error: audioBookError,
    isLoading: audioBookLoading,
  } = useAudioBook(audiobookId);

  const audioBook = audioBookResponse?.data;

  // Cart functionality
  const { data: isInCart, isLoading: isInCartLoading } =
    useIsInCart(audiobookId);
  const addToCartMutation = useAddToCart();
  const removeFromCartMutation = useRemoveFromCart();

  const handleAddToCart = () => {
    if (!audioBook) return;

    addToCartMutation.mutate({
      audiobook_id: audioBook.id,
    });
  };

  const handleRemoveFromCart = () => {
    if (!audioBook) return;

    removeFromCartMutation.mutate(audioBook.id);
  };

  if (audioBookLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (audioBookError || !audioBook) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="p-6 flex gap-4">
          <div className="flex flex-col gap-4">
            <Image
              src={"/images/placeholder.png"}
              alt={audioBook.title}
              width={100}
              height={100}
              className="w-48 h-48 object-cover rounded-md"
            />
            <div className="flex gap-2">
              {/* <Button variant="outline">
                <Edit className="h-4 w-4" />
                Edit
              </Button> */}
              {isInCart ? (
                <Button
                  variant="destructive"
                  className="w-full"
                  onClick={handleRemoveFromCart}
                  disabled={removeFromCartMutation.isPending || isInCartLoading}
                >
                  {removeFromCartMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                  {removeFromCartMutation.isPending
                    ? "Removing..."
                    : "Remove from cart"}
                </Button>
              ) : (
                <Button
                  variant="secondary"
                  className="w-full"
                  onClick={handleAddToCart}
                  disabled={addToCartMutation.isPending || isInCartLoading}
                >
                  {addToCartMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <ShoppingCart className="h-4 w-4" />
                  )}
                  {addToCartMutation.isPending ? "Adding..." : "Add to cart"}
                </Button>
              )}
            </div>
          </div>
          <div className="flex flex-col flex-1 gap-2">
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Title</h2>
              <p className="font-semibold text-lg">{audioBook.title}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Author</h2>
              <p className="font-semibold text-lg">{audioBook.author}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Summary</h2>
              <p className=" text-sm">{audioBook.summary}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Tags</h2>
              <p className="text-xs">{audioBook.tags?.join(", ")}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
