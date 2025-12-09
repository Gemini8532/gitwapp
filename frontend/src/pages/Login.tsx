import { useState, type FormEvent } from 'react';
import { useMutation } from '@tanstack/react-query';
import { login } from '../api';
import { useNavigate } from 'react-router-dom';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const mutation = useMutation({
    mutationFn: () => login(username, password),
    onSuccess: (data) => {
      localStorage.setItem('token', data.token);
      navigate('/dashboard');
    },
    onError: (error) => {
      alert(error.message);
    }
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    mutation.mutate();
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-900">
      <div className="w-full max-w-md rounded-lg bg-gray-800 p-8 shadow-lg border border-gray-700">
        <h2 className="mb-6 text-center text-3xl font-bold text-white">Sign In</h2>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-400">Username</label>
            <input
              type="text"
              className="mt-1 block w-full rounded-md border border-gray-600 bg-gray-700 px-3 py-2 text-white placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-400">Password</label>
            <input
              type="password"
              className="mt-1 block w-full rounded-md border border-gray-600 bg-gray-700 px-3 py-2 text-white placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button
            type="submit"
            className="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900"
            disabled={mutation.isPending}
          >
            {mutation.isPending ? 'Signing in...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  );
}
