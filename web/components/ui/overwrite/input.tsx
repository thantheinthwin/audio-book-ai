import * as React from "react";
import { Eye, EyeOff } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input as BaseInput } from "@/components/ui/input";

interface InputProps extends React.ComponentProps<typeof BaseInput> {
  showPasswordToggle?: boolean;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, showPasswordToggle = false, ...props }, ref) => {
    const [showPassword, setShowPassword] = React.useState(false);
    const isPasswordType = type === "password";
    const shouldShowToggle = showPasswordToggle && isPasswordType;

    const inputType = shouldShowToggle && showPassword ? "text" : type;

    return (
      <div className="relative">
        <BaseInput
          ref={ref}
          type={inputType}
          className={cn(shouldShowToggle && "pr-10", className)}
          {...props}
        />
        {shouldShowToggle && (
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            aria-label={showPassword ? "Hide password" : "Show password"}
          >
            {showPassword ? (
              <EyeOff className="h-4 w-4" />
            ) : (
              <Eye className="h-4 w-4" />
            )}
          </button>
        )}
      </div>
    );
  }
);

Input.displayName = "Input";

export { Input };
