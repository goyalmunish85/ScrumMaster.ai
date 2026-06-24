import React, { useState, useEffect } from 'react';

type Task = {
  id: string;
  title: string;
  status: string;
  assignee: string | null;
  due_date: string | null;
  updated_at: string;
  client: string | null;
  team: string | null;
  task_type: string | null;
  sprint: string | null;
  source_name: string | null;
};

export default function TaskTable() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [sortField, setSortField] = useState<keyof Task>('updated_at');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;
    const fetchTasks = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${baseUrl}/api/v1/tasks`);
        if (!res.ok) {
          throw new Error(`Failed to fetch tasks: ${res.status}`);
        }
        const data = await res.json();
        if (isMounted) setTasks(data || []);
      } catch (err: unknown) {
        console.error('Failed to fetch tasks:', err);
        if (isMounted) {
          if (err instanceof Error) {
            setError(err.message);
          } else {
            setError('Unknown error');
          }
        }
      } finally {
        if (isMounted) setIsLoading(false);
      }
    };
    fetchTasks();
    return () => { isMounted = false; };
  }, []);

  const handleSort = (field: keyof Task) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const sortedTasks = [...tasks].sort((a, b) => {
    let aVal = a[sortField];
    let bVal = b[sortField];

    if (aVal === null || aVal === undefined) aVal = '';
    if (bVal === null || bVal === undefined) bVal = '';

    if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
    if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
    return 0;
  });

  const renderSortIcon = (field: keyof Task) => {
    if (sortField !== field) return null;
    return (
      <span className="ml-1 text-indigo-400">
        {sortDirection === 'asc' ? '↑' : '↓'}
      </span>
    );
  };

  if (isLoading) {
    return (
      <div className="w-full h-full flex items-center justify-center p-8 text-slate-400">
        <svg className="animate-spin h-6 w-6 mr-3 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Loading tasks...
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center p-8 text-rose-400">
        <svg className="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <span>Error loading tasks: {error}</span>
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col bg-[#0B0F19] rounded-2xl border border-slate-800/80 shadow-2xl overflow-hidden">
      <div className="flex-1 overflow-auto custom-scrollbar">
        <table className="w-full text-left text-sm text-slate-300 min-w-[800px]">
          <thead className="bg-slate-800/60 sticky top-0 z-10 backdrop-blur-md">
            <tr>
              <th
                className="px-6 py-4 font-semibold text-xs uppercase tracking-wider text-slate-400 cursor-pointer hover:bg-slate-700/50 transition-colors select-none group"
                onClick={() => handleSort('title')}
              >
                <div className="flex items-center">
                  Title {renderSortIcon('title')}
                </div>
              </th>
              <th
                className="px-6 py-4 font-semibold text-xs uppercase tracking-wider text-slate-400 cursor-pointer hover:bg-slate-700/50 transition-colors select-none group"
                onClick={() => handleSort('status')}
              >
                <div className="flex items-center">
                  Status {renderSortIcon('status')}
                </div>
              </th>
              <th
                className="px-6 py-4 font-semibold text-xs uppercase tracking-wider text-slate-400 cursor-pointer hover:bg-slate-700/50 transition-colors select-none group"
                onClick={() => handleSort('assignee')}
              >
                <div className="flex items-center">
                  Assignee {renderSortIcon('assignee')}
                </div>
              </th>
              <th
                className="px-6 py-4 font-semibold text-xs uppercase tracking-wider text-slate-400 cursor-pointer hover:bg-slate-700/50 transition-colors select-none group"
                onClick={() => handleSort('due_date')}
              >
                <div className="flex items-center">
                  Due Date {renderSortIcon('due_date')}
                </div>
              </th>
              <th
                className="px-6 py-4 font-semibold text-xs uppercase tracking-wider text-slate-400 cursor-pointer hover:bg-slate-700/50 transition-colors select-none group"
                onClick={() => handleSort('updated_at')}
              >
                <div className="flex items-center">
                  Updated {renderSortIcon('updated_at')}
                </div>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {sortedTasks.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-slate-500 italic">
                  No tasks available.
                </td>
              </tr>
            ) : (
              sortedTasks.map((task) => {
                let statusBadge = "bg-slate-500/10 text-slate-400 border-slate-500/20";
                if (task.status === 'DONE') statusBadge = "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
                if (task.status === 'IN_PROGRESS') statusBadge = "bg-amber-500/10 text-amber-400 border-amber-500/20";
                if (task.status === 'BLOCKED') statusBadge = "bg-rose-500/10 text-rose-400 border-rose-500/20";

                return (
                  <tr key={task.id} className="hover:bg-slate-800/30 transition-colors group">
                    <td className="px-6 py-4">
                      <div className="font-medium text-slate-200 group-hover:text-indigo-300 transition-colors">
                        {task.title}
                      </div>
                      <div className="flex gap-2 mt-1">
                        {task.client && <span className="text-[10px] text-slate-500">{task.client}</span>}
                        {task.source_name && <span className="text-[10px] text-slate-500 bg-slate-800 px-1.5 rounded">{task.source_name.split(':')[0]}</span>}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 rounded-md text-[10px] font-bold border ${statusBadge}`}>
                        {task.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      {task.assignee ? (
                        <div className="flex items-center gap-2">
                          <div className="h-6 w-6 rounded-full bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
                            <span className="text-[10px] font-bold text-indigo-300">
                              {task.assignee.substring(0, 2).toUpperCase()}
                            </span>
                          </div>
                          <span className="text-sm text-slate-300">{task.assignee}</span>
                        </div>
                      ) : (
                        <span className="text-slate-600 italic text-sm">Unassigned</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {task.due_date ? (
                        <span className="text-sm text-slate-300">
                          {new Date(task.due_date).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>
                      ) : (
                        <span className="text-slate-600">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-xs text-slate-500">
                      {new Date(task.updated_at).toLocaleString(undefined, {
                        month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
                      })}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
