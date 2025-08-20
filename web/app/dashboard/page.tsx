"use client";

import { useUser, useSession } from "@/hooks/use-auth";
import { useAudioBooks } from "@/hooks/use-audiobooks";
import { InfoIcon, Users, BookOpen, BarChart3, Loader2 } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { redirect } from "next/navigation";
import { useEffect } from "react";

export default function DashboardPage() {
  const { data: user, isLoading: userLoading, error: userError } = useUser();
  const { data: session, isLoading: sessionLoading } = useSession();
  const { data: audiobooks, isLoading: audiobooksLoading } = useAudioBooks();

  const isLoading = userLoading || sessionLoading;

  useEffect(() => {
    if (!isLoading && !user) {
      redirect("/auth/login");
    }

    if (!isLoading && user) {
      const userRole = user.user_metadata?.role || "user";
      if (userRole !== "admin") {
        redirect("/");
      }
    }
  }, [user, isLoading]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading dashboard...</span>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect in useEffect
  }

  const userRole = user.user_metadata?.role || "user";
  const totalAudiobooks = audiobooks?.data?.length || 0;

  return (
    <div className="space-y-6">
      <div className="w-full">
        <div className="bg-blue-50 dark:bg-blue-950 text-sm p-3 px-5 rounded-md text-foreground flex gap-3 items-center">
          <InfoIcon size="16" strokeWidth={2} />
          Welcome to the Admin Dashboard. You have full administrative
          privileges.
        </div>
      </div>

      <div className="flex flex-col gap-6">
        <div>
          <h1 className="font-bold text-3xl mb-2">Admin Dashboard</h1>
          <p className="text-muted-foreground">
            Manage your audio book platform
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                User Management
              </CardTitle>
              <CardDescription>
                Manage user accounts, roles, and permissions
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Button className="w-full" asChild>
                  <Link href="/dashboard/users">View All Users</Link>
                </Button>
                <Button variant="outline" className="w-full" asChild>
                  <Link href="/dashboard/users/create">Create User</Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BookOpen className="h-5 w-5" />
                Audio Book Management
              </CardTitle>
              <CardDescription>
                Upload, edit, and manage audio book content
                {!audiobooksLoading && (
                  <span className="block text-sm font-medium">
                    Total Books: {totalAudiobooks}
                  </span>
                )}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Button className="w-full" asChild>
                  <Link href="/dashboard/audiobooks">View All Books</Link>
                </Button>
                <Button variant="outline" className="w-full" asChild>
                  <Link href="/dashboard/audiobooks/create">
                    Upload New Book
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BarChart3 className="h-5 w-5" />
                Analytics
              </CardTitle>
              <CardDescription>
                View platform statistics and user insights
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Button className="w-full" asChild>
                  <Link href="/dashboard/analytics">View Analytics</Link>
                </Button>
                <Button variant="outline" className="w-full" asChild>
                  <Link href="/dashboard/reports">Generate Reports</Link>
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="flex flex-col gap-2 items-start">
          <h2 className="font-bold text-2xl mb-4">Admin User Details</h2>
          <pre className="text-xs font-mono p-3 rounded border max-h-32 overflow-auto">
            {JSON.stringify(user, null, 2)}
          </pre>
        </div>
      </div>
    </div>
  );
}
