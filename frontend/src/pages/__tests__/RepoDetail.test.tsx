import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { RepoDetail } from '../RepoDetail';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '../../mocks/server';

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderRepoDetail = (repoId = '1') => {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/repos/${repoId}`]}>
        <Routes>
          <Route path="/repos/:id" element={<RepoDetail />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
};

describe('RepoDetail Page', () => {
  it('renders repo status', async () => {
    server.use(
      http.get('/api/repos/1/status', () => {
        return HttpResponse.json({
          Clean: true,
          Ahead: 1,
          Behind: 2,
          Branch: 'main',
          Worktree: null
        });
      })
    );

    renderRepoDetail();

    await waitFor(() => {
      expect(screen.getByText('main')).toBeInTheDocument();
      expect(screen.getByText('Clean')).toBeInTheDocument();
      expect(screen.getByText('1')).toBeInTheDocument(); // Ahead
      expect(screen.getByText('2')).toBeInTheDocument(); // Behind
    });
  });

  it('renders changes and stages file', async () => {
    const user = userEvent.setup();
    const stageHandler = vi.fn();

    server.use(
      http.get('/api/repos/1/status', () => {
        return HttpResponse.json({
          Clean: false,
          Ahead: 0,
          Behind: 0,
          Branch: 'dev',
          Worktree: {
            "file.txt": { Staging: 0 } // 0 usually means Untracked or Modified depending on map
          }
        });
      }),
      http.post('/api/repos/1/stage', async ({ request }) => {
        const body = await request.json();
        stageHandler(body);
        return HttpResponse.json({});
      })
    );

    renderRepoDetail();

    await waitFor(() => {
      expect(screen.getByText('file.txt')).toBeInTheDocument();
    });

    const stageBtn = screen.getByTitle('Stage');
    await user.click(stageBtn);

    await waitFor(() => {
        expect(stageHandler).toHaveBeenCalledWith({ file: 'file.txt' });
    });
  });

  it('handles commit flow', async () => {
    const user = userEvent.setup();
    const commitHandler = vi.fn();

    server.use(
      http.get('/api/repos/1/status', () => {
        return HttpResponse.json({
          Clean: false,
          Branch: 'main',
          Worktree: {}
        });
      }),
      http.post('/api/repos/1/commit', async ({ request }) => {
        const body = await request.json();
        commitHandler(body);
        return HttpResponse.json({});
      })
    );

    renderRepoDetail();

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument());

    const commitBtn = screen.getByRole('button', { name: /commit/i });
    const textarea = screen.getByPlaceholderText(/commit message/i);

    // Initial state: button disabled
    expect(commitBtn).toBeDisabled();

    // Type message
    await user.type(textarea, 'Fix bug');
    expect(commitBtn).toBeEnabled();

    // Click commit
    await user.click(commitBtn);

    await waitFor(() => {
        expect(commitHandler).toHaveBeenCalledWith({ message: 'Fix bug' });
    });
    
    // Should clear message
    expect(textarea).toHaveValue('');
  });

  it('handles push and pull', async () => {
    const user = userEvent.setup();
    const pushHandler = vi.fn();
    const pullHandler = vi.fn();

    server.use(
      http.get('/api/repos/1/status', () => {
        return HttpResponse.json({ Clean: true, Branch: 'main' });
      }),
      http.post('/api/repos/1/push', () => {
        pushHandler();
        return HttpResponse.json({});
      }),
      http.post('/api/repos/1/pull', () => {
        pullHandler();
        return HttpResponse.json({});
      })
    );

    renderRepoDetail();

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument());

    await user.click(screen.getByRole('button', { name: /push/i }));
    await waitFor(() => expect(pushHandler).toHaveBeenCalled());

    await user.click(screen.getByRole('button', { name: /pull/i }));
    await waitFor(() => expect(pullHandler).toHaveBeenCalled());
  });
});
