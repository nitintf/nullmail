"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { FloatingInput } from "@/components/ui/floating-input";
import { 
  Mail, 
  ArrowRight, 
  Sparkles
} from "lucide-react";
import { useRouter } from "next/navigation";

export function EmailGenerator() {
  const [word, setWord] = useState("");
  const [error, setError] = useState("");
  const router = useRouter();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    
    if (!word.trim()) {
      setError("Please enter a word for your email");
      return;
    }
    
    if (word.trim().length < 3) {
      setError("Email name must be at least 3 characters");
      return;
    }
    
    if (!/^[a-zA-Z0-9_-]+$/.test(word.trim())) {
      setError("Email name can only contain letters, numbers, hyphens, and underscores");
      return;
    }
    
    router.push(`/inbox/${word.trim()}`);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setWord(e.target.value);
    if (error) setError("");
  };

  return (
    <div className="max-w-3xl mx-auto mb-16">
      <div className="text-center mb-8">
        <div className="flex items-center justify-center gap-2 mb-4">
          <Sparkles className="h-6 w-6 text-primary" />
          <h2 className="text-2xl font-bold">Generate Your Disposable Email</h2>
        </div>
        <p className="text-muted-foreground">
          Choose any word and get instant access to a temporary inbox
        </p>
      </div>

      <div className="space-y-6">
        <div className="space-y-4">
          <FloatingInput
            id="word"
            type="text"
            label="Enter your email name"
            placeholder=""
            value={word}
            onChange={handleInputChange}
            error={error}
            helperText="Use letters, numbers, hyphens, and underscores only"
            leftIcon={<Mail className="h-4 w-4" />}
            autoFocus
            className="text-lg"
          />
          
          <div className="flex gap-2">
            <Button 
              type="submit" 
              disabled={!word.trim() || !!error} 
              size="lg"
              className="flex-1 h-12 text-base font-medium shadow-md hover:shadow-lg transition-all duration-200"
              onClick={handleSubmit}
            >
              <ArrowRight className="h-5 w-5 mr-2" />
              Generate Email
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
} 