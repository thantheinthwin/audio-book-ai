"use client";

import { Search, Filter, Heart, Clock, User, ExternalLink, ShoppingCart, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import Link from "next/link";
import Image from "next/image";
import { Loader2 } from "lucide-react";
import { useAudioBooks } from "@/hooks";
import { useUser } from "@/hooks/use-auth";
import { useAddToCart, useIsInCart, useIsAudioBookPurchased } from "@/hooks/use-cart";
import { redirect } from "next/navigation";
import { useEffect } from "react";

export default function LibraryPage() {
  const { data: user, isLoading: userLoading } = useUser();

  // Fetch library data using React Query
  // const { data: libraryResponse, isLoading, error } = useLibrary();
  const { data: libraryResponse, isLoading, error } = useAudioBooks();

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
          <Loader2 className="h-6 w-6 animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  const userLibrary = (libraryResponse as any)?.data || [];

  const formatDuration = (seconds: number) => {
    if (!seconds) return "N/A";
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatFileSize = (bytes: number) => {
    if (!bytes) return "N/A";
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(1)} MB`;
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "completed":
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            Completed
          </Badge>
        );
      case "processing":
        return (
          <Badge variant="secondary" className="bg-blue-100 text-blue-800">
            Processing
          </Badge>
        );
      case "failed":
        return <Badge variant="destructive">Failed</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const AudioBooksTable = ({ books }: { books: any[] }) => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-16">Cover</TableHead>
          <TableHead>Title</TableHead>
          <TableHead>Author</TableHead>
          <TableHead>Duration</TableHead>
          <TableHead>Price</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {books.map((book) => (
          <AudioBookRow key={book.id} book={book} />
        ))}
      </TableBody>
    </Table>
  );

  const AudioBookRow = ({ book }: { book: any }) => {
    const addToCartMutation = useAddToCart();
    const { data: isInCart } = useIsInCart(book.id);
    const { data: isPurchased } = useIsAudioBookPurchased(book.id);

    const handleAddToCart = () => {
      addToCartMutation.mutate({ audiobook_id: book.id });
    };

    return (
      <TableRow className="hover:bg-muted/50">
        <TableCell>
          <div className="w-12 h-12 rounded-md overflow-hidden">
            {book.cover_image_url || book.cover_image ? (
              <Image
                src={book.cover_image_url || book.cover_image}
                alt={book.title}
                width={48}
                height={48}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full bg-muted flex items-center justify-center">
                <span className="text-xs text-muted-foreground">
                  No Image
                </span>
              </div>
            )}
          </div>
        </TableCell>
        <TableCell>
          <div className="font-medium">{book.title}</div>
        </TableCell>
        <TableCell>
          <div className="flex items-center gap-1">
            <User className="h-3 w-3 text-muted-foreground" />
            {book.author}
          </div>
        </TableCell>
        <TableCell>
          <div className="flex items-center gap-1">
            <Clock className="h-3 w-3 text-muted-foreground" />
            {formatDuration(book.duration_seconds || book.duration)}
          </div>
        </TableCell>
        <TableCell>
          <div className="font-medium">${book.price?.toFixed(2) || "0.00"}</div>
        </TableCell>
        <TableCell>
          <div className="flex flex-col gap-1">
            {getStatusBadge(book.status)}
            {isPurchased && (
              <Badge variant="default" className="bg-green-100 text-green-800">
                <Package className="h-3 w-3 mr-1" />
                Purchased
              </Badge>
            )}
          </div>
        </TableCell>
        <TableCell>
          <div className="flex gap-2">
            {isPurchased ? (
              <Link href={`/audiobooks/${book.id}`}>
                <Button variant="outline" size="sm">
                  Listen
                </Button>
              </Link>
            ) : (
              <>
                <Link href={`/library/${book.id}`}>
                  <Button variant="outline" size="sm">
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </Link>
                {!isInCart && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleAddToCart}
                    disabled={addToCartMutation.isPending}
                  >
                    <ShoppingCart className="h-4 w-4" />
                  </Button>
                )}
                {isInCart && (
                  <Badge variant="secondary" className="bg-blue-100 text-blue-800">
                    In Cart
                  </Badge>
                )}
              </>
            )}
          </div>
        </TableCell>
      </TableRow>
    );
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Loading your library...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">Error Loading Library</h2>
          <p className="text-muted-foreground mb-4">
            There was an error loading your library. Please try again.
          </p>
          <Button onClick={() => window.location.reload()}>Retry</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-bold text-3xl mb-2">Audio Book Library</h1>
        <p className="text-muted-foreground">
          Discover and manage your audio book collection
        </p>
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input placeholder="Search audio books..." className="pl-10" />
        </div>
        <Button variant="outline">
          <Filter className="h-4 w-4 mr-2" />
          Filter
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>My Library</CardTitle>
          <CardDescription>
            Audio books you&apos;ve added to your personal library
          </CardDescription>
        </CardHeader>
        <CardContent>
          {userLibrary.length > 0 ? (
            <AudioBooksTable books={userLibrary} />
          ) : (
            <div className="text-center py-12">
              <div className="text-muted-foreground mb-4">
                <Heart className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">
                  Your Library is Empty
                </h3>
                <p className="mb-4">
                  Start building your audio book collection by adding books from
                  the catalog
                </p>
              </div>
              <Link href="/audiobooks/create">
                <Button>
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Browse All Books
                </Button>
              </Link>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
