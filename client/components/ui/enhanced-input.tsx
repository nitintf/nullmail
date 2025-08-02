"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Sparkles, Mail } from "lucide-react";

interface EnhancedInputProps extends React.ComponentProps<"input"> {
  label?: string;
  icon?: React.ReactNode;
  error?: string;
  success?: boolean;
  variant?: "default" | "email" | "search";
}

const EnhancedInput = React.forwardRef<HTMLInputElement, EnhancedInputProps>(
  ({ className, label, icon, error, success, variant = "default", ...props }, ref) => {
    const [isFocused, setIsFocused] = React.useState(false);
    const [hasValue, setHasValue] = React.useState(false);

    const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
      setIsFocused(true);
      props.onFocus?.(e);
    };

    const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
      setIsFocused(false);
      props.onBlur?.(e);
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      setHasValue(e.target.value.length > 0);
      props.onChange?.(e);
    };

    const getVariantStyles = () => {
      switch (variant) {
        case "email":
          return "bg-gradient-to-r from-primary/5 to-primary/10 border-primary/20 focus:border-primary";
        case "search":
          return "bg-muted/50 border-muted-foreground/20 focus:border-ring";
        default:
          return "bg-background border-border focus:border-ring";
      }
    };

    return (
      <div className="relative">
        <div
          className={cn(
            "relative flex items-center rounded-xl border-2 transition-all duration-300",
            "focus-within:shadow-lg focus-within:shadow-primary/10",
            getVariantStyles(),
            error && "border-destructive focus-within:border-destructive",
            success && "border-green-500 focus-within:border-green-500",
            className
          )}
        >
          {icon && (
            <div className="absolute left-4 text-muted-foreground transition-colors duration-200">
              {icon}
            </div>
          )}
          
          <input
            ref={ref}
            className={cn(
              "w-full bg-transparent px-4 py-4 text-base font-medium outline-none transition-all duration-200",
              "placeholder:text-muted-foreground/60",
              icon && "pl-12",
              label && "pt-6 pb-2",
              "focus:placeholder:text-muted-foreground/40"
            )}
            onFocus={handleFocus}
            onBlur={handleBlur}
            onChange={handleChange}
            {...props}
          />
          
          {label && (
            <label
              className={cn(
                "absolute left-4 top-4 text-sm font-medium transition-all duration-200 pointer-events-none",
                (isFocused || hasValue) && "text-xs text-primary -translate-y-2",
                !isFocused && !hasValue && "text-muted-foreground",
                icon && "left-12"
              )}
            >
              {label}
            </label>
          )}

          {variant === "email" && hasValue && (
            <div className="absolute right-4 text-primary animate-pulse">
              <Sparkles className="h-4 w-4" />
            </div>
          )}

          {success && (
            <div className="absolute right-4 text-green-500">
              <div className="h-4 w-4 rounded-full bg-green-500 flex items-center justify-center">
                <div className="h-2 w-2 rounded-full bg-white" />
              </div>
            </div>
          )}
        </div>

        {error && (
          <p className="mt-2 text-sm text-destructive animate-in slide-in-from-top-1">
            {error}
          </p>
        )}
      </div>
    );
  }
);

EnhancedInput.displayName = "EnhancedInput";

export { EnhancedInput }; 