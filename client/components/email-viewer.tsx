"use client";

import React from "react";
import { Button } from "@/components/ui/button";
import { 
  Mail, 
  User, 
  Calendar,
  Star,
  Reply,
  Forward,
  MoreVertical,
  Printer,
  Trash2,
  Eye,
  Code
} from "lucide-react";
import { Email } from "@/lib/types";

interface EmailViewerProps {
  email: Email | null;
}

interface ViewMode {
  mode: 'html' | 'text';
}

export function EmailViewer({ email }: EmailViewerProps) {
  const [viewMode, setViewMode] = React.useState<'html' | 'text'>('text');

  const handlePrint = () => {
    window.print();
  };

  const handleReply = () => {
    console.log('Reply to:', email?.id);
  };

  const handleForward = () => {
    console.log('Forward:', email?.id);
  };

  const handleDelete = () => {
    console.log('Delete:', email?.id);
  };

  if (!email) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center text-muted-foreground">
          <Mail className="h-16 w-16 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-medium mb-2">Select a message</h3>
          <p className="text-sm">
            Choose an email from the list to view its content
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {/* Email Header */}
      <div className="border-b p-6 flex-shrink-0">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h2 className="text-xl font-semibold mb-2">{email.subject}</h2>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <User className="h-4 w-4" />
                <span>{email.from}</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                <span>{new Date(email.timestamp).toLocaleString()}</span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={handleReply} title="Reply">
              <Reply className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" onClick={handleForward} title="Forward">
              <Forward className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" onClick={handlePrint} title="Print">
              <Printer className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" onClick={handleDelete} title="Delete">
              <Trash2 className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon">
              <Star className={`h-4 w-4 ${email.starred ? 'fill-yellow-400 text-yellow-400' : ''}`} />
            </Button>
            <div className="flex border rounded-md">
              <Button 
                variant={viewMode === 'text' ? 'default' : 'ghost'} 
                size="sm" 
                onClick={() => setViewMode('text')}
                className="rounded-r-none"
              >
                <Eye className="h-3 w-3 mr-1" />
                Text
              </Button>
              <Button 
                variant={viewMode === 'html' ? 'default' : 'ghost'} 
                size="sm" 
                onClick={() => setViewMode('html')}
                className="rounded-l-none"
              >
                <Code className="h-3 w-3 mr-1" />
                HTML
              </Button>
            </div>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Email Body */}
      <div className="flex-1 p-6 overflow-y-auto min-h-0">
        <div className="prose prose-sm max-w-none">
          {viewMode === 'text' ? (
            <div className="whitespace-pre-wrap text-sm leading-relaxed">
              {email.body.toString()}
            </div>
          ) : (
            <div 
              className="text-sm leading-relaxed"
              dangerouslySetInnerHTML={{ __html: email.body }}
            />
          )}
        </div>
      </div>
    </div>
  );
} 