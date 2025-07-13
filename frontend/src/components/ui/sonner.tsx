import { Toaster as Sonner } from "sonner"
import { useTheme } from "@/lib/theme"
import type { ComponentProps } from "react"

const Toaster = ({ ...props }: ComponentProps<typeof Sonner>) => {
  const { theme } = useTheme()
  
  // Resolve system theme to actual theme for sonner
  const resolvedTheme = theme === 'system' 
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : theme

  return (
    <Sonner
      theme={resolvedTheme as ComponentProps<typeof Sonner>["theme"]}
      className="toaster group"
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
        } as React.CSSProperties
      }
      {...props}
    />
  )
}

export { Toaster }
