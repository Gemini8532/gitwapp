import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { Login } from '../Login';
import { AuthProvider } from '../../context/AuthContext';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { http, HttpResponse } from 'msw';
import { server } from '../../mocks/server';

// Helper to render with providers
const renderLogin = () => {
  return render(
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<div>Dashboard</div>} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
};

// Since we are not navigating *to* /login in the test but rendering it directly inside router,
// we need to ensure the initial entry is correct if we want to test navigation *away* from it.
// However, the standard BrowserRouter uses window.location. A better approach for testing 
// routing is MemoryRouter. Let's create a custom render helper using MemoryRouter.

import { MemoryRouter } from 'react-router-dom';

const renderLoginWithRouter = () => {
  return render(
    <AuthProvider>
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<div>Dashboard Page</div>} />
        </Routes>
      </MemoryRouter>
    </AuthProvider>
  );
};

describe('Login Page', () => {
  it('renders login form', () => {
    renderLoginWithRouter();
    
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('handles user input', async () => {
    const user = userEvent.setup();
    renderLoginWithRouter();

    const usernameInput = screen.getByLabelText(/username/i);
    const passwordInput = screen.getByLabelText(/password/i);

    await user.type(usernameInput, 'testuser');
    await user.type(passwordInput, 'password123');

    expect(usernameInput).toHaveValue('testuser');
    expect(passwordInput).toHaveValue('password123');
  });

  it('redirects to dashboard on successful login', async () => {
    const user = userEvent.setup();
    
    // Mock successful login
    server.use(
      http.post('/api/login', async () => {
        return HttpResponse.json({ token: 'fake-jwt-token' });
      })
    );

    renderLoginWithRouter();

    await user.type(screen.getByLabelText(/username/i), 'admin');
    await user.type(screen.getByLabelText(/password/i), 'secret');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    // Expect to see Dashboard (redirected)
    await waitFor(() => {
      expect(screen.getByText('Dashboard Page')).toBeInTheDocument();
    });
  });

  it('displays error message on failed login', async () => {
    const user = userEvent.setup();

    // Mock failed login
    server.use(
      http.post('/api/login', () => {
        return new HttpResponse(null, { status: 401 });
      })
    );

    renderLoginWithRouter();

    await user.type(screen.getByLabelText(/username/i), 'wrong');
    await user.type(screen.getByLabelText(/password/i), 'wrong');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    // Expect error message
    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
    });
    
    // Should still be on login page
    expect(screen.queryByText('Dashboard Page')).not.toBeInTheDocument();
  });
});
