'use client';

import React, { useState, useEffect, useCallback, useMemo } from 'react';

// Strict typing for GDPR compliance (no extra PII fields)
export interface Task {
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
  jira_key?: string; // from backend export
}

// Sorting logic types
type SortField = keyof Task;
type SortDirection = 'asc' | 'desc';

export default function TaskTable() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [sortField, setSortField] = useState<SortField>('updated_at');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError(null);

    // Aggressive Caching: Try loading from sessionStorage first
    const cachedTasks = sessionStorage.getItem('aios_tasks_cache');
    if (cachedTasks) {
      try {
        const parsed = JSON.parse(cachedTasks);
        setTasks(parsed);
        // We still fetch in the background to ensure fresh data, but UI is unblocked
      } catch (e) {
        console.error('Failed to parse cached tasks', e);
      }
    }

    const maxRetries = 3;
    const baseDelay = 1000;
    const timeoutMs = 8000; // Strict timeout boundary

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

      try {
        const res = await fetch('http://localhost:8080/api/v1/tasks', {
          signal: controller.signal,
          headers: {
            'Accept': 'application/json',
          },
        });

        clearTimeout(timeoutId);

        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }

        const data: Task[] = await res.json();

        // Cache the result
        sessionStorage.setItem('aios_tasks_cache', JSON.stringify(data));
        setTasks(data || []);
        setError(null);
        break; // Success, break retry loop

      } catch (err: unknown) {
        clearTimeout(timeoutId);

        const isAbort = err instanceof Error && err.name === 'AbortError';
        const msg = isAbort ? 'Request timed out' : err instanceof Error ? err.message : String(err);

        if (attempt === maxRetries) {
          setError(`Failed to fetch tasks after ${maxRetries} retries: ${msg}`);
          console.error('Task fetch failed entirely:', err);
        } else {
          console.warn(`Fetch attempt ${attempt + 1} failed, retrying... (${msg})`);
          // Exponential backoff
          await new Promise(resolve => setTimeout(resolve, baseDelay * Math.pow(2, attempt)));
        }
      }
    }

    setLoading(false);
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchTasks();
  }, [fetchTasks]);

  const handleSort = (field: SortField) => {
    if (field === sortField) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedTasks = useMemo(() => {
    return [...tasks].sort((a, b) => {
      let aVal = a[sortField];
      let bVal = b[sortField];

      if (aVal === null || aVal === undefined) aVal = '';
      if (bVal === null || bVal === undefined) bVal = '';

      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });
  }, [tasks, sortField, sortDirection]);

  const renderSortIcon = (field: SortField) => {
    if (sortField !== field) return <span className="w-4 h-4 inline-block opacity-0 group-hover:opacity-50">↕</span>;
    return sortDirection === 'asc' ? <span className="w-4 h-4 inline-block text-indigo-400">↑</span> : <span className="w-4 h-4 inline-block text-indigo-400">↓</span>;
  };

  const statusColors: Record<string, string> = {
    'DONE': 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    'IN_PROGRESS': 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    'BLOCKED': 'bg-rose-500/10 text-rose-400 border-rose-500/20',
    'DRAFT': 'bg-slate-500/10 text-slate-400 border-slate-500/20',
  };

  return (
    <div className="flex flex-col h-full w-full bg-[#070b14] border border-slate-800/60 rounded-2xl overflow-hidden shadow-xl">
      {/* Header */}
      <div className="h-16 px-6 border-b border-slate-800/60 flex items-center justify-between bg-slate-900/40 backdrop-blur-md shrink-0">
        <h2 className="font-bold tracking-tight text-slate-100 flex items-center gap-2">
          <svg className="w-4 h-4 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          Task Data Grid
        </h2>
        <div className="flex items-center gap-3">
          <button
            onClick={fetchTasks}
            disabled={loading}
            className="p-2 rounded-lg bg-slate-800 border border-slate-700 text-slate-400 hover:bg-slate-700 hover:text-slate-200 transition-colors disabled:opacity-50"
            title="Refresh Tasks"
          >
            <svg className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="p-4 m-4 bg-rose-500/10 border border-rose-500/20 rounded-xl flex items-center gap-3">
          <svg className="w-5 h-5 text-rose-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm text-rose-200">{error}</p>
        </div>
      )}

      {/* Data Grid */}
      <div className="flex-1 overflow-auto custom-scrollbar">
        <table className="w-full text-left text-sm text-slate-300">
          <thead className="bg-slate-800/50 text-xs uppercase text-slate-500 sticky top-0 z-10 backdrop-blur-md">
            <tr>
              <th className="px-6 py-4 font-semibold tracking-wider cursor-pointer group hover:text-slate-300" onClick={() => handleSort('title')}>
                <div className="flex items-center gap-1">Title {renderSortIcon('title')}</div>
              </th>
              <th className="px-6 py-4 font-semibold tracking-wider cursor-pointer group hover:text-slate-300" onClick={() => handleSort('status')}>
                <div className="flex items-center gap-1">Status {renderSortIcon('status')}</div>
              </th>
              <th className="px-6 py-4 font-semibold tracking-wider cursor-pointer group hover:text-slate-300" onClick={() => handleSort('assignee')}>
                <div className="flex items-center gap-1">Assignee {renderSortIcon('assignee')}</div>
              </th>
              <th className="px-6 py-4 font-semibold tracking-wider cursor-pointer group hover:text-slate-300" onClick={() => handleSort('due_date')}>
                <div className="flex items-center gap-1">Due Date {renderSortIcon('due_date')}</div>
              </th>
              <th className="px-6 py-4 font-semibold tracking-wider cursor-pointer group hover:text-slate-300" onClick={() => handleSort('updated_at')}>
                <div className="flex items-center gap-1">Last Updated {renderSortIcon('updated_at')}</div>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {loading && tasks.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center">
                  <div className="flex justify-center items-center gap-2 text-slate-500">
                    <svg className="animate-spin h-5 w-5 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Loading tasks...
                  </div>
                </td>
              </tr>
            ) : sortedTasks.length > 0 ? (
              sortedTasks.map((task) => (
                <tr key={task.id} className="hover:bg-slate-800/20 transition-colors group">
                  <td className="px-6 py-4">
                    <div className="flex flex-col gap-1">
                      <span className="font-medium text-slate-200">{task.title}</span>
                      <div className="flex flex-wrap gap-1.5 mt-1">
                         {task.source_name && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                            {task.source_name.split(':')[0]}
                          </span>
                        )}
                        {task.client && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-fuchsia-500/10 text-fuchsia-400 border border-fuchsia-500/20">
                            {task.client}
                          </span>
                        )}
                        {task.task_type && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                            {task.task_type}
                          </span>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded-md text-[10px] font-bold border ${statusColors[task.status] || 'bg-slate-800 text-slate-300 border-slate-700'}`}>
                      {task.status}
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
                        <span className="text-sm text-slate-300 truncate max-w-[120px]" title={task.assignee}>
                          {task.assignee}
                        </span>
                      </div>
                    ) : (
                      <span className="text-sm text-slate-500 italic">Unassigned</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-slate-400">
                    {task.due_date ? task.due_date.substring(0, 10) : <span className="text-slate-600">-</span>}
                  </td>
                  <td className="px-6 py-4 text-sm text-slate-500">
                    {new Date(task.updated_at).toLocaleString()}
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center">
                  <div className="flex flex-col items-center justify-center text-slate-500">
                    <svg className="w-8 h-8 mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    <p className="font-medium">No tasks found</p>
                  </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
