import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import { ArrowUp, ArrowDown, Plus, Check } from 'lucide-react';
import clsx from 'clsx';

interface GitStatus {
  Clean: boolean;
  Ahead: number;
  Behind: number;
  Branch: string;
  Worktree: any; // Simplified for now
}

export const RepoDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [commitMessage, setCommitMessage] = useState('');

  const { data: status, isLoading, error } = useQuery<GitStatus>({
    queryKey: ['repo', id, 'status'],
    queryFn: async () => {
      const response = await api.get(`/repos/${id}/status`);
      return response.data;
    },
    refetchInterval: 5000,
  });

  const stageMutation = useMutation({
    mutationFn: (file: string) => api.post(`/repos/${id}/stage`, { file }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['repo', id, 'status'] }),
  });

  const commitMutation = useMutation({
    mutationFn: (message: string) => api.post(`/repos/${id}/commit`, { message }),
    onSuccess: () => {
      setCommitMessage('');
      queryClient.invalidateQueries({ queryKey: ['repo', id, 'status'] });
    },
  });

  const pushMutation = useMutation({
    mutationFn: () => api.post(`/repos/${id}/push`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['repo', id, 'status'] }),
  });

  const pullMutation = useMutation({
    mutationFn: () => api.post(`/repos/${id}/pull`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['repo', id, 'status'] }),
  });

  if (isLoading) return <div>Loading status...</div>;
  if (error) return <div className="text-red-500">Error loading status</div>;
  if (!status) return <div>No status available</div>;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
        <div>
          <h1 className="text-xl font-bold flex items-center gap-2">
            <span className="text-gray-500">Branch:</span> {status.Branch}
          </h1>
          <div className="flex items-center gap-2 mt-1">
            <span className={clsx("px-2 py-0.5 rounded text-sm font-medium", status.Clean ? "bg-green-100 text-green-800" : "bg-yellow-100 text-yellow-800")}>
              {status.Clean ? "Clean" : "Dirty"}
            </span>
            {status.Ahead > 0 && <span className="flex items-center text-blue-600 text-sm"><ArrowUp className="w-4 h-4 mr-1" /> {status.Ahead}</span>}
            {status.Behind > 0 && <span className="flex items-center text-orange-600 text-sm"><ArrowDown className="w-4 h-4 mr-1" /> {status.Behind}</span>}
          </div>
        </div>
        <div className="flex gap-2 mt-4 sm:mt-0">
          <button
            onClick={() => pullMutation.mutate()}
            disabled={pullMutation.isPending || status.Behind === 0}
            className="btn-secondary flex items-center gap-1"
          >
            <ArrowDown className="w-4 h-4" /> Pull
          </button>
          <button
            onClick={() => pushMutation.mutate()}
            disabled={pushMutation.isPending || status.Ahead === 0}
            className="btn-primary flex items-center gap-1"
          >
            <ArrowUp className="w-4 h-4" /> Push
          </button>
        </div>
      </div>

      {/* Changes / Staging */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Working Directory (Simplified view - assuming API returns file list in Worktree or similar, 
             currently our API returns go-git Status object which is map[string]FileStatus) */}
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <h2 className="font-semibold mb-4">Changes</h2>
          {!status.Clean && status.Worktree ? (
            <ul className="space-y-2">
              {Object.entries(status.Worktree).map(([file, stat]: [string, any]) => (
                stat.Staging === 1 ? null : ( // 1 usually means Unmodified in go-git if strictly typed, but here it's serialized JSON. 
                  // Actually go-git Status is a map. '?' means untracked etc.
                  // We need to verify what the API actually sends for "Worktree".
                  // For now, listing all keys in Status map.
                  <li key={file} className="flex justify-between items-center p-2 bg-gray-50 dark:bg-gray-700 rounded">
                    <span className="truncate">{file}</span>
                    <button
                      onClick={() => stageMutation.mutate(file)}
                      className="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded text-blue-600"
                      title="Stage"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                  </li>
                )
              ))}
            </ul>
          ) : <p className="text-gray-500 text-sm">No changes.</p>}
        </div>

        {/* Commit Area */}
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <h2 className="font-semibold mb-4">Commit</h2>
          <textarea
            value={commitMessage}
            onChange={(e) => setCommitMessage(e.target.value)}
            placeholder="Commit message..."
            className="w-full h-32 p-3 border rounded-md dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-blue-500 mb-4"
          />
          <button
            onClick={() => commitMutation.mutate(commitMessage)}
            disabled={!commitMessage || commitMutation.isPending}
            className="w-full btn-primary flex justify-center items-center gap-2"
          >
            <Check className="w-4 h-4" /> Commit
          </button>
        </div>
      </div>
    </div>
  );
};
