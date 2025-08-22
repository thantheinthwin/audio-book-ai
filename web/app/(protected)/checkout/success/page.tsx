"use client";

import { useUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle, Download, Home, Library, Receipt } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { redirect, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { motion } from "framer-motion";

interface CheckoutSuccessData {
  order_id: string;
  purchased_items: Array<{
    id: string;
    user_id: string;
    audiobook_id: string;
    purchase_price: number;
    purchased_at: string;
    transaction_id?: string;
    payment_status: string;
    audiobook: {
      id: string;
      title: string;
      author: string;
      description: string;
      price: number;
      cover_image_url?: string;
      language: string;
      duration_seconds?: number;
    };
  }>;
  total_amount: number;
  transaction_id: string;
  checkout_completed_at: string;
}

export default function CheckoutSuccessPage() {
  const { data: user, isLoading: userLoading } = useUser();
  const searchParams = useSearchParams();
  const [checkoutData, setCheckoutData] = useState<CheckoutSuccessData | null>(
    null
  );
  const [isLoading, setIsLoading] = useState(true);

  // Check if user is authenticated
  useEffect(() => {
    if (!userLoading && !user) {
      redirect("/auth/login");
    }
  }, [user, userLoading]);

  // Load checkout data from URL params or localStorage
  useEffect(() => {
    if (user) {
      // Try to get data from URL params first
      const orderId = searchParams.get("order_id");
      const transactionId = searchParams.get("transaction_id");

      if (orderId && transactionId) {
        // If we have URL params, construct the data
        const totalAmount = parseFloat(searchParams.get("total_amount") || "0");
        const checkoutData: CheckoutSuccessData = {
          order_id: orderId,
          transaction_id: transactionId,
          total_amount: totalAmount,
          checkout_completed_at: new Date().toISOString(),
          purchased_items: [], // This would need to be passed via URL or localStorage
        };
        setCheckoutData(checkoutData);
      } else {
        // Try to get from localStorage (fallback)
        const storedData = localStorage.getItem("checkout_success_data");
        if (storedData) {
          try {
            setCheckoutData(JSON.parse(storedData));
            // Clear the stored data after reading
            localStorage.removeItem("checkout_success_data");
          } catch (error) {
            console.error("Failed to parse checkout data:", error);
          }
        }
      }
      setIsLoading(false);
    }
  }, [user, searchParams]);

  if (userLoading || isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <motion.div
          className="text-center"
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.5 }}
        >
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
          >
            <CheckCircle className="h-12 w-12 mx-auto text-green-500 mb-4" />
          </motion.div>
          <p className="text-muted-foreground">Loading...</p>
        </motion.div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  if (!checkoutData) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <motion.div
          className="text-center"
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.6 }}
        >
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.3, type: "spring", stiffness: 200 }}
          >
            <CheckCircle className="h-12 w-12 mx-auto text-green-500 mb-4" />
          </motion.div>
          <motion.h2
            className="text-2xl font-bold mb-2"
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.5, duration: 0.4 }}
          >
            Checkout Successful!
          </motion.h2>
          <motion.p
            className="text-muted-foreground mb-4"
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.7, duration: 0.4 }}
          >
            Your order has been processed successfully.
          </motion.p>
          <motion.div
            className="space-x-4"
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.9, duration: 0.4 }}
          >
            <Link href="/library">
              <Button>
                <Library className="h-4 w-4 mr-2" />
                Go to Library
              </Button>
            </Link>
            <Link href="/">
              <Button variant="outline">
                <Home className="h-4 w-4 mr-2" />
                Go Home
              </Button>
            </Link>
          </motion.div>
        </motion.div>
      </div>
    );
  }

  return (
    <motion.div
      className="space-y-6"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      {/* Success Header */}
      <motion.div
        className="text-center space-y-4"
        initial={{ scale: 0.9, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.2 }}
      >
        <motion.div
          className="flex justify-center"
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.4, type: "spring", stiffness: 200 }}
        >
          <div className="bg-green-100 p-3 rounded-full">
            <CheckCircle className="h-8 w-8 text-green-600" />
          </div>
        </motion.div>
        <motion.h1
          className="text-3xl font-bold"
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.6, duration: 0.4 }}
        >
          Checkout Successful!
        </motion.h1>
        <motion.p
          className="text-muted-foreground"
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.8, duration: 0.4 }}
        >
          Thank you for your purchase. Your audiobooks are now available in your
          library.
        </motion.p>
      </motion.div>

      {/* Order Summary */}
      <motion.div
        initial={{ scale: 0.9, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 1.0 }}
      >
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Receipt className="h-5 w-5" />
              Order Summary
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium">Order ID:</span>
                <p className="text-muted-foreground">{checkoutData.order_id}</p>
              </div>
              <div>
                <span className="font-medium">Transaction ID:</span>
                <p className="text-muted-foreground">
                  {checkoutData.transaction_id}
                </p>
              </div>
              <div>
                <span className="font-medium">Date:</span>
                <p className="text-muted-foreground">
                  {new Date(
                    checkoutData.checkout_completed_at
                  ).toLocaleDateString()}
                </p>
              </div>
              <div>
                <span className="font-medium">Total Amount:</span>
                <p className="font-semibold text-lg">
                  ${checkoutData.total_amount.toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* Purchased Items */}
      {checkoutData.purchased_items.length > 0 && (
        <motion.div
          initial={{ scale: 0.9, opacity: 0, y: 20 }}
          animate={{ scale: 1, opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 1.2 }}
        >
          <Card>
            <CardHeader>
              <CardTitle>Purchased Items</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {checkoutData.purchased_items.map((item, index) => (
                  <motion.div
                    key={item.id}
                    className="flex gap-4 p-4 border rounded-lg"
                    initial={{ scale: 0.9, opacity: 0, x: -20 }}
                    animate={{ scale: 1, opacity: 1, x: 0 }}
                    transition={{ duration: 0.5, delay: 1.4 + index * 0.1 }}
                  >
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
                      <h3 className="font-semibold text-lg">
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
                            {Math.floor(item.audiobook.duration_seconds / 60)}m
                          </Badge>
                        )}
                        <Badge
                          variant="default"
                          className="bg-green-100 text-green-800"
                        >
                          Purchased
                        </Badge>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">
                        ${item.purchase_price.toFixed(2)}
                      </span>
                      <Button size="sm" variant="outline">
                        <Download className="h-4 w-4 mr-2" />
                        Download
                      </Button>
                    </div>
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}

      {/* Action Buttons */}
      <motion.div
        className="flex justify-center gap-4"
        initial={{ scale: 0.9, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 1.6 }}
      >
        <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
          <Link href="/library">
            <Button size="lg">
              <Library className="h-4 w-4 mr-2" />
              Go to Library
            </Button>
          </Link>
        </motion.div>
        <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
          <Link href="/purchases">
            <Button size="lg" variant="outline">
              <Receipt className="h-4 w-4 mr-2" />
              View Purchase History
            </Button>
          </Link>
        </motion.div>
        <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
          <Link href="/">
            <Button size="lg" variant="outline">
              <Home className="h-4 w-4 mr-2" />
              Go Home
            </Button>
          </Link>
        </motion.div>
      </motion.div>
    </motion.div>
  );
}
