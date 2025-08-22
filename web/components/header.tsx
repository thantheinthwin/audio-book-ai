"use client";

import Link from "next/link";

interface User {
  id: string;
  email: string;
  role: string;
}

interface HeaderProps {
  user: User | null;
}

export function Header({ user }: HeaderProps) {
  return (
    <header className="flex justify-between items-center p-4">
      <Navigation user={user} />
    </header>
  );
}

interface NavItem {
  label: string;
  href: string;
}

const navItems: Record<User["role"], NavItem[]> = {
  admin: [{ label: "Audio Books", href: "/audiobooks" }],
  user: [{ label: "Library", href: "/library" }],
};

interface NavigationProps {
  user: User | null;
}

const Navigation = ({ user }: NavigationProps) => {
  return (
    <nav>
      <ul className="flex gap-2">
        {navItems[user?.role || "user"].map((item) => (
          <li key={item.href}>
            <Link
              href={item.href}
              className="hover:underline hover:text-primary text-sm"
            >
              {item.label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
};
