"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "./ui/button";
import { createClient } from "@/lib/supabase/client";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { User } from "@supabase/supabase-js";

export function UserProfile() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const supabase = createClient();
  const router = useRouter();

  useEffect(() => {
    const getUser = async () => {
      const {
        data: { user },
      } = await supabase.auth.getUser();
      setUser(user);
      setLoading(false);
    };

    getUser();

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((event, session) => {
      setUser(session?.user ?? null);
      setLoading(false);
    });

    return () => subscription.unsubscribe();
  }, [supabase.auth]);

  const handleLogout = async () => {
    await supabase.auth.signOut();
    router.push("/auth/login");
  };

  if (loading) {
    return (
      <div className="flex items-center gap-3 p-2 rounded-md bg-accent/50">
        <div className="w-8 h-8 bg-muted animate-pulse rounded-full" />
        <div className="flex-1 min-w-0">
          <div className="h-4 bg-muted animate-pulse rounded" />
          <div className="h-3 bg-muted animate-pulse rounded mt-1" />
        </div>
      </div>
    );
  }

  if (user) {
    return (
      <div className="flex items-center gap-3 rounded-md">
        {/* <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-pink-500 rounded-full flex items-center justify-center">
          <span className="text-white text-xs font-semibold">
            {user.email?.charAt(0).toUpperCase() || "U"}
          </span>
        </div> */}
        <div className="flex-1 min-w-0">
          <div className="text-xs truncate">{user.email || "No email"}</div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className=" hover:bg-accent"
          onClick={handleLogout}
          title="Logout"
        >
          <LogOut className="h-3 w-3" />
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3 p-2 rounded-md bg-accent/50">
      <div className="w-8 h-8 bg-muted rounded-full" />
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium truncate">Not signed in</div>
        <div className="text-xs text-muted-foreground truncate">
          Please sign in
        </div>
      </div>
      <Button asChild size="sm" variant="ghost" className="h-6 w-6 p-0">
        <Link href="/auth/login" title="Sign in">
          <LogOut className="h-3 w-3" />
        </Link>
      </Button>
    </div>
  );
}
