"use client";

import { cn } from "@/lib/utils";
import { createClient } from "@/lib/supabase/client";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Input } from "./ui/overwrite/input";

const adminSignUpSchema = z
  .object({
    email: z.string().email("Please enter a valid email address"),
    password: z.string().min(6, "Password must be at least 6 characters"),
    confirmPassword: z.string(),
    invitationCode: z.string().min(1, "Invitation code is required"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type AdminSignUpFormValues = z.infer<typeof adminSignUpSchema>;

// Simple invitation codes for testing
const VALID_INVITATION_CODES = ["ADMIN2024", "SUPERADMIN", "TEST123"];

export function AdminSignUpForm({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const form = useForm<AdminSignUpFormValues>({
    resolver: zodResolver(adminSignUpSchema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
      invitationCode: "",
    },
  });

  const onSubmit = async (data: AdminSignUpFormValues) => {
    // Simple invitation code validation
    if (!VALID_INVITATION_CODES.includes(data.invitationCode)) {
      setError(
        "Invalid invitation code. Please contact the administrator to get an invitation code."
      );
      return;
    }

    const supabase = createClient();
    setIsLoading(true);
    setError(null);

    try {
      // First, create the user account
      const { data: authData, error: authError } = await supabase.auth.signUp({
        email: data.email,
        password: data.password,
        options: {
          emailRedirectTo: `${window.location.origin}/protected`,
          data: {
            role: "admin",
            invitationCode: data.invitationCode, // Store for audit trail
          },
        },
      });

      if (authError) throw authError;

      // If user creation is successful, update the user's role in the database
      if (authData.user) {
        // Note: In a real application, you would typically:
        // 1. Create a custom function in Supabase to handle role assignment
        // 2. Use RLS (Row Level Security) policies to control access
        // 3. Have an admin-only endpoint to assign roles

        // For now, we'll store the role in user metadata
        const { error: updateError } = await supabase.auth.updateUser({
          data: {
            role: "admin",
            invitationCode: data.invitationCode,
          },
        });

        if (updateError) {
          console.warn("Could not update user role:", updateError);
        }
      }

      router.push("/auth/sign-up-success");
    } catch (error: unknown) {
      setError(error instanceof Error ? error.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">Admin Sign Up</CardTitle>
          <CardDescription>
            Create a new admin account (Testing Mode)
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
            <p className="text-sm text-blue-800">
              <strong>Testing Invitation Codes:</strong> Please contact the
              administrator to get an invitation code.
            </p>
          </div>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="invitationCode"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Invitation Code</FormLabel>
                    <FormControl>
                      <Input
                        type="text"
                        placeholder="Enter invitation code"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input
                        type="email"
                        placeholder="admin@example.com"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder="Enter your password"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="confirmPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Confirm Password</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder="Confirm your password"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {error && <p className="text-sm text-red-500">{error}</p>}

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading
                  ? "Creating admin account..."
                  : "Create Admin Account"}
              </Button>
            </form>
          </Form>

          <div className="mt-4 text-center text-sm">
            Already have an account?{" "}
            <Link href="/auth/login" className="underline underline-offset-4">
              Login
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
