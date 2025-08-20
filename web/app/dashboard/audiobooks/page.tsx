"use client";

import { useUser } from "@/hooks/use-auth";
import { useAudioBooks } from "@/hooks/use-audiobooks";
import {
  Plus,
  Search,
  Filter,
  MoreHorizontal,
  BookOpen,
  Loader2,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { redirect } from "next/navigation";
import { useEffect } from "react";

export default function AudioBooksPage() {
  const { data: user, isLoading: userLoading } = useUser();
  const {
    data: audiobooksResponse,
    isLoading: audiobooksLoading,
    error: audiobooksError,
  } = useAudioBooks();

  const audioBooks = audiobooksResponse?.data || [];

  useEffect(() => {
    if (!userLoading && !user) {
      redirect("/auth/login");
    }

    if (!userLoading && user) {
      const userRole = user.user_metadata?.role || "user";
      if (userRole !== "admin") {
        redirect("/");
      }
    }
  }, [user, userLoading]);

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading...</span>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="font-bold text-3xl mb-2">Audio Books</h1>
          <p className="text-muted-foreground">
            Manage your audio book collection
          </p>
        </div>
        <Button asChild>
          <Link href="/dashboard/audiobooks/create">
            <Plus className="h-4 w-4 mr-2" />
            Upload New Book
          </Link>
        </Button>
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

      {/* Loading State */}
      {audiobooksLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="flex items-center gap-2">
            <Loader2 className="h-6 w-6 animate-spin" />
            <span>Loading audio books...</span>
          </div>
        </div>
      )}

      {/* Error State */}
      {audiobooksError && (
        <Card className="text-center py-12">
          <CardContent>
            <div className="text-red-600 mb-4">
              <h3 className="text-lg font-semibold mb-2">
                Error Loading Audio Books
              </h3>
              <p className="text-sm">
                {audiobooksError instanceof Error
                  ? audiobooksError.message
                  : "An error occurred"}
              </p>
            </div>
            <Button onClick={() => window.location.reload()}>Try Again</Button>
          </CardContent>
        </Card>
      )}

      {/* Audio Books Grid */}
      {!audiobooksLoading && !audiobooksError && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {audioBooks.map((book) => (
            <Card key={book.id} className="hover:shadow-lg transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <CardTitle className="text-lg line-clamp-2">
                      {book.title}
                    </CardTitle>
                    <CardDescription className="line-clamp-1">
                      by {book.author}
                    </CardDescription>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem asChild>
                        <Link href={`/dashboard/audiobooks/${book.id}`}>
                          View Details
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild>
                        <Link href={`/dashboard/audiobooks/${book.id}/edit`}>
                          Edit
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-red-600">
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent>
                {book.cover_image && (
                  <div className="mb-4">
                    <img
                      src={book.cover_image}
                      alt={book.title}
                      className="w-full h-32 object-cover rounded-md"
                    />
                  </div>
                )}

                <div className="flex justify-between items-center">
                  <div className="flex gap-2">
                    {book.duration && (
                      <Badge variant="secondary">
                        {Math.round(book.duration / 60)} min
                      </Badge>
                    )}
                    <Badge variant="outline">
                      {new Date(book.created_at).toLocaleDateString()}
                    </Badge>
                  </div>
                  <Button size="sm" asChild>
                    <Link href={`/dashboard/audiobooks/${book.id}`}>
                      Manage
                    </Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {!audiobooksLoading && !audiobooksError && audioBooks.length === 0 && (
        <Card className="text-center py-12">
          <CardContent>
            <div className="text-muted-foreground mb-4">
              <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <h3 className="text-lg font-semibold mb-2">No Audio Books Yet</h3>
              <p className="mb-4">
                Get started by uploading your first audio book
              </p>
            </div>
            <Button asChild>
              <Link href="/dashboard/audiobooks/create">
                <Plus className="h-4 w-4 mr-2" />
                Upload Your First Book
              </Link>
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
