import { useQuery } from '@tanstack/react-query';
import { getRepoStatus } from '../api';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, GitCommit, Upload, Download, RefreshCw } from 'lucide-react';

export default function RepoDetail() {
  const { id } = useParams<{ id: string }>();
  const token = localStorage.getItem('token') || '';
  const navigate = useNavigate();

  const { data: status, isLoading, error, refetch } = useQuery({
    queryKey: ['repoStatus', id],
    queryFn: () => getRepoStatus(token, id || ''),
    enabled: !!token && !!id,
  });

  if (!token) {
    setTimeout(() => navigate('/'), 0);
    return null;
  }

  if (isLoading) return <div className="text-white p-8">Loading...</div>;
  if (error) return <div className="text-red-500 p-8">Error loading status</div>;

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <nav className="border-b border-gray-700 bg-gray-800 p-4">
        <div className="container mx-auto flex items-center gap-4">
          <Link to="/dashboard" className="text-gray-400 hover:text-white">
            <ArrowLeft size={20} />
          </Link>
          <h1 className="text-xl font-bold">Repository Status</h1>
        </div>
      </nav>

      <div className="container mx-auto p-8">
        <div className="grid gap-6 md:grid-cols-2">

            {/* Control Bar */}
            <div className="col-span-full bg-gray-800 p-4 rounded-lg border border-gray-700 flex flex-wrap gap-4 items-center justify-between">
                <div className="flex items-center gap-2">
                    <span className="text-gray-400">Branch:</span>
                    <span className="font-mono bg-gray-700 px-2 py-1 rounded">{status.current_branch}</span>
                </div>
                <div className="flex gap-2">
                    <button onClick={() => refetch()} className="flex items-center gap-2 px-3 py-1 bg-gray-700 rounded hover:bg-gray-600">
                        <RefreshCw size={16} /> Fetch
                    </button>
                    <button className="flex items-center gap-2 px-3 py-1 bg-blue-600 rounded hover:bg-blue-500 disabled:opacity-50" disabled={status.behind === 0}>
                        <Download size={16} /> Pull {status.behind > 0 && `(${status.behind})`}
                    </button>
                     <button className="flex items-center gap-2 px-3 py-1 bg-green-600 rounded hover:bg-green-500 disabled:opacity-50" disabled={status.ahead === 0}>
                        <Upload size={16} /> Push {status.ahead > 0 && `(${status.ahead})`}
                    </button>
                </div>
            </div>

          <div className="rounded-lg bg-gray-800 p-6 border border-gray-700">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                <GitCommit className="text-yellow-500"/> Status
            </h2>
            <div className="space-y-2">
                <div className="flex justify-between">
                    <span className="text-gray-400">Clean:</span>
                    <span className={status.is_clean ? "text-green-500" : "text-red-500"}>
                        {status.is_clean ? "Yes" : "No"}
                    </span>
                </div>
                <div className="flex justify-between">
                    <span className="text-gray-400">Has Changes:</span>
                    <span className={status.has_changes ? "text-yellow-500" : "text-gray-500"}>
                        {status.has_changes ? "Yes" : "No"}
                    </span>
                </div>
                 <div className="flex justify-between">
                    <span className="text-gray-400">Ahead:</span>
                    <span className={status.ahead > 0 ? "text-blue-500" : "text-gray-500"}>
                        {status.ahead} commits
                    </span>
                </div>
                 <div className="flex justify-between">
                    <span className="text-gray-400">Behind:</span>
                    <span className={status.behind > 0 ? "text-orange-500" : "text-gray-500"}>
                        {status.behind} commits
                    </span>
                </div>
            </div>
          </div>

           <div className="rounded-lg bg-gray-800 p-6 border border-gray-700">
             <h2 className="text-lg font-semibold mb-4">Actions</h2>
             <p className="text-gray-400 text-sm">
                 Detailed staging and commit features are coming soon.
             </p>
           </div>
        </div>
      </div>
    </div>
  );
}
