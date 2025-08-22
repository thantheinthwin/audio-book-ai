import { Footer } from "@/components/footer";
import { Header } from "@/components/header";
import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { QueryProvider } from "@/components/providers/query-provider";

export default async function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  return (
    <QueryProvider>
      <div className="flex flex-col min-h-screen">
        <Header user={data.claims.user_metadata} />
        <main className="flex-1 p-4">{children}</main>
        <Footer user={data.claims.user_metadata} />
      </div>
    </QueryProvider>
  );
}
