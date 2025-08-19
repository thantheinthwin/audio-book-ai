import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { publicAPI, libraryAPI } from "@/lib/api";
import { Search, Filter, Play, Heart, Clock, User } from "lucide-react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default async function LibraryPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  // Fetch audio books
  let audioBooks = [];
  let userLibrary = [];

  try {
    const [audioBooksResponse, libraryResponse] = await Promise.allSettled([
      publicAPI.getPublicAudioBooks(),
      libraryAPI.getLibrary(),
    ]);

    if (audioBooksResponse.status === "fulfilled") {
      audioBooks = audioBooksResponse.value.data || [];
    }

    if (libraryResponse.status === "fulfilled") {
      userLibrary = libraryResponse.value.data || [];
    }
  } catch (error) {
    console.error("Failed to fetch audio books:", error);
  }

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const AudioBookCard = ({
    book,
    isInLibrary = false,
  }: {
    book: any;
    isInLibrary?: boolean;
  }) => (
    <Card className="hover:shadow-lg transition-shadow group">
      <CardHeader className="pb-3">
        <div className="relative">
          {book.cover_image && (
            <img
              src={book.cover_image}
              alt={book.title}
              className="w-full h-48 object-cover rounded-md mb-3"
            />
          )}
          <Button
            size="sm"
            className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
            variant="secondary"
          >
            <Heart className="h-4 w-4" />
          </Button>
          <Button
            size="sm"
            className="absolute bottom-2 left-2 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Play className="h-4 w-4" />
          </Button>
        </div>
        <CardTitle className="text-lg line-clamp-2">{book.title}</CardTitle>
        <CardDescription className="flex items-center gap-1">
          <User className="h-3 w-3" />
          {book.author}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {book.description && (
          <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
            {book.description}
          </p>
        )}
        <div className="flex justify-between items-center">
          <div className="flex gap-2">
            {book.duration && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDuration(book.duration)}
              </Badge>
            )}
            {isInLibrary && <Badge variant="default">In Library</Badge>}
          </div>
          <Button size="sm" variant="outline">
            {isInLibrary ? "Remove" : "Add to Library"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-bold text-3xl mb-2">Audio Book Library</h1>
        <p className="text-muted-foreground">
          Discover and listen to amazing audio books
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

      {/* Tabs */}
      <Tabs defaultValue="all" className="w-full">
        <TabsList>
          <TabsTrigger value="all">All Books ({audioBooks.length})</TabsTrigger>
          <TabsTrigger value="library">
            My Library ({userLibrary.length})
          </TabsTrigger>
          <TabsTrigger value="recent">Recently Played</TabsTrigger>
          <TabsTrigger value="favorites">Favorites</TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {audioBooks.map((book) => (
              <AudioBookCard
                key={book.id}
                book={book}
                isInLibrary={userLibrary.some(
                  (libBook: any) => libBook.id === book.id
                )}
              />
            ))}
          </div>
        </TabsContent>

        <TabsContent value="library" className="space-y-6">
          {userLibrary.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {userLibrary.map((book) => (
                <AudioBookCard key={book.id} book={book} isInLibrary={true} />
              ))}
            </div>
          ) : (
            <Card className="text-center py-12">
              <CardContent>
                <div className="text-muted-foreground mb-4">
                  <Heart className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-semibold mb-2">
                    Your Library is Empty
                  </h3>
                  <p className="mb-4">
                    Start building your audio book collection by adding books
                    from the catalog
                  </p>
                </div>
                <Button>Browse All Books</Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="recent" className="space-y-6">
          <Card className="text-center py-12">
            <CardContent>
              <div className="text-muted-foreground mb-4">
                <Clock className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">
                  No Recent Activity
                </h3>
                <p className="mb-4">
                  Start listening to audio books to see your recent activity
                  here
                </p>
              </div>
              <Button>Start Listening</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="favorites" className="space-y-6">
          <Card className="text-center py-12">
            <CardContent>
              <div className="text-muted-foreground mb-4">
                <Heart className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">No Favorites Yet</h3>
                <p className="mb-4">
                  Like audio books to add them to your favorites
                </p>
              </div>
              <Button>Discover Books</Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
