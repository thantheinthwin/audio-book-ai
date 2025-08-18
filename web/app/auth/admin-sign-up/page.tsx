import { AdminSignUpForm } from "@/components/admin-sign-up-form";

export default function AdminSignUpPage() {
  return (
    <div className="flex-1 flex flex-col w-full px-8 sm:max-w-md justify-center gap-2">
      <AdminSignUpForm />
    </div>
  );
}
