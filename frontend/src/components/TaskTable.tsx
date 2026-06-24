import React, { useState, useEffect } from 'react';

export type Task = {
  id: string;
  title: string;
  status: string;
  priority: string;
  labels: string;
  project: string;
  jira_key: string;
  team: string;
  task_type: string;
  sprint: string;
  parent_key: string;
  source_name: string;
  due_date: string | null;
  created_at: string;
  updated_at: string;
  assignee?: string; // Adding assignee based on what page.tsx uses for tasks
};

type SortKey = keyof Task;

type TaskTableProps = {
  onClose: () => void;
};

export default function TaskTable({ onClose }: TaskTableProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [sortKey, setSortKey] = useState<SortKey>('updated_at');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/tasks`, {
            headers: {
                ...(process.env.NEXT_PUBLIC_MVP_TOKEN ? { Authorization: `Bearer ${process.env.NEXT_PUBLIC_MVP_TOKEN}` } : {})
            }
        });
        const data = await res.json();
        setTasks(data || []);
      } catch (err) {
        console.error('Failed to fetch tasks:', err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchTasks();
  }, []);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };

  const sortedTasks = [...tasks].sort((a, b) => {
    const aVal = a[sortKey];
    const bVal = b[sortKey];

    if (aVal === bVal) return 0;
    if (aVal === null || aVal === undefined) return sortOrder === 'asc' ? 1 : -1;
    if (bVal === null || bVal === undefined) return sortOrder === 'asc' ? -1 : 1;

    if (aVal < bVal) return sortOrder === 'asc' ? -1 : 1;
    if (aVal > bVal) return sortOrder === 'asc' ? 1 : -1;
    return 0;
  });

  const getSortIcon = (key: SortKey) => {
    if (sortKey !== key) return <span className="ml-1 opacity-0 group-hover:opacity-50">↕</span>;
    return sortOrder === 'asc' ? <span className="ml-1 text-indigo-400">↑</span> : <span className="ml-1 text-indigo-400">↓</span>;
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-[#0B0F19] border border-slate-800 w-full max-w-6xl h-[85vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">

        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50 shrink-0">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
              <svg
                className="w-5 h-5 text-indigo-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
                />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-slate-100">
              Task Data Grid
            </h2>
          </div>
          <button
            onClick={onClose}
            className="h-10 w-10 rounded-xl flex items-center justify-center hover:bg-slate-800 text-slate-500 transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500"
            aria-label="Close task table"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-6 bg-[#0B0F19] custom-scrollbar relative">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center h-full space-y-4">
              <div className="relative h-16 w-16">
                <div className="absolute inset-0 rounded-full border-4 border-indigo-500/20"></div>
                <div className="absolute inset-0 rounded-full border-4 border-t-indigo-500 animate-spin"></div>
              </div>
              <p className="text-slate-400 font-medium">Loading tasks...</p>
            </div>
          ) : tasks.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full">
               <svg
                className="w-12 h-12 text-slate-600 mb-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                />
              </svg>
              <p className="text-slate-400 text-lg">No tasks found.</p>
            </div>
          ) : (
            <div className="rounded-2xl border border-slate-800/80 bg-slate-900/40 overflow-hidden shadow-lg shadow-black/50">
              <table className="w-full text-left text-sm text-slate-300 whitespace-nowrap">
                <thead className="bg-slate-800/80 text-xs uppercase text-slate-400 sticky top-0 z-10 shadow-sm shadow-slate-950">
                  <tr>
                    {[
                      { key: 'jira_key', label: 'Key' },
                      { key: 'title', label: 'Title' },
                      { key: 'status', label: 'Status' },
                      { key: 'priority', label: 'Priority' },
                      { key: 'project', label: 'Project' },
                      { key: 'assignee', label: 'Assignee' },
                      { key: 'due_date', label: 'Due Date' },
                      { key: 'updated_at', label: 'Updated At' },
                    ].map((col) => (
                      <th
                        key={col.key}
                        onClick={() => handleSort(col.key as SortKey)}
                        className="px-6 py-4 font-semibold tracking-wider cursor-pointer hover:bg-slate-700/50 transition-colors group select-none"
                      >
                        <div className="flex items-center">
                          {col.label}
                          {getSortIcon(col.key as SortKey)}
                        </div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/50">
                  {sortedTasks.map((task) => (
                    <tr
                      key={task.id}
                      className="hover:bg-slate-800/40 transition-colors group"
                    >
                      <td className="px-6 py-4 font-mono text-xs text-slate-400 group-hover:text-indigo-300 transition-colors">
                        {task.jira_key || '-'}
                      </td>
                      <td className="px-6 py-4 font-medium text-slate-200 max-w-xs truncate" title={task.title}>
                        {task.title}
                      </td>
                      <td className="px-6 py-4">
                        {task.status === 'BLOCKED' ? (
                          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">BLOCKED</span>
                        ) : task.status === 'IN_PROGRESS' ? (
                          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">IN PROGRESS</span>
                        ) : task.status === 'DONE' ? (
                          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">DONE</span>
                        ) : (
                          <span className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-slate-500/10 text-slate-400 border border-slate-500/20">{task.status || 'DRAFT'}</span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-1 rounded-md text-xs font-medium border ${
                          task.priority?.toLowerCase() === 'high' ? 'bg-rose-500/10 text-rose-400 border-rose-500/20' :
                          task.priority?.toLowerCase() === 'medium' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' :
                          task.priority?.toLowerCase() === 'low' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                          'bg-slate-800 text-slate-400 border-slate-700'
                        }`}>
                          {task.priority || '-'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-slate-400">
                        {task.project || '-'}
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
                           <span className="text-slate-500 italic text-sm">Unassigned</span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-slate-400">
                        {task.due_date ? new Date(task.due_date).toLocaleDateString() : '-'}
                      </td>
                      <td className="px-6 py-4 text-slate-400 text-xs">
                        {new Date(task.updated_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
