import React from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { ArrowLeft, FileDiff } from 'lucide-react';
import { DiffViewer } from '../components/DiffViewer';
import { getErrorMessage } from '../utils/errorUtils';

export const RepoDiff: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const [searchParams] = useSearchParams();
    const file = searchParams.get('file');
    const navigate = useNavigate();

    const { data: diffContent, isLoading, error } = useQuery({
        queryKey: ['repo', id, 'diff', file],
        queryFn: async () => {
            if (!file) return null;
            const response = await api.get(`/repos/${id}/diff?file=${encodeURIComponent(file)}`);
            return response.data;
        },
        enabled: !!file,
    });

    if (!file) return <div className="p-8 text-red-500">No file specified</div>;

    return (
        <div className="flex flex-col gap-6 h-[calc(100vh-100px)] w-full px-2 sm:px-4">
            {/* Header */}
            <div className="flex items-center justify-between bg-white dark:bg-gray-800 p-4 rounded-lg shadow shrink-0">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => navigate(-1)}
                        className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full transition-colors"
                        title="Back to Repository"
                    >
                        <ArrowLeft className="w-5 h-5" />
                    </button>
                    <div>
                        <h1 className="text-xl font-bold flex items-center gap-2">
                            <FileDiff className="w-5 h-5 text-blue-500" />
                            {file}
                        </h1>
                        <p className="text-sm text-gray-500 dark:text-gray-400">Diff view</p>
                    </div>
                </div>
            </div>

            {/* Diff Content */}
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow flex-1 overflow-hidden flex flex-col">
                <div className="flex-1 overflow-auto p-4 bg-gray-50 dark:bg-gray-950">
                    {isLoading ? (
                        <div className="flex justify-center items-center h-full">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                        </div>
                    ) : error ? (
                        <div className="p-4 text-red-500 bg-red-50 dark:bg-red-900/10 rounded">
                            Error loading diff: {getErrorMessage(error)}
                        </div>
                    ) : (
                        <DiffViewer diffText={diffContent} />
                    )}
                </div>
            </div>
        </div>
    );
};
