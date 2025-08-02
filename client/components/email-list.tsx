"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SearchInput } from "@/components/ui/search-input";
import { 
  Mail, 
  Download,
  Trash2,
  Check,
  CheckSquare,
  Square,
  Minus
} from "lucide-react";
import { Email } from "@/lib/types";

interface EmailListProps {
  emails: Email[];
  selectedEmail: Email | null;
  onEmailSelect: (email: Email) => void;
}

export function EmailList({ emails, selectedEmail, onEmailSelect }: EmailListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedEmails, setSelectedEmails] = useState<Set<string>>(new Set());
  const [isSelectMode, setIsSelectMode] = useState(false);

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return `${days}d ago`;
  };

  const filteredEmails = emails.filter(email =>
    email.subject.toLowerCase().includes(searchTerm.toLowerCase()) ||
    email.from.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleClearSearch = () => {
    setSearchTerm("");
  };

  const handleEmailClick = (email: Email) => {
    if (isSelectMode) {
      const newSelected = new Set(selectedEmails);
      if (newSelected.has(email.id)) {
        newSelected.delete(email.id);
      } else {
        newSelected.add(email.id);
      }
      setSelectedEmails(newSelected);
    } else {
      onEmailSelect(email);
    }
  };

  const handleSelectModeToggle = () => {
    setIsSelectMode(!isSelectMode);
    setSelectedEmails(new Set());
  };

  const handleSelectAll = () => {
    if (selectedEmails.size === filteredEmails.length) {
      setSelectedEmails(new Set());
    } else {
      setSelectedEmails(new Set(filteredEmails.map(email => email.id)));
    }
  };

  const handleDownload = () => {
    // Simulate download functionality
    console.log('Downloading emails:', Array.from(selectedEmails));
    // Here you would implement actual download logic
  };

  const handleDelete = () => {
    // Simulate delete functionality
    console.log('Deleting emails:', Array.from(selectedEmails));
    setSelectedEmails(new Set());
    setIsSelectMode(false);
    // Here you would implement actual delete logic
  };

  const isAllSelected = filteredEmails.length > 0 && selectedEmails.size === filteredEmails.length;
  const isPartiallySelected = selectedEmails.size > 0 && selectedEmails.size < filteredEmails.length;

  return (
    <div className="w-96 border-r bg-card/50 flex flex-col h-full">
      <div className="p-4 border-b space-y-3 flex-shrink-0">
        <SearchInput
          placeholder="Search emails..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          onClear={handleClearSearch}
        />
        
        {/* Selection Controls */}
        {!isSelectMode ? (
          <div className="flex items-center justify-start">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleSelectModeToggle}
            >
              Select
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectModeToggle}
                className="text-red-600 hover:text-red-700"
              >
                Cancel
              </Button>
              
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectAll}
                className="flex items-center gap-1"
              >
                {isAllSelected ? (
                  <CheckSquare className="h-4 w-4 text-primary" />
                ) : isPartiallySelected ? (
                  <Minus className="h-4 w-4 text-primary" />
                ) : (
                  <Square className="h-4 w-4" />
                )}
                <span className="text-xs">
                  {isAllSelected ? "Deselect All" : "Select All"}
                </span>
              </Button>
            </div>
            
            {selectedEmails.size > 0 && (
              <div className="flex items-center justify-between bg-primary/5 p-2 rounded">
                <Badge variant="secondary" className="text-xs">
                  {selectedEmails.size} selected
                </Badge>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleDownload}
                    className="text-blue-600 hover:text-blue-700"
                  >
                    <Download className="h-4 w-4 mr-1" />
                    Download
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleDelete}
                    className="text-red-600 hover:text-red-700"
                  >
                    <Trash2 className="h-4 w-4 mr-1" />
                    Delete All
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        {filteredEmails.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">
            <Mail className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p className="text-sm font-medium mb-2">No messages found</p>
            <p className="text-xs">
              {searchTerm ? "Try adjusting your search terms" : "Emails will appear here when received"}
            </p>
          </div>
        ) : (
          <div className="divide-y">
            {filteredEmails.map((email) => (
              <div
                key={email.id}
                className={`p-4 cursor-pointer transition-all duration-200 hover:bg-muted/50 ${
                  selectedEmail?.id === email.id && !isSelectMode ? 'bg-muted border-r-2 border-r-primary' : ''
                } ${selectedEmails.has(email.id) ? 'bg-primary/10 border-r-2 border-r-primary' : ''}`}
                onClick={() => handleEmailClick(email)}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      {isSelectMode && (
                        <div className={`w-4 h-4 rounded border-2 flex items-center justify-center ${
                          selectedEmails.has(email.id) 
                            ? 'bg-primary border-primary' 
                            : 'border-muted-foreground/30'
                        }`}>
                          {selectedEmails.has(email.id) && (
                            <Check className="h-3 w-3 text-white" />
                          )}
                        </div>
                      )}
                      <p className="text-sm font-medium truncate">
                        {email.from}
                      </p>
                    </div>
                    <p className="text-sm text-muted-foreground truncate">
                      {email.subject}
                    </p>
                    <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                      {email.body.toString().substring(0, 100)}...
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-1 ml-2">
                    <span className="text-xs text-muted-foreground">
                      {formatTime(email.timestamp)}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
} 