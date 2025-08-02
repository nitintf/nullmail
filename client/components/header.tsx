"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { 
  ArrowLeft, 
  Copy, 
  Check, 
  RefreshCw
} from "lucide-react";
import { useRouter } from "next/navigation";

interface HeaderProps {
  emailAddress: string;
  onRefresh?: () => void;
  loading?: boolean;
}

export function Header({ emailAddress, onRefresh, loading = false }: HeaderProps) {
  const [copied, setCopied] = useState(false);
  const router = useRouter();

  const copyEmail = () => {
    navigator.clipboard.writeText(emailAddress);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="border-b bg-card sticky top-0 z-50">
      <div className="px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => router.push("/")}
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
              <h1 className="text-xl font-semibold">Inbox</h1>
              <p className="text-sm text-muted-foreground">{emailAddress}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={copyEmail}
              disabled={copied}
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-600" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="outline"
              size="icon"
              onClick={onRefresh}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
} 