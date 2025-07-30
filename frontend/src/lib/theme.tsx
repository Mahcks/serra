import React, { createContext, useContext, useEffect, useState } from 'react';
import { backendApi } from '@/lib/api';

export type Theme = 'dark' | 'light' | 'system';

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
};

type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
  theme: 'system',
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  defaultTheme = 'system', // Default to system theme
  storageKey = 'serra-ui-theme',
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(() => {
    if (typeof window !== 'undefined') {
      return (localStorage.getItem(storageKey) as Theme) || defaultTheme;
    }
    return defaultTheme;
  });
  const [userSettingsLoaded, setUserSettingsLoaded] = useState(false);

  // Load theme from user settings on app start
  useEffect(() => {
    const loadUserTheme = async () => {
      try {
        const settings = await backendApi.getUserSettings();
        if (settings?.account_settings?.theme) {
          const userTheme = settings.account_settings.theme as Theme;
          if (userTheme !== theme) {
            setTheme(userTheme);
            localStorage.setItem(storageKey, userTheme);
          }
        }
      } catch (error) {
        // User might not be authenticated yet, that's okay
        console.log('Could not load user theme settings:', error);
      } finally {
        setUserSettingsLoaded(true);
      }
    };

    if (!userSettingsLoaded) {
      loadUserTheme();
    }
  }, [userSettingsLoaded, theme, storageKey]);

  useEffect(() => {
    const root = window.document.documentElement;

    root.classList.remove('light', 'dark');

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
        .matches
        ? 'dark'
        : 'light';

      root.classList.add(systemTheme);
      return;
    }

    root.classList.add(theme);
  }, [theme]);

  const value = {
    theme,
    setTheme: (newTheme: Theme) => {
      localStorage.setItem(storageKey, newTheme);
      setTheme(newTheme);
      
      // Also save to user settings
      if (userSettingsLoaded) {
        backendApi.updateUserSettings({
          account_settings: {
            theme: newTheme
          }
        }).catch(error => {
          console.error('Failed to save theme to user settings:', error);
        });
      }
    },
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined)
    throw new Error('useTheme must be used within a ThemeProvider');

  return context;
};