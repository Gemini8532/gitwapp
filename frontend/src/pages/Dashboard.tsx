import React from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import { Link } from 'react-router-dom';
import { FolderGit, RefreshCw, ArrowUp, ArrowDown } from 'lucide-react';
import clsx from 'clsx';
import { shortenPath } from '../utils/pathUtils';
import { getErrorMessage } from '../utils/errorUtils';

interface Repository {
  id: string;
  name: string;
  path: string;
}

interface GitStatus {
  Clean: boolean;
  Ahead: number;
  Behind: number;
  Branch: string;
}

interface RepoWithStatus extends Repository {
  status?: GitStatus;
  statusError?: boolean;
}

export const Dashboard: React.FC = () => {
  const queryClient = useQueryClient();

  const { data: repos, isLoading, error } = useQuery<Repository[]>({
    queryKey: ['repos'],
    queryFn: async () => {
      const response = await api.get('/repos');
      return response.data;
    },
  });

  // Fetch status for each repo
  const reposWithStatus = useQuery<RepoWithStatus[]>({
    queryKey: ['repos-with-status', repos?.map(r => r.id)],
    queryFn: async () => {
      if (!repos) return [];

      const reposWithStatusPromises = repos.map(async (repo) => {
        try {
          const statusResponse = await api.get(`/repos/${repo.id}/status`);
          return { ...repo, status: statusResponse.data };
        } catch (error) {
          return { ...repo, statusError: true };
        }
      });

      return Promise.all(reposWithStatusPromises);
    },
    enabled: !!repos && repos.length > 0,
  });

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['repos'] });
    queryClient.invalidateQueries({ queryKey: ['repos-with-status'] });
  };

  if (isLoading) return <div className="text-center py-10">Loading repositories...</div>;
  if (error) return <div className="text-center py-10 text-red-500">Error loading repositories: {getErrorMessage(error)}</div>;

  const displayRepos: RepoWithStatus[] = reposWithStatus.data || repos?.map(r => ({ ...r })) || [];

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Repositories</h1>
        <button
          onClick={handleRefresh}
          disabled={reposWithStatus.isLoading}
          className="btn-secondary flex items-center gap-2"
          title="Refresh"
        >
          <RefreshCw className={clsx("w-4 h-4", reposWithStatus.isLoading && "animate-spin")} />
          Refresh
        </button>
      </div>

      {repos?.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 rounded-lg shadow border border-dashed border-gray-300 dark:border-gray-700">
          <FolderGit className="mx-auto h-12 w-12 text-gray-400" />
          <p className="mt-2 text-gray-500 dark:text-gray-400">No repositories tracked yet.</p>
          <p className="text-sm text-gray-400">Use the CLI to add repositories.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {displayRepos.map((repo) => (
            <Link
              key={repo.id}
              to={`/repos/${repo.id}`}
              className="block p-6 bg-white dark:bg-gray-800 rounded-lg shadow hover:shadow-md transition-shadow border border-gray-200 dark:border-gray-700"
            >
              <div className="flex items-center space-x-3">
                <FolderGit className="h-6 w-6 text-blue-500" />
                <h3 className="text-lg font-medium truncate" title={repo.name}>{shortenPath(repo.name)}</h3>
              </div>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 truncate" title={repo.path}>
                {shortenPath(repo.path)}
              </p>

              {/* Status Summary */}
              {repo.status && (
                <div className="mt-3 flex items-center gap-2 flex-wrap">
                  <span className={clsx(
                    "px-2 py-0.5 rounded text-xs font-medium",
                    repo.status.Clean ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200" : "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200"
                  )}>
                    {repo.status.Clean ? "Clean" : "Dirty"}
                  </span>
                  <span className="text-xs text-gray-500">{repo.status.Branch}</span>
                  {repo.status.Ahead > 0 && (
                    <span className="flex items-center text-blue-600 dark:text-blue-400 text-xs">
                      <ArrowUp className="w-3 h-3 mr-0.5" /> {repo.status.Ahead}
                    </span>
                  )}
                  {repo.status.Behind > 0 && (
                    <span className="flex items-center text-orange-600 dark:text-orange-400 text-xs">
                      <ArrowDown className="w-3 h-3 mr-0.5" /> {repo.status.Behind}
                    </span>
                  )}
                </div>
              )}
              {repo.statusError && (
                <p className="mt-3 text-xs text-red-500">Failed to load status</p>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
};
