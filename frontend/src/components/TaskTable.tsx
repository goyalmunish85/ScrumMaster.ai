'use client';

import React from 'react';

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
  tasks: Task[];
  onRowClick?: (taskId: string) => void;
}

export default function TaskTable({ tasks, onRowClick }: TaskTableProps) {
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

  return (
    <div className="w-full bg-slate-900/40 border border-slate-800/80 rounded-2xl overflow-hidden shadow-lg shadow-black/20">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-slate-300 border-collapse">
          <thead className="bg-slate-800/50 text-xs uppercase text-slate-500 tracking-wider">
            <tr>
              <th scope="col" className="px-6 py-4 font-semibold">Title</th>
              <th scope="col" className="px-6 py-4 font-semibold">Status</th>
              <th scope="col" className="px-6 py-4 font-semibold">Assignee</th>
              <th scope="col" className="px-6 py-4 font-semibold">Due Date</th>
              <th scope="col" className="px-6 py-4 font-semibold">Source</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {tasks.length > 0 ? (
              tasks.map((task) => (
                <tr
                  key={task.id}
                  className={`hover:bg-slate-800/30 transition-colors duration-200 group ${onRowClick ? 'cursor-pointer' : ''}`}
                  onClick={() => onRowClick && onRowClick(task.id)}
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
