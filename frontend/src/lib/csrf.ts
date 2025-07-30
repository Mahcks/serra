// CSRF token management
let csrfToken: string | null = null;

/**
 * Fetches a new CSRF token from the server
 */
export async function getCSRFToken(): Promise<string> {
  try {
    const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090/v1';
    const response = await fetch(`${baseURL}/csrf-token`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch CSRF token');
    }

    const data = await response.json();
    const token = data.csrf_token;
    console.log('🔒 CSRF token fetched:', token?.substring(0, 10) + '...');
    return token;
  } catch (error) {
    console.error('Error fetching CSRF token:', error);
    throw error;
  }
}

/**
 * Gets the current CSRF token, always fetching a fresh one since tokens are single-use
 */
export async function getCurrentCSRFToken(): Promise<string> {
  // Always fetch a fresh token since backend uses single-use CSRF tokens
  return await getCSRFToken();
}

/**
 * Clears the current CSRF token (forces refetch on next request)
 */
export function clearCSRFToken(): void {
  csrfToken = null;
}

/**
 * Adds CSRF token to request headers
 */
export async function addCSRFTokenToHeaders(headers: Record<string, string> = {}): Promise<Record<string, string>> {
  try {
    const token = await getCurrentCSRFToken();
    return {
      ...headers,
      'X-CSRF-Token': token,
    };
  } catch (error) {
    console.error('Failed to add CSRF token to headers:', error);
    return headers;
  }
}