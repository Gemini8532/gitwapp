import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { Link } from 'react-router-dom';
import { FolderGit } from 'lucide-react';

interface Repository {
  id: string;
  name: string;
  path: string;
}

export const Dashboard: React.FC = () => {
  const { data: repos, isLoading, error } = useQuery<Repository[]>({
    queryKey: ['repos'],
    queryFn: async () => {
      const response = await api.get('/repos');
      return response.data;
    },
  });

  if (isLoading) return <div className="text-center py-10">Loading repositories...</div>;
  if (error) return <div className="text-center py-10 text-red-500">Error loading repositories</div>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Repositories</h1>
      {repos?.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 rounded-lg shadow border border-dashed border-gray-300 dark:border-gray-700">
          <FolderGit className="mx-auto h-12 w-12 text-gray-400" />
          <p className="mt-2 text-gray-500 dark:text-gray-400">No repositories tracked yet.</p>
          <p className="text-sm text-gray-400">Use the CLI to add repositories.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {repos?.map((repo) => (
            <Link
              key={repo.id}
              to={`/repos/${repo.id}`}
              className="block p-6 bg-white dark:bg-gray-800 rounded-lg shadow hover:shadow-md transition-shadow border border-gray-200 dark:border-gray-700"
            >
              <div className="flex items-center space-x-3">
                <FolderGit className="h-6 w-6 text-blue-500" />
                <h3 className="text-lg font-medium truncate" title={repo.name}>{repo.name}</h3>
              </div>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 truncate" title={repo.path}>
                {repo.path}
              </p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
};
