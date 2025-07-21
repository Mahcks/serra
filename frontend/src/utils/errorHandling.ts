// Frontend error handling utilities for Serra API errors
import { APIError, APIErrorResponseBodyError } from '../types';

export const ERROR_CODES = {
  // Generic client errors
  UNAUTHORIZED: 10401,
  TOKEN_EXPIRED: 10402,
  INVALID_TOKEN: 10403,
  INSUFFICIENT_PERMISSIONS: 10404,
  BAD_REQUEST: 10405,
  FORBIDDEN: 10406,
  CONFLICT: 10407,
  BAD_GATEWAY: 10408,

  // Client type errors
  VALIDATION_REJECTED: 10410,
  MISSING_ENVIRONMENT_VARIABLE: 10411,

  // Server errors
  INTERNAL_SERVER_ERROR: 10500,
  NOT_FOUND: 10501,
  INVALID_SIGNATURE: 10502,

  // Configuration errors
  NO_RADARR_INSTANCES: 10600,
  NO_SONARR_INSTANCES: 10601,
  INVALID_QUALITY_PROFILE: 10602,
  RADARR_CONNECTION: 10603,
  SONARR_CONNECTION: 10604,

  // Request validation errors
  DUPLICATE_REQUEST: 10610,
  INVALID_MEDIA_TYPE: 10611,
  MISSING_TMDB_ID: 10612,
  INVALID_SEASONS: 10613,
  REQUEST_NOT_APPROVED: 10614,
  SEASON_PARSING_FAILED: 10615,

  // Permission errors
  NO_REQUEST_PERMISSION: 10620,
  NO_APPROVAL_PERMISSION: 10621,
  NO_MANAGE_PERMISSION: 10622,
  NO_4K_PERMISSION: 10623,

  // Processing errors
  RADARR_ADD_FAILED: 10630,
  SONARR_ADD_FAILED: 10631,
  PROCESSING_TIMEOUT: 10632,
} as const;

// Extract API error from different error response formats
export const extractApiError = (error: any): APIError | null => {
  // Handle axios error format
  if (error?.response?.data?.error) {
    return error.response.data.error as APIError;
  }
  
  // Handle direct API error format
  if (error?.error_code && error?.message) {
    return error as APIError;
  }
  
  // Handle nested error
  if (error?.error && error.error.error_code && error.error.message) {
    return error.error as APIError;
  }
  
  return null;
};

// Get user-friendly error message with fallback
export const getErrorMessage = (error: any): string => {
  const apiError = extractApiError(error);
  if (apiError) {
    return apiError.message;
  }
  
  // Fallback patterns for legacy error handling
  if (error?.response?.data?.message) {
    return error.response.data.message;
  }
  
  if (error?.message) {
    return error.message;
  }
  
  return "An unexpected error occurred";
};

// Get error code with fallback
export const getErrorCode = (error: any): number | null => {
  const apiError = extractApiError(error);
  return apiError?.error_code || null;
};

export const getErrorAction = (error: APIError | any): string => {
  const apiError = extractApiError(error);
  const code = apiError?.error_code;
  
  if (!code) {
    return "Please try again or contact support if the problem continues.";
  }

  switch (code) {
    case ERROR_CODES.INSUFFICIENT_PERMISSIONS:
    case ERROR_CODES.UNAUTHORIZED:
    case ERROR_CODES.FORBIDDEN:
      return "You don't have permission to perform this action. Contact your administrator.";
    
    case ERROR_CODES.NO_RADARR_INSTANCES:
    case ERROR_CODES.NO_SONARR_INSTANCES:
    case ERROR_CODES.INVALID_QUALITY_PROFILE:
      return "Contact your administrator to fix the configuration.";
    
    case ERROR_CODES.RADARR_CONNECTION:
    case ERROR_CODES.SONARR_CONNECTION:
      return "Try again later or contact support if the issue persists.";
    
    case ERROR_CODES.DUPLICATE_REQUEST:
      return "Check your existing requests to see the current status.";
    
    case ERROR_CODES.NO_REQUEST_PERMISSION:
    case ERROR_CODES.NO_APPROVAL_PERMISSION:
    case ERROR_CODES.NO_MANAGE_PERMISSION:
    case ERROR_CODES.NO_4K_PERMISSION:
      return "Contact your administrator to request access.";
    
    case ERROR_CODES.INVALID_SEASONS:
      return "Please select valid season numbers and try again.";
    
    case ERROR_CODES.BAD_REQUEST:
    case ERROR_CODES.VALIDATION_REJECTED:
      return "Please check your input and try again.";
    
    default:
      return "Please try again or contact support if the problem continues.";
  }
};

export const isConfigurationError = (error: any): boolean => {
  const code = getErrorCode(error);
  return code ? [
    ERROR_CODES.NO_RADARR_INSTANCES,
    ERROR_CODES.NO_SONARR_INSTANCES,
    ERROR_CODES.INVALID_QUALITY_PROFILE,
  ].includes(code) : false;
};

export const isPermissionError = (error: any): boolean => {
  const code = getErrorCode(error);
  return code ? [
    ERROR_CODES.INSUFFICIENT_PERMISSIONS,
    ERROR_CODES.UNAUTHORIZED,
    ERROR_CODES.FORBIDDEN,
    ERROR_CODES.NO_REQUEST_PERMISSION,
    ERROR_CODES.NO_APPROVAL_PERMISSION,
    ERROR_CODES.NO_MANAGE_PERMISSION,
    ERROR_CODES.NO_4K_PERMISSION,
  ].includes(code) : false;
};

export const isServiceError = (error: any): boolean => {
  const code = getErrorCode(error);
  return code ? [
    ERROR_CODES.RADARR_CONNECTION,
    ERROR_CODES.SONARR_CONNECTION,
    ERROR_CODES.RADARR_ADD_FAILED,
    ERROR_CODES.SONARR_ADD_FAILED,
    ERROR_CODES.BAD_GATEWAY,
  ].includes(code) : false;
};

export const isValidationError = (error: any): boolean => {
  const code = getErrorCode(error);
  return code ? [
    ERROR_CODES.BAD_REQUEST,
    ERROR_CODES.VALIDATION_REJECTED,
    ERROR_CODES.INVALID_MEDIA_TYPE,
    ERROR_CODES.MISSING_TMDB_ID,
    ERROR_CODES.INVALID_SEASONS,
  ].includes(code) : false;
};

// Enhanced error handler for common use in components
export const handleApiError = (error: any, customMessage?: string): { message: string; action: string; type: 'error' | 'warning' | 'info' } => {
  const apiError = extractApiError(error);
  const message = customMessage || getErrorMessage(error);
  const action = getErrorAction(error);
  
  let type: 'error' | 'warning' | 'info' = 'error';
  
  if (isPermissionError(error)) {
    type = 'warning';
  } else if (isValidationError(error)) {
    type = 'info';
  }
  
  return { message, action, type };
};

// Example usage in a React component:
/*
import { handleApiError, isPermissionError, isConfigurationError } from '@/utils/errorHandling';

const MyComponent = () => {
  const handleError = (error: any) => {
    const { message, action, type } = handleApiError(error);
    
    // Show toast with appropriate styling
    if (type === 'warning') {
      toast.warning(message);
    } else if (type === 'info') {
      toast.info(message);
    } else {
      toast.error(message);
    }
    
    // Show action guidance
    setErrorAction(action);
    
    // Handle specific error types
    if (isPermissionError(error)) {
      setShowPermissionRequest(true);
    } else if (isConfigurationError(error)) {
      setShowAdminContact(true);
    }
  };
  
  return (
    // Your component JSX
  );
};
*/