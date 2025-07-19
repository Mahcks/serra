// Frontend error handling utilities for Serra API errors

export const ERROR_CODES = {
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

export interface ApiError {
  message: string;
  code: number;
  fields?: Record<string, any>;
}

export const getErrorAction = (error: ApiError): string => {
  switch (error.code) {
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
    
    default:
      return "Please try again or contact support if the problem continues.";
  }
};

export const isConfigurationError = (error: ApiError): boolean => {
  return [
    ERROR_CODES.NO_RADARR_INSTANCES,
    ERROR_CODES.NO_SONARR_INSTANCES,
    ERROR_CODES.INVALID_QUALITY_PROFILE,
  ].includes(error.code);
};

export const isPermissionError = (error: ApiError): boolean => {
  return [
    ERROR_CODES.NO_REQUEST_PERMISSION,
    ERROR_CODES.NO_APPROVAL_PERMISSION,
    ERROR_CODES.NO_MANAGE_PERMISSION,
    ERROR_CODES.NO_4K_PERMISSION,
  ].includes(error.code);
};

export const isServiceError = (error: ApiError): boolean => {
  return [
    ERROR_CODES.RADARR_CONNECTION,
    ERROR_CODES.SONARR_CONNECTION,
    ERROR_CODES.RADARR_ADD_FAILED,
    ERROR_CODES.SONARR_ADD_FAILED,
  ].includes(error.code);
};

// Example usage in a React component:
/*
const handleRequestError = (error: ApiError) => {
  if (isConfigurationError(error)) {
    // Show admin contact banner
    setShowAdminContact(true);
  } else if (isPermissionError(error)) {
    // Show permission request modal
    setShowPermissionRequest(true);
  } else if (isServiceError(error)) {
    // Show service status indicator
    setServiceStatus('degraded');
  }
  
  // Show the user-friendly error message
  toast.error(error.message);
  
  // Show actionable next step
  setErrorAction(getErrorAction(error));
};
*/