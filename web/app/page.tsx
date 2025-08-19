import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { publicAPI, libraryAPI } from "@/lib/api";
import { BookOpen, Headphones, Clock, Heart, TrendingUp } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default async function HomePage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  // Check user role
  const userRole = data.claims.user_metadata.role || "user";

  // If admin, redirect to admin dashboard
  if (userRole === "admin") {
    redirect("/dashboard");
  }

  // Fetch user's audio book data
  let audioBooks: any[] = [];
  let userLibrary: any[] = [];
  let recentBooks: any[] = [];

  try {
    const [audioBooksResponse, libraryResponse] = await Promise.allSettled([
      publicAPI.getPublicAudioBooks(),
      libraryAPI.getLibrary(),
    ]);

    if (audioBooksResponse.status === "fulfilled") {
      audioBooks = audioBooksResponse.value.data || [];
      // Get recent books (last 4)
      recentBooks = audioBooks.slice(0, 4);
    }

    if (libraryResponse.status === "fulfilled") {
      userLibrary = (libraryResponse.value as any).data || [];
    }
  } catch (error) {
    console.error("Failed to fetch audio books:", error);
  }

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  return (
    <div className="space-y-6">
      {/* Welcome Header */}
      <div className="text-center space-y-2">
        <h1 className="text-4xl font-bold">Welcome to Audio Book AI</h1>
        <p className="text-xl text-muted-foreground">
          Discover, listen, and enjoy amazing audio books
        </p>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Books</CardTitle>
            <BookOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{audioBooks.length}</div>
            <p className="text-xs text-muted-foreground">
              Available audio books
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">My Library</CardTitle>
            <Headphones className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{userLibrary.length}</div>
            <p className="text-xs text-muted-foreground">
              Books in your library
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Listening Time
            </CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0h</div>
            <p className="text-xs text-muted-foreground">
              Total listening time
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Favorites</CardTitle>
            <Heart className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">Liked books</p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="h-5 w-5" />
              Browse Library
            </CardTitle>
            <CardDescription>
              Discover new audio books and add them to your collection
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild className="w-full">
              <Link href="/library">Explore Audio Books</Link>
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Continue Listening
            </CardTitle>
            <CardDescription>
              Pick up where you left off with your recent audio books
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild className="w-full" variant="outline">
              <Link href="/library">View Recent</Link>
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Recent Books */}
      {recentBooks.length > 0 && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-2xl font-bold">Recently Added</h2>
            <Button variant="outline" asChild>
              <Link href="/library">View All</Link>
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {recentBooks.map((book) => (
              <Card key={book.id} className="hover:shadow-lg transition-shadow">
                <CardContent className="p-4">
                  {book.cover_image && (
                    <img
                      src={book.cover_image}
                      alt={book.title}
                      className="w-full h-32 object-cover rounded-md mb-3"
                    />
                  )}
                  <h3 className="font-semibold text-sm line-clamp-2 mb-2">
                    {book.title}
                  </h3>
                  <p className="text-xs text-muted-foreground mb-2">
                    by {book.author}
                  </p>
                  <div className="flex justify-between items-center">
                    {book.duration && (
                      <Badge variant="secondary" className="text-xs">
                        {formatDuration(book.duration)}
                      </Badge>
                    )}
                    <Button size="sm" asChild>
                      <Link href={`/library/${book.id}`}>Listen</Link>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Features */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Headphones className="h-5 w-5" />
              High Quality Audio
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Enjoy crystal clear audio with our high-quality streaming
              technology.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="h-5 w-5" />
              AI-Powered Features
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Get transcripts, summaries, and intelligent recommendations
              powered by AI.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Track Progress
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Never lose your place with automatic progress tracking and
              bookmarks.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
