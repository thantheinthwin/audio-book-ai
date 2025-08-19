"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { UserProfile } from "@/components/user-profile";

import {
  ChevronDown,
  ChevronRight,
  BookOpen,
  FileText,
  Menu,
  X,
  Clock,
} from "lucide-react";

interface SidebarItem {
  label: string;
  href?: string;
  icon: React.ReactNode;
  children?: SidebarItem[];
  badge?: string;
}

const platformItems: SidebarItem[] = [
  {
    label: "Audio Books",
    href: "/dashboard/audiobooks",
    icon: <BookOpen className="h-4 w-4" />,
  },
  {
    label: "Add New Book",
    href: "/dashboard/audiobooks/upload",
    icon: <FileText className="h-4 w-4" />,
  },
  {
    label: "Upload Progress",
    href: "/dashboard/uploads/progress",
    icon: <Clock className="h-4 w-4" />,
  },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [expandedItems, setExpandedItems] = useState<string[]>([]);
  const pathname = usePathname();

  const toggleExpanded = (label: string) => {
    setExpandedItems((prev) =>
      prev.includes(label)
        ? prev.filter((item) => item !== label)
        : [...prev, label]
    );
  };

  const SidebarItem = ({
    item,
    level = 0,
  }: {
    item: SidebarItem;
    level?: number;
  }) => {
    const isExpanded = expandedItems.includes(item.label);
    const isActive = item.href && pathname === item.href;
    const hasChildren = item.children && item.children.length > 0;

    return (
      <div>
        <div
          className={cn(
            "flex items-center justify-between px-3 py-2 text-sm rounded-md cursor-pointer transition-colors",
            level > 0 && "ml-4",
            isActive
              ? "bg-primary text-primary-foreground"
              : "hover:bg-accent hover:text-accent-foreground"
          )}
        >
          <div className="flex items-center gap-3 flex-1">
            {item.icon}
            {item.href ? (
              <Link href={item.href} className="flex-1">
                {item.label}
              </Link>
            ) : (
              <span className="flex-1">{item.label}</span>
            )}
          </div>
          {hasChildren && (
            <button
              onClick={() => toggleExpanded(item.label)}
              className="p-1 hover:bg-accent rounded"
            >
              {isExpanded ? (
                <ChevronDown className="h-3 w-3" />
              ) : (
                <ChevronRight className="h-3 w-3" />
              )}
            </button>
          )}
        </div>
        {hasChildren && isExpanded && (
          <div className="mt-1">
            {item.children!.map((child, index) => (
              <SidebarItem key={index} item={child} level={level + 1} />
            ))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-background flex">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          "fixed inset-y-0 left-0 z-50 w-64 bg-muted/50 backdrop-blur-xl border-r transform transition-transform duration-200 ease-in-out lg:translate-x-0 lg:static lg:inset-0",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">A</span>
              </div>
              <div>
                <div className="font-semibold text-sm">Audio Book AI</div>
                <div className="text-xs text-muted-foreground">Testing</div>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="lg:hidden"
              onClick={() => setSidebarOpen(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* Navigation */}
          <div className="flex-1 overflow-y-auto p-4 space-y-6">
            {/* Navigation */}
            <div className="space-y-1">
              {platformItems.map((item, index) => (
                <SidebarItem key={index} item={item} />
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="p-4 border-t space-y-3">
            {/* <div className="flex items-center gap-2 text-sm hover:bg-accent rounded-md px-2 py-1 cursor-pointer">
              <Headphones className="h-4 w-4" />
              <span>Support</span>
            </div>
            <div className="flex items-center gap-2 text-sm hover:bg-accent rounded-md px-2 py-1 cursor-pointer">
              <Send className="h-4 w-4" />
              <span>Feedback</span>
            </div> */}

            {/* User Profile */}
            <UserProfile />
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="w-full">
        {/* Top Navigation */}
        <header className="sticky top-0 z-30 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b">
          <div className="flex items-center justify-between px-4 py-3">
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="sm"
                className="lg:hidden"
                onClick={() => setSidebarOpen(true)}
              >
                <Menu className="h-5 w-5" />
              </Button>
              {/* <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <div className="w-4 h-4 bg-muted rounded" />
                <span>Building Your Application</span>
                <ChevronRight className="h-3 w-3" />
                <span className="text-foreground">Dashboard</span>
              </div> */}
            </div>
            {/* <div className="flex items-center gap-2">
              <ClientAuthButton />
              <ThemeSwitcher />
            </div> */}
          </div>
          {/* Page Content */}
          <main className="p-6">{children}</main>
        </header>
      </div>
    </div>
  );
}
