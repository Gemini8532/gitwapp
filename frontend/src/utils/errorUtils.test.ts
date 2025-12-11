import { describe, it, expect } from 'vitest';
import { getErrorMessage } from './errorUtils';
import { AxiosError, AxiosHeaders } from 'axios';

describe('getErrorMessage', () => {
  it('returns data from axios error if it is a string', () => {
    const error = new AxiosError(
      'Request failed',
      'ERR_BAD_REQUEST',
      undefined,
      undefined,
      {
        data: 'Server error message',
        status: 400,
        statusText: 'Bad Request',
        headers: {},
        config: { headers: new AxiosHeaders() }
      }
    );
    expect(getErrorMessage(error)).toBe('Server error message');
  });

  it('returns error field from axios error data object', () => {
    const error = new AxiosError(
      'Request failed',
      'ERR_BAD_REQUEST',
      undefined,
      undefined,
      {
        data: { error: 'Detailed error message' },
        status: 400,
        statusText: 'Bad Request',
        headers: {},
        config: { headers: new AxiosHeaders() }
      }
    );
    expect(getErrorMessage(error)).toBe('Detailed error message');
  });

  it('returns message field from axios error data object', () => {
    const error = new AxiosError(
      'Request failed',
      'ERR_BAD_REQUEST',
      undefined,
      undefined,
      {
        data: { message: 'Another detailed message' },
        status: 400,
        statusText: 'Bad Request',
        headers: {},
        config: { headers: new AxiosHeaders() }
      }
    );
    expect(getErrorMessage(error)).toBe('Another detailed message');
  });

  it('fallbacks to status text if no data in axios error', () => {
    const error = new AxiosError(
      'Request failed',
      'ERR_BAD_REQUEST',
      undefined,
      undefined,
      {
        data: null,
        status: 404,
        statusText: 'Not Found',
        headers: {},
        config: { headers: new AxiosHeaders() }
      }
    );
    expect(getErrorMessage(error)).toBe('Error: 404 Not Found');
  });

  it('fallbacks to axios message if no response', () => {
    const error = new AxiosError('Network Error');
    expect(getErrorMessage(error)).toBe('Network Error');
  });

  it('returns message from Error object', () => {
    const error = new Error('Standard error');
    expect(getErrorMessage(error)).toBe('Standard error');
  });

  it('returns string if error is a string', () => {
    expect(getErrorMessage('String error')).toBe('String error');
  });

  it('returns default message for unknown types', () => {
    expect(getErrorMessage(123)).toBe('An unexpected error occurred');
  });
});
