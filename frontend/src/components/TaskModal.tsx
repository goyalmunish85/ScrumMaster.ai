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
  priority: string | null;
  labels: string | null;
  project: string | null;
  jira_key: string | null;
  parent_key: string | null;
};

interface TaskModalProps {
  task: Task | null;
  onClose: () => void;
}

export default function TaskModal({ task, onClose }: TaskModalProps) {
  if (!task) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-slate-900 border border-slate-800 w-full max-w-2xl max-h-[85vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
              <svg className="w-5 h-5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-slate-100 truncate pr-4">
              {task.title}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="h-10 w-10 rounded-xl flex items-center justify-center hover:bg-slate-800 text-slate-500 transition-colors shrink-0"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar bg-[#0B0F19] flex flex-col gap-6 text-slate-300">
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Status</span>
              <span className="text-sm font-medium">{task.status}</span>
            </div>

            {task.assignee && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Assignee</span>
                <div className="flex items-center gap-1.5">
                  <div className="h-5 w-5 rounded-full bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30 shrink-0">
                    <span className="text-[9px] font-bold text-indigo-300">
                      {task.assignee.substring(0, 2).toUpperCase()}
                    </span>
                  </div>
                  <span className="text-sm font-medium truncate">{task.assignee}</span>
                </div>
              </div>
            )}

            {task.priority && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Priority</span>
                <span className="text-sm font-medium">{task.priority}</span>
              </div>
            )}

            {task.due_date && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Due Date</span>
                <span className="text-sm font-medium">{task.due_date.substring(0, 10)}</span>
              </div>
            )}

            {task.project && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Project</span>
                <span className="text-sm font-medium">{task.project}</span>
              </div>
            )}

            {task.jira_key && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Jira Key</span>
                <span className="text-sm font-medium px-2 py-0.5 rounded bg-slate-800 border border-slate-700 inline-block w-max">{task.jira_key}</span>
              </div>
            )}

            {task.client && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Client</span>
                <span className="text-sm font-medium">{task.client}</span>
              </div>
            )}

            {task.team && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Team</span>
                <span className="text-sm font-medium">{task.team}</span>
              </div>
            )}

            {task.task_type && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Task Type</span>
                <span className="text-sm font-medium">{task.task_type}</span>
              </div>
            )}

            {task.sprint && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Sprint</span>
                <span className="text-sm font-medium">{task.sprint}</span>
              </div>
            )}

            {task.source_name && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Source</span>
                <span className="text-sm font-medium">{task.source_name}</span>
              </div>
            )}

            {task.parent_key && (
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Parent</span>
                <span className="text-sm font-medium px-2 py-0.5 rounded bg-slate-800 border border-slate-700 inline-block w-max">{task.parent_key}</span>
              </div>
            )}
          </div>

          {task.labels && (
            <div className="flex flex-col gap-1">
               <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Labels</span>
               <div className="flex flex-wrap gap-2">
                 {task.labels.split(',').map((label, idx) => (
                   <span key={idx} className="px-2 py-1 rounded-md text-[10px] font-medium bg-slate-800 border border-slate-700 text-slate-300">
                     {label.trim()}
                   </span>
                 ))}
               </div>
            </div>
          )}

          <div className="border-t border-slate-800/60 pt-4 mt-auto">
            <div className="flex flex-col gap-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Last Updated</span>
                <span className="text-sm font-medium text-slate-400">
                  {new Date(task.updated_at).toLocaleString()}
                </span>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
