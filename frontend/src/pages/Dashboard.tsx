import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';

const Dashboard = () => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [newRepoPath, setNewRepoPath] = useState('');

  const { data: repos, isLoading } = useQuery({
    queryKey: ['repos'],
    queryFn: async () => {
      const token = localStorage.getItem('token');
      const res = await axios.get('/api/repos', {
        headers: { Authorization: `Bearer ${token}` }
      });
      return res.data;
    }
  });

  const createRepoMutation = useMutation({
    mutationFn: async (path: string) => {
      const token = localStorage.getItem('token');
      await axios.post('/api/repos', { path }, {
        headers: { Authorization: `Bearer ${token}` }
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['repos'] });
      setNewRepoPath('');
    }
  });

  const handleLogout = () => {
      localStorage.removeItem('token');
      navigate('/login');
  }

  if (isLoading) return <div className="p-8">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="p-4 bg-white shadow">
        <div className="flex items-center justify-between container mx-auto">
          <h1 className="text-xl font-bold">GitWapp Dashboard</h1>
          <button onClick={handleLogout} className="text-red-500 hover:underline">Logout</button>
        </div>
      </nav>

      <div className="container mx-auto p-4 mt-4">
        <div className="mb-6 p-4 bg-white rounded shadow">
          <h2 className="text-lg font-bold mb-2">Add Repository</h2>
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="/absolute/path/to/repo"
              value={newRepoPath}
              onChange={(e) => setNewRepoPath(e.target.value)}
              className="flex-1 px-3 py-2 border rounded"
            />
            <button
              onClick={() => createRepoMutation.mutate(newRepoPath)}
              className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
            >
              Add
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {repos && repos.map((repo: any) => (
            <Link to={`/repo/${repo.ID}`} key={repo.ID} className="block">
              <div className="p-4 bg-white rounded shadow hover:shadow-lg transition">
                <h3 className="text-lg font-bold">{repo.Name}</h3>
                <p className="text-sm text-gray-500 truncate">{repo.Path}</p>
                {/* Status Badges Placeholder */}
                <div className="mt-2 flex gap-1">
                  <span className="px-2 py-0.5 text-xs bg-gray-200 rounded">Clean</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
