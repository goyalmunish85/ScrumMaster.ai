'use client';

import React, { useState, useEffect, useCallback, useMemo } from 'react';

export type Task = {
  id: string;
  title: string;
  status: 'BLOCKED' | 'IN_PROGRESS' | 'DRAFT' | 'DONE';
  assignee: string | null;
  due_date: string | null;
  updated_at: string;
  client: string | null;
  team: string | null;
  task_type: string | null;
  sprint: string | null;
  source_name: string | null;
};

interface TaskTableProps {
  searchQuery?: string;
}

type SortField = 'title' | 'status' | 'assignee' | 'due_date' | 'source_name';
type SortDirection = 'asc' | 'desc';

export default function TaskTable({ searchQuery = '' }: TaskTableProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [sortField, setSortField] = useState<SortField>('title');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  useEffect(() => {
    let isMounted = true;

    const fetchTasks = async (retryCount = 0) => {
      const maxRetries = 3;
      const baseDelay = 1000;

      setIsLoading(true);
      setError(null);

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 8000); // Strict 8s timeout

      try {
        const res = await fetch('/api/tasks', {
          signal: controller.signal,
        });
        clearTimeout(timeoutId);

        if (!res.ok) {
          throw new Error(`Failed to fetch tasks: ${res.statusText}`);
        }

        const data = await res.json();
        if (isMounted) {
          setTasks(data || []);
          setIsLoading(false);
        }
      } catch (err: unknown) {
        clearTimeout(timeoutId);
        if (!isMounted) return;

        if (err instanceof Error && err.name === 'AbortError') {
          // It's a timeout
        }

        if (retryCount < maxRetries) {
          console.warn(`Fetch failed, retrying (${retryCount + 1}/${maxRetries})...`, err);
          setTimeout(() => {
            if (isMounted) fetchTasks(retryCount + 1);
          }, baseDelay * Math.pow(2, retryCount));
        } else {
          console.error(`Max retries reached (${maxRetries}) or unrecoverable error:`, err);
          setError('Failed to load tasks. Please try again later.');
          setIsLoading(false);
        }
      }
    };

    fetchTasks();

    return () => {
      isMounted = false;
    };
  }, []);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const getStatusWeight = (status: string) => {
    switch (status) {
      case 'BLOCKED': return 0;
      case 'IN_PROGRESS': return 1;
      case 'DRAFT': return 2;
      case 'DONE': return 3;
      default: return 4;
    }
  };

  const filteredAndSortedTasks = useMemo(() => {
    let result = [...tasks];

    if (searchQuery) {
      const lowerQuery = searchQuery.toLowerCase();
      result = result.filter(
        (task) =>
          task.title.toLowerCase().includes(lowerQuery) ||
          (task.assignee && task.assignee.toLowerCase().includes(lowerQuery)) ||
          (task.source_name && task.source_name.toLowerCase().includes(lowerQuery))
      );
    }

    result.sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case 'title':
          cmp = a.title.localeCompare(b.title);
          break;
        case 'status':
          cmp = getStatusWeight(a.status) - getStatusWeight(b.status);
          break;
        case 'assignee':
          cmp = (a.assignee || '').localeCompare(b.assignee || '');
          break;
        case 'due_date':
          const dateA = a.due_date ? new Date(a.due_date).getTime() : 0;
          const dateB = b.due_date ? new Date(b.due_date).getTime() : 0;
          cmp = dateA - dateB;
          break;
        case 'source_name':
          cmp = (a.source_name || '').localeCompare(b.source_name || '');
          break;
        default:
          cmp = 0;
      }
      return sortDirection === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [tasks, searchQuery, sortField, sortDirection]);

  const getStatusBadge = (status: Task['status']) => {
    switch (status) {
      case 'BLOCKED':
        return (
          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20 flex items-center gap-1.5 w-max">
            <span className="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"></span>
            BLOCKED
          </span>
        );
      case 'IN_PROGRESS':
        return (
          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 flex items-center gap-1.5 w-max">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
            IN PROGRESS
          </span>
        );
      case 'DONE':
        return (
          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5 w-max">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
            DONE
          </span>
        );
      default:
        return (
          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-slate-500/10 text-slate-400 border border-slate-500/20 flex items-center gap-1.5 w-max">
            <span className="w-1.5 h-1.5 rounded-full bg-slate-500"></span>
            DRAFT
          </span>
        );
    }
  };

  const renderSortIcon = (field: SortField) => {
    if (sortField !== field) return null;
    return (
      <span className="ml-1 text-indigo-400">
        {sortDirection === 'asc' ? '↑' : '↓'}
      </span>
    );
  };

  return (
    <div className="w-full bg-slate-900/40 border border-slate-800/80 rounded-2xl overflow-hidden shadow-lg shadow-black/20">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-slate-300 border-collapse">
          <thead className="bg-slate-800/50 text-xs uppercase text-slate-500 tracking-wider">
            <tr>
              <th scope="col" className="px-6 py-4 font-semibold cursor-pointer hover:bg-slate-800 transition-colors" onClick={() => handleSort('title')}>
                Title {renderSortIcon('title')}
              </th>
              <th scope="col" className="px-6 py-4 font-semibold cursor-pointer hover:bg-slate-800 transition-colors" onClick={() => handleSort('status')}>
                Status {renderSortIcon('status')}
              </th>
              <th scope="col" className="px-6 py-4 font-semibold cursor-pointer hover:bg-slate-800 transition-colors" onClick={() => handleSort('assignee')}>
                Assignee {renderSortIcon('assignee')}
              </th>
              <th scope="col" className="px-6 py-4 font-semibold cursor-pointer hover:bg-slate-800 transition-colors" onClick={() => handleSort('due_date')}>
                Due Date {renderSortIcon('due_date')}
              </th>
              <th scope="col" className="px-6 py-4 font-semibold cursor-pointer hover:bg-slate-800 transition-colors" onClick={() => handleSort('source_name')}>
                Source {renderSortIcon('source_name')}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {isLoading ? (
               <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-slate-500 italic">
                  Loading tasks...
                </td>
              </tr>
            ) : error ? (
               <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-rose-500 italic">
                  {error}
                </td>
              </tr>
            ) : filteredAndSortedTasks.length > 0 ? (
              filteredAndSortedTasks.map((task) => (
                <tr
                  key={task.id}
                  className="hover:bg-slate-800/30 transition-colors duration-200 group"
                >
                  <td className="px-6 py-4 font-medium text-slate-200">
                    {task.title}
                  </td>
                  <td className="px-6 py-4">
                    {getStatusBadge(task.status)}
                  </td>
                  <td className="px-6 py-4 text-slate-400">
                    {task.assignee ? (
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-indigo-500/20 flex items-center justify-center text-xs font-medium text-indigo-300 border border-indigo-500/30">
                          {task.assignee.charAt(0).toUpperCase()}
                        </div>
                        {task.assignee}
                      </div>
                    ) : (
                      <span className="italic text-slate-500">Unassigned</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-slate-400">
                    {task.due_date ? (
                      new Date(task.due_date).toLocaleDateString()
                    ) : (
                      '-'
                    )}
                  </td>
                  <td className="px-6 py-4">
                    {task.source_name ? (
                      <span className="px-2 py-1 rounded-md text-[10px] font-medium bg-slate-800 text-slate-300 border border-slate-700">
                        {task.source_name}
                      </span>
                    ) : (
                      '-'
                    )}
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-slate-500 italic">
                  No tasks found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
