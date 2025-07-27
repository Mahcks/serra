// CSRF token management
let csrfToken: string | null = null;

/**
 * Fetches a new CSRF token from the server
 */
export async function getCSRFToken(): Promise<string> {
  try {
    const response = await fetch('/v1/csrf-token', {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch CSRF token');
    }

    const data = await response.json();
    csrfToken = data.csrf_token;
    return csrfToken;
  } catch (error) {
    console.error('Error fetching CSRF token:', error);
    throw error;
  }
}

/**
 * Gets the current CSRF token, fetching a new one if needed
 */
export async function getCurrentCSRFToken(): Promise<string> {
  if (!csrfToken) {
    return await getCSRFToken();
  }
  return csrfToken;
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