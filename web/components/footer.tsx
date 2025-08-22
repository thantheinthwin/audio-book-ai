import { LogoutButton } from "./logout-button";
import { ThemeSwitcher } from "./theme-switcher";

interface User {
  id: string;
  email: string;
  role: string;
}

interface FooterProps {
  user: User | null;
}

export function Footer({ user }: FooterProps) {
  return (
    <footer className="flex justify-between items-center p-4">
      <ThemeSwitcher />
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">{user?.email}</span>
        <LogoutButton />
      </div>
    </footer>
  );
}
