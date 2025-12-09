import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Dashboard } from '../Dashboard';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '../../mocks/server';

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false, // Don't retry in tests
    },
  },
});

const renderDashboard = () => {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Dashboard />
      </MemoryRouter>
    </QueryClientProvider>
  );
};

describe('Dashboard Page', () => {
  it('renders loading state initially', () => {
    renderDashboard();
    expect(screen.getByText(/loading repositories/i)).toBeInTheDocument();
  });

  it('renders list of repositories', async () => {
    server.use(
      http.get('/api/repos', () => {
        return HttpResponse.json([
          { id: '1', name: 'repo-one', path: '/tmp/repo-one' },
          { id: '2', name: 'repo-two', path: '/tmp/repo-two' },
        ]);
      })
    );

    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('repo-one')).toBeInTheDocument();
      expect(screen.getByText('repo-two')).toBeInTheDocument();
    });
    
    expect(screen.getByText('/tmp/repo-one')).toBeInTheDocument();
  });

  it('renders empty state when no repos', async () => {
    server.use(
      http.get('/api/repos', () => {
        return HttpResponse.json([]);
      })
    );

    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText(/no repositories tracked yet/i)).toBeInTheDocument();
    });
  });

  it('handles error state', async () => {
     server.use(
      http.get('/api/repos', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText(/error loading repositories/i)).toBeInTheDocument();
    });
  });
});
