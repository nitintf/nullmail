"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

interface FloatingInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const FloatingInput = React.forwardRef<HTMLInputElement, FloatingInputProps>(
  ({ className, label, error, helperText, leftIcon, rightIcon, placeholder, ...props }, ref) => {
    return (
      <div className="relative group">
        <div className="relative flex items-center justify-center">
          {leftIcon && (
            <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground z-10">
              {leftIcon}
            </div>
          )}
          
          <input
            ref={ref}
            className={cn(
              "peer w-full h-14 px-4 pt-2 pb-2 text-base bg-transparent border rounded-lg transition-all duration-200 outline-none",
              "border-input hover:border-ring/50 focus:border-ring",
              "focus:ring-ring/50 focus:ring-[3px] focus:shadow-md",
              "disabled:opacity-50 disabled:cursor-not-allowed",
              error && "border-destructive focus:border-destructive focus:ring-destructive/50",
              leftIcon && "pl-10",
              rightIcon && "pr-10",
              className
            )}
            placeholder={'Enter your email name'}
            {...props}
          />
          {rightIcon && (
            <div className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground">
              {rightIcon}
            </div>
          )}
        </div>

        {(error || helperText) && (
          <div className="mt-2 text-sm">
            {error && (
              <p className="text-destructive flex items-center gap-1">
                <span className="w-1 h-1 bg-destructive rounded-full"></span>
                {error}
              </p>
            )}
            {helperText && !error && (
              <p className="text-muted-foreground">{helperText}</p>
            )}
          </div>
        )}
      </div>
    );
  }
);

FloatingInput.displayName = "FloatingInput";

export { FloatingInput }; 