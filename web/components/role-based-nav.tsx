"use client";

import { createClient } from "@/lib/supabase/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useEffect, useState } from "react";
import Link from "next/link";

interface User {
  id: string;
  email: string;
  role: string;
}

export function RoleBasedNav() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const supabase = createClient();

  useEffect(() => {
    const getUser = async () => {
      const {
        data: { user },
      } = await supabase.auth.getUser();
      if (user) {
        setUser({
          id: user.id,
          email: user.email || "",
          role: (user.user_metadata?.role as string) || "user",
        });
      }
      setLoading(false);
    };

    getUser();

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange(async (event, session) => {
      if (session?.user) {
        setUser({
          id: session.user.id,
          email: session.user.email || "",
          role: (session.user.user_metadata?.role as string) || "user",
        });
      } else {
        setUser(null);
      }
      setLoading(false);
    });

    return () => subscription.unsubscribe();
  }, [supabase.auth]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Loading...</CardTitle>
        </CardHeader>
      </Card>
    );
  }

  if (!user) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Welcome to Audio Book AI</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-4">
            Please sign in to access your account.
          </p>
          <div className="flex gap-2">
            <Button asChild>
              <Link href="/auth/login">Login</Link>
            </Button>
            <Button variant="outline" asChild>
              <Link href="/auth/sign-up">Sign Up</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Welcome, {user.email}
          <Badge variant={user.role === "admin" ? "default" : "secondary"}>
            {user.role}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <h3 className="font-medium mb-2">User Features</h3>
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" size="sm" asChild>
                <Link href="/">Dashboard</Link>
              </Button>
              <Button variant="outline" size="sm" asChild>
                <Link href="/library">My Library</Link>
              </Button>
              <Button variant="outline" size="sm" asChild>
                <Link href="/playlists">Playlists</Link>
              </Button>
              <Button variant="outline" size="sm" asChild>
                <Link href="/progress">Progress</Link>
              </Button>
            </div>
          </div>

          {user.role === "admin" && (
            <div>
              <h3 className="font-medium mb-2">Admin Features</h3>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" size="sm" asChild>
                  <Link href="/dashboard">Admin Dashboard</Link>
                </Button>
                <Button variant="outline" size="sm" asChild>
                  <Link href="/dashboard/users">Manage Users</Link>
                </Button>
                <Button variant="outline" size="sm" asChild>
                  <Link href="/dashboard/audiobooks">Manage Audio Books</Link>
                </Button>
                <Button variant="outline" size="sm" asChild>
                  <Link href="/dashboard/analytics">Analytics</Link>
                </Button>
              </div>
            </div>
          )}

          <div className="pt-4 border-t">
            <Button variant="outline" size="sm" asChild>
              <Link href="/auth/update-password">Update Password</Link>
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
