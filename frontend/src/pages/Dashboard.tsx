import { useQuery } from '@tanstack/react-query';
import { getRepos } from '../api';
import { Link, useNavigate } from 'react-router-dom';
import { LogOut, GitBranch, Folder } from 'lucide-react';

interface Repo {
  id: string;
  name: string;
  path: string;
}

export default function Dashboard() {
  const navigate = useNavigate();
  const token = localStorage.getItem('token') || '';

  const { data: repos, isLoading, error } = useQuery({
    queryKey: ['repos'],
    queryFn: () => getRepos(token),
    enabled: !!token,
  });

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/');
  };

  if (!token) {
      // UseNavigate should be used inside useEffect ideally but this works for simple redirect
      setTimeout(() => navigate('/'), 0);
      return null;
  }

  if (isLoading) return <div className="text-white p-8">Loading...</div>;
  if (error) return <div className="text-red-500 p-8">Error loading repos</div>;

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <nav className="border-b border-gray-700 bg-gray-800 p-4">
        <div className="flex items-center justify-between container mx-auto">
          <h1 className="text-xl font-bold flex items-center gap-2">
            <GitBranch className="text-blue-500" /> GitWapp
          </h1>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 rounded-md bg-gray-700 px-3 py-2 text-sm hover:bg-gray-600"
          >
            <LogOut size={16} /> Logout
          </button>
        </div>
      </nav>

      <div className="container mx-auto p-8">
        <h2 className="text-2xl font-bold mb-6">Repositories</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {repos?.map((repo: Repo) => (
            <Link
              key={repo.id}
              to={`/repo/${repo.id}`}
              className="block rounded-lg border border-gray-700 bg-gray-800 p-6 hover:border-blue-500 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-semibold">{repo.name}</h3>
                  <div className="mt-2 flex items-center text-sm text-gray-400">
                    <Folder size={16} className="mr-1" />
                    <span className="truncate max-w-[200px]">{repo.path}</span>
                  </div>
                </div>
              </div>
            </Link>
          ))}
          {repos?.length === 0 && (
            <div className="col-span-full text-center text-gray-500 py-12">
              No repositories tracked. Use the CLI to add one.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
