import { http, HttpResponse } from 'msw';

export const handlers = [
  // Example handler
  http.get('/api/health', () => {
    return HttpResponse.json({ status: 'ok' });
  }),
];
