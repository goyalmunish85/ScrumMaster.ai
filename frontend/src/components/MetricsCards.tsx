import React from 'react';

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

interface MetricsCardsProps {
  tasks: Task[];
}

export default function MetricsCards({ tasks }: MetricsCardsProps) {
  // Defensive check: unconditionally validate input
  const safeTasks = Array.isArray(tasks) ? tasks : [];

  const totalTasks = safeTasks.length;
  const inProgressTasks = safeTasks.filter((t) => t.status === 'IN_PROGRESS').length;
  const blockedTasks = safeTasks.filter((t) => t.status === 'BLOCKED').length;
  const completedTasks = safeTasks.filter((t) => t.status === 'DONE').length;

  const metrics = [
    {
      label: 'Total Active',
      value: totalTasks,
      color: 'text-indigo-400',
      bgColor: 'bg-indigo-500/10',
      borderColor: 'border-indigo-500/20',
      icon: (
        <svg className="w-5 h-5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      ),
    },
    {
      label: 'In Progress',
      value: inProgressTasks,
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20',
      icon: (
        <svg className="w-5 h-5 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
    },
    {
      label: 'Blocked',
      value: blockedTasks,
      color: 'text-rose-400',
      bgColor: 'bg-rose-500/10',
      borderColor: 'border-rose-500/20',
      icon: (
        <svg className="w-5 h-5 text-rose-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      ),
    },
    {
      label: 'Completed',
      value: completedTasks,
      color: 'text-cyan-400',
      bgColor: 'bg-cyan-500/10',
      borderColor: 'border-cyan-500/20',
      icon: (
        <svg className="w-5 h-5 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      ),
    },
  ];

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 md:px-8 max-w-7xl mx-auto w-full relative z-10">
      {metrics.map((metric, idx) => (
        <div
          key={idx}
          className={`flex items-center gap-4 p-4 rounded-2xl border backdrop-blur-sm transition-all duration-300 hover:scale-[1.02] cursor-default shadow-sm ${metric.bgColor} ${metric.borderColor}`}
        >
          <div className="flex-shrink-0">
            {metric.icon}
          </div>
          <div className="flex flex-col">
            <span className="text-[11px] font-semibold tracking-wider uppercase text-slate-400 mb-0.5">
              {metric.label}
            </span>
            <span className={`text-2xl font-bold tracking-tight ${metric.color}`}>
              {metric.value}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}
