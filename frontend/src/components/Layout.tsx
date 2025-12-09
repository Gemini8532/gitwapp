import React from 'react';
import { Outlet, useNavigate, useLocation, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '../context/AuthContext';
import { api } from '../services/api';
import { LogOut, Github } from 'lucide-react';

interface Repo {
  id: string;
  name: string;
}

export const Layout: React.FC = () => {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams<{ id: string }>();

  const { data: repos } = useQuery<Repo[]>({
    queryKey: ['repos'],
    queryFn: async () => {
      const response = await api.get('/repos');
      return response.data;
    },
    // We only strictly *need* this if we have an ID, but it's fine to have it cached
    enabled: !!id,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  const currentRepo = repos?.find(r => r.id === id);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const isDiffView = location.pathname.endsWith('/diff');

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <nav className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <div
                className="flex items-center cursor-pointer"
                onClick={() => navigate('/')}
              >
                <Github className="h-8 w-8 text-blue-600" />
                <span className="ml-2 text-xl font-bold">Gitwapp</span>
              </div>
              {currentRepo && (
                <>
                  <span className="mx-3 text-gray-300 text-2xl font-light">/</span>
                  <span className="text-xl font-medium text-gray-700 dark:text-gray-200 truncate max-w-[200px] sm:max-w-md">
                    {currentRepo.name}
                  </span>
                </>
              )}
            </div>
            <div className="flex items-center">
              <button
                onClick={handleLogout}
                className="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400"
                title="Logout"
              >
                <LogOut className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </nav>
      <main className={`mx-auto py-6 sm:px-6 lg:px-8 ${isDiffView ? 'max-w-[100vw] px-0' : 'max-w-7xl'}`}>
        <Outlet />
      </main>
    </div>
  );
};
