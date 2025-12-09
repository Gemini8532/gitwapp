export const API_URL = import.meta.env.VITE_API_URL || '';

export async function login(username: string, password: string): Promise<{ token: string }> {
  const res = await fetch(`${API_URL}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error('Login failed');
  return res.json();
}

export async function getRepos(token: string): Promise<any[]> {
  const res = await fetch(`${API_URL}/api/repos`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to fetch repos');
  return res.json();
}

export async function getRepoStatus(token: string, id: string): Promise<any> {
  const res = await fetch(`${API_URL}/api/repos/${id}/status`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to fetch repo status');
  return res.json();
}
