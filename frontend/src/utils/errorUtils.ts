import axios, { AxiosError } from 'axios';

/**
 * Extracts a user-friendly error message from an error object.
 * Handles Axios errors, Error objects, and unknown error types.
 *
 * @param error The error object caught in a try/catch block or returned by a promise.
 * @param defaultMessage A fallback message if no specific error information is found.
 * @returns A string containing the user-friendly error message.
 */
export const getErrorMessage = (error: unknown, defaultMessage: string = 'An unexpected error occurred'): string => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError;
    // Check if the server returned a specific error message in the response body
    if (axiosError.response?.data) {
        const data = axiosError.response.data;
        if (typeof data === 'string') {
            return data;
        }
        // If the backend returns a JSON object with an 'error' or 'message' field
        if (typeof data === 'object' && data !== null) {
            if ('error' in data && typeof (data as any).error === 'string') {
                return (data as any).error;
            }
            if ('message' in data && typeof (data as any).message === 'string') {
                return (data as any).message;
            }
        }
    }
    // Fallback to HTTP status text if available
    if (axiosError.response?.statusText) {
        return `Error: ${axiosError.response.status} ${axiosError.response.statusText}`;
    }
    // Fallback to the error message from Axios
    return axiosError.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === 'string') {
    return error;
  }

  return defaultMessage;
};
