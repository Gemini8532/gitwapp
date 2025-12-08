import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

const RepoDetail = () => {
  const { id } = useParams();

  const { data: status, isLoading } = useQuery({
    queryKey: ['repoStatus', id],
    queryFn: async () => {
      const token = localStorage.getItem('token');
      const res = await axios.get(`/api/repos/${id}/status`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      return res.data;
    }
  });

  if (isLoading) return <div className="p-8">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="p-4 bg-white shadow">
        <div className="flex items-center justify-between container mx-auto">
          <div className="flex items-center gap-4">
            <Link to="/" className="text-blue-500 hover:underline">&larr; Back</Link>
            <h1 className="text-xl font-bold">Repo Detail</h1>
          </div>
        </div>
      </nav>

      <div className="container mx-auto p-4 mt-4">
        <div className="bg-white p-4 rounded shadow mb-4">
          <div className="flex gap-4 items-center">
             <div className="font-bold">Branch: <span className="font-normal">main</span></div>
             <button className="px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200">Fetch</button>
             <button className="px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200">Pull</button>
             <button className="px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200">Push</button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-white p-4 rounded shadow">
             <h3 className="font-bold mb-2">Working Directory</h3>
             <p className="text-gray-500 italic">No changes</p>
          </div>
          <div className="bg-white p-4 rounded shadow">
             <h3 className="font-bold mb-2">Staged Changes</h3>
             <p className="text-gray-500 italic">No changes</p>

             <div className="mt-4 border-t pt-4">
                <textarea className="w-full border rounded p-2 mb-2" placeholder="Commit message"></textarea>
                <button className="w-full bg-blue-500 text-white py-2 rounded disabled:bg-blue-300" disabled>Commit</button>
             </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RepoDetail;
