"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { 
  Clock, 
  Mail, 
  ExternalLink
} from "lucide-react";
import { EmailList } from "./email-list";
import { EmailViewer } from "./email-viewer";
import { EmailListSkeleton } from "./email-list-skeleton";
import { EmailViewerSkeleton } from "./email-viewer-skeleton";
import { Header } from "./header";
import { emailsQueryOptions } from "@/lib/queries";
import { Email } from "@/lib/types";

interface InboxContainerProps {
  emailAddress: string;
}

export function InboxContainer({ emailAddress }: InboxContainerProps) {
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
  
  const { data: emails = [], isLoading, error, refetch } = useQuery(emailsQueryOptions(emailAddress));

  const handleEmailSelect = (email: Email) => {
    setSelectedEmail(email);
  };

  return (
    <div className="h-screen bg-background flex flex-col overflow-hidden">
      <Header 
        emailAddress={emailAddress} 
        onRefresh={() => refetch()}
        loading={isLoading}
      />

      <div className="flex flex-1 overflow-hidden">
        {isLoading ? (
          <>
            <EmailListSkeleton />
            <EmailViewerSkeleton />
          </>
        ) : error ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center text-muted-foreground">
              <Mail className="h-16 w-16 mx-auto mb-4 opacity-50" />
              <h3 className="text-lg font-medium mb-2">Error loading emails</h3>
              <p className="text-sm">
                Failed to fetch emails. Please try again.
              </p>
            </div>
          </div>
        ) : (
          <>
            <EmailList 
              emails={emails} 
              selectedEmail={selectedEmail} 
              onEmailSelect={handleEmailSelect} 
            />
            <EmailViewer email={selectedEmail} />
          </>
        )}
      </div>

      {/* Info Panel - Fixed at bottom */}
      <div className="border-t bg-card/50 sticky bottom-0">
        <div className="px-6 py-3">
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                <span>Auto-delete in 24 hours</span>
              </div>
              <div className="flex items-center gap-2">
                <Mail className="h-4 w-4" />
                <span>{emails.length} messages received</span>
              </div>
              <div className="flex items-center gap-2">
                <ExternalLink className="h-4 w-4" />
                <span>Share this address freely</span>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="secondary">
                Inbox
              </Badge>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 