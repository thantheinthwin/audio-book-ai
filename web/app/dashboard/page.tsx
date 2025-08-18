import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { InfoIcon, Users, BookOpen, BarChart3 } from "lucide-react";
import { DeployButton } from "@/components/deploy-button";
import { EnvVarWarning } from "@/components/env-var-warning";
import { AuthButton } from "@/components/auth-button";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { hasEnvVars } from "@/lib/utils";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default async function DashboardPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  // Check if user is admin
  const userRole = data.claims.user_metadata.role || "user";
  if (userRole !== "admin") {
    redirect("/");
  }

  return (
    <main className="min-h-screen flex flex-col items-center">
      <div className="flex-1 w-full flex flex-col gap-20 items-center">
        <nav className="w-full flex justify-center border-b border-b-foreground/10 h-16">
          <div className="w-full max-w-5xl flex justify-between items-center p-3 px-5 text-sm">
            <div className="flex gap-5 items-center font-semibold">
              <Link href={"/"}>Audio Book AI</Link>
              <div className="flex items-center gap-2">
                <DeployButton />
              </div>
            </div>
            {!hasEnvVars ? <EnvVarWarning /> : <AuthButton />}
          </div>
        </nav>
        <div className="flex-1 flex flex-col gap-20 max-w-5xl p-5">
          <div className="flex-1 w-full flex flex-col gap-12">
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
                  {JSON.stringify(data.claims, null, 2)}
                </pre>
              </div>
            </div>
          </div>
        </div>

        <footer className="w-full flex items-center justify-center border-t mx-auto text-center text-xs gap-8 py-16">
          <p>
            Powered by{" "}
            <a
              href="https://supabase.com/?utm_source=create-next-app&utm_medium=template&utm_term=nextjs"
              target="_blank"
              className="font-bold hover:underline"
              rel="noreferrer"
            >
              Supabase
            </a>
          </p>
          <ThemeSwitcher />
        </footer>
      </div>
    </main>
  );
}
