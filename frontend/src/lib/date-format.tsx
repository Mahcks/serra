import React, { createContext, useContext, useEffect, useState } from 'react';
import { backendApi } from '@/lib/api';

interface DateTimeSettings {
  timezone: string;
  dateFormat: string;
  timeFormat: string;
  language: string;
}

interface DateTimeContextType {
  settings: DateTimeSettings;
  formatDate: (date: Date | string) => string;
  formatTime: (date: Date | string) => string;
  formatDateTime: (date: Date | string) => string;
  formatRelativeTime: (date: Date | string) => string;
}

const defaultSettings: DateTimeSettings = {
  timezone: 'UTC',
  dateFormat: 'YYYY-MM-DD',
  timeFormat: '24h',
  language: 'en'
};

const DateTimeContext = createContext<DateTimeContextType | undefined>(undefined);

export function DateTimeProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<DateTimeSettings>(defaultSettings);
  const [settingsLoaded, setSettingsLoaded] = useState(false);

  // Load settings from user preferences
  useEffect(() => {
    const loadSettings = async () => {
      try {
        const userSettings = await backendApi.getUserSettings();
        if (userSettings?.account_settings) {
          const { timezone, date_format, time_format, language } = userSettings.account_settings;
          setSettings({
            timezone: timezone || defaultSettings.timezone,
            dateFormat: date_format || defaultSettings.dateFormat,
            timeFormat: time_format || defaultSettings.timeFormat,
            language: language || defaultSettings.language
          });
        }
      } catch (error) {
        console.log('Could not load date/time settings:', error);
      } finally {
        setSettingsLoaded(true);
      }
    };

    if (!settingsLoaded) {
      loadSettings();
    }
  }, [settingsLoaded]);

  const parseDate = (date: Date | string): Date => {
    return typeof date === 'string' ? new Date(date) : date;
  };

  const formatDate = (date: Date | string): string => {
    const d = parseDate(date);
    if (isNaN(d.getTime())) return 'Invalid Date';

    const options: Intl.DateTimeFormatOptions = {
      timeZone: settings.timezone,
    };

    // Apply date format preference
    switch (settings.dateFormat) {
      case 'MM/DD/YYYY':
        options.month = '2-digit';
        options.day = '2-digit';
        options.year = 'numeric';
        break;
      case 'DD/MM/YYYY':
        options.day = '2-digit';
        options.month = '2-digit';
        options.year = 'numeric';
        break;
      case 'YYYY-MM-DD':
      default:
        options.year = 'numeric';
        options.month = '2-digit';
        options.day = '2-digit';
        break;
    }

    try {
      let formatted = d.toLocaleDateString(settings.language, options);
      
      // Ensure the format matches user preference
      if (settings.dateFormat === 'YYYY-MM-DD') {
        const parts = formatted.split('/');
        if (parts.length === 3) {
          // Convert MM/DD/YYYY or DD/MM/YYYY to YYYY-MM-DD
          const year = parts[2];
          const month = settings.language === 'en' ? parts[0] : parts[1];
          const day = settings.language === 'en' ? parts[1] : parts[0];
          formatted = `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
        }
      }
      
      return formatted;
    } catch (error) {
      return d.toISOString().split('T')[0];
    }
  };

  const formatTime = (date: Date | string): string => {
    const d = parseDate(date);
    if (isNaN(d.getTime())) return 'Invalid Time';

    const options: Intl.DateTimeFormatOptions = {
      timeZone: settings.timezone,
      hour: '2-digit',
      minute: '2-digit',
      hour12: settings.timeFormat === '12h'
    };

    try {
      return d.toLocaleTimeString(settings.language, options);
    } catch (error) {
      return d.toISOString().split('T')[1].slice(0, 5);
    }
  };

  const formatDateTime = (date: Date | string): string => {
    return `${formatDate(date)} ${formatTime(date)}`;
  };

  const formatRelativeTime = (date: Date | string): string => {
    const d = parseDate(date);
    if (isNaN(d.getTime())) return 'Invalid Date';

    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 7) {
      return formatDate(date);
    } else if (days > 0) {
      return `${days}d ago`;
    } else if (hours > 0) {
      return `${hours}h ago`;
    } else if (minutes > 0) {
      return `${minutes}m ago`;
    } else {
      return 'Just now';
    }
  };

  const value: DateTimeContextType = {
    settings,
    formatDate,
    formatTime,
    formatDateTime,
    formatRelativeTime
  };

  return (
    <DateTimeContext.Provider value={value}>
      {children}
    </DateTimeContext.Provider>
  );
}

export const useDateTime = () => {
  const context = useContext(DateTimeContext);
  if (!context) {
    throw new Error('useDateTime must be used within a DateTimeProvider');
  }
  return context;
};