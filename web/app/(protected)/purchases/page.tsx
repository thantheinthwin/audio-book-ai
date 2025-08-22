"use client";

import { usePurchaseHistory } from "@/hooks/use-cart";
import { useUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShoppingBag, DollarSign, Calendar, Package } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { redirect } from "next/navigation";
import { useEffect } from "react";

export default function PurchasesPage() {
  const { data: user, isLoading: userLoading } = useUser();
  const { data: purchaseHistory, isLoading, error } = usePurchaseHistory(20, 0);

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
          <ShoppingBag className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <ShoppingBag className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Loading purchase history...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <ShoppingBag className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Failed to load purchase history</p>
        </div>
      </div>
    );
  }

  if (!purchaseHistory || purchaseHistory.total_items === 0) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Purchase History</h1>
          <p className="text-muted-foreground">Your purchase history is empty</p>
        </div>

        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <ShoppingBag className="h-16 w-16 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold mb-2">No purchases yet</h3>
            <p className="text-muted-foreground mb-4">
              Start shopping to see your purchase history here
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
        <h1 className="text-3xl font-bold">Purchase History</h1>
        <p className="text-muted-foreground">
          {purchaseHistory.total_items} purchase{purchaseHistory.total_items !== 1 ? "s" : ""} • Total spent: ${purchaseHistory.total_spent.toFixed(2)}
        </p>
      </div>

      <div className="grid gap-6">
        {/* Purchase Items */}
        <div className="space-y-4">
          {purchaseHistory.purchases.map((purchase) => (
            <Card key={purchase.id}>
              <CardContent className="p-6">
                <div className="flex gap-4">
                  <div className="flex-shrink-0">
                    <Image
                      src={
                        purchase.audiobook.cover_image_url ||
                        "/images/placeholder.png"
                      }
                      alt={purchase.audiobook.title}
                      width={80}
                      height={80}
                      className="w-20 h-20 object-cover rounded-md"
                    />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-start">
                      <div className="flex-1 min-w-0">
                        <h3 className="font-semibold text-lg truncate">
                          {purchase.audiobook.title}
                        </h3>
                        <p className="text-muted-foreground">
                          by {purchase.audiobook.author}
                        </p>
                        <div className="flex items-center gap-2 mt-2">
                          <Badge variant="secondary">
                            {purchase.audiobook.language.toUpperCase()}
                          </Badge>
                          {purchase.audiobook.duration_seconds && (
                            <Badge variant="outline">
                              {Math.floor(purchase.audiobook.duration_seconds / 60)}
                              m
                            </Badge>
                          )}
                          <Badge variant="default" className="bg-green-100 text-green-800">
                            <Package className="h-3 w-3 mr-1" />
                            Purchased
                          </Badge>
                        </div>
                        <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                          <div className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            <span>
                              {new Date(purchase.purchased_at).toLocaleDateString()}
                            </span>
                          </div>
                          {purchase.transaction_id && (
                            <span>Transaction: {purchase.transaction_id}</span>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <div className="flex items-center gap-1">
                            <DollarSign className="h-4 w-4" />
                            <span className="font-semibold text-lg">
                              {purchase.purchase_price.toFixed(2)}
                            </span>
                          </div>
                        </div>
                        <Link href={`/audiobooks/${purchase.audiobook_id}`}>
                          <Button variant="outline" size="sm">
                            Listen
                          </Button>
                        </Link>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
