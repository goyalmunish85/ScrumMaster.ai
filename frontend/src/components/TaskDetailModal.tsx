'use client';

import React, { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';

interface ActivityResponse {
  id: string;
  task_id: string | null;
  event_type: string;
  payload: unknown;
  created_at: string;
}

interface TaskDetailData {
  id: string;
  title: string;
  description: string;
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
  activities: ActivityResponse[];
}

interface TaskDetailModalProps {
  taskId: string;
  onClose: () => void;
}

export default function TaskDetailModal({ taskId, onClose }: TaskDetailModalProps) {
  const [taskData, setTaskData] = useState<TaskDetailData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  useEffect(() => {
    const fetchTask = async () => {
      setLoading(true);
      setError(null);
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${apiUrl}/api/v1/tasks/${taskId}`);
        if (!res.ok) {
          throw new Error('Failed to fetch task details');
        }
        const data = await res.json();
        setTaskData(data);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError('An error occurred');
        }
      } finally {
        setLoading(false);
      }
    };

    if (taskId) {
      fetchTask();
    }
  }, [taskId]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'BLOCKED': return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
      case 'IN_PROGRESS': return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'DONE': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      default: return 'bg-slate-500/10 text-slate-400 border-slate-500/20';
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
      <div
        className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm transition-opacity"
        onClick={onClose}
        aria-hidden="true"
      ></div>
      <div
        className="relative w-full max-w-4xl max-h-[90vh] bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl shadow-black/50 flex flex-col animate-in fade-in zoom-in-95 duration-200"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-slate-400 hover:text-white bg-slate-800/50 hover:bg-slate-800 rounded-full transition-colors z-10"
          aria-label="Close modal"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        {loading ? (
          <div className="flex-1 flex items-center justify-center p-12">
            <div className="flex flex-col items-center gap-4">
              <div className="w-8 h-8 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
              <p className="text-slate-400 font-medium">Loading task details...</p>
            </div>
          </div>
        ) : error ? (
          <div className="flex-1 flex items-center justify-center p-12">
            <div className="text-center">
              <svg className="w-12 h-12 text-rose-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h3 className="text-lg font-bold text-white mb-2">Error Loading Task</h3>
              <p className="text-slate-400">{error}</p>
            </div>
          </div>
        ) : taskData ? (
          <>
            <div className="p-6 sm:p-8 border-b border-slate-800 flex-shrink-0">
              <div className="flex items-center gap-3 mb-4 flex-wrap">
                <span className={`px-2.5 py-1 rounded-md text-xs font-bold border ${getStatusColor(taskData.status)} uppercase tracking-wider`}>
                  {taskData.status}
                </span>
                {taskData.jira_key && (
                  <span className="px-2.5 py-1 rounded-md text-xs font-medium bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                    {taskData.jira_key}
                  </span>
                )}
                {taskData.task_type && (
                  <span className="px-2.5 py-1 rounded-md text-xs font-medium bg-slate-800 text-slate-300 border border-slate-700">
                    {taskData.task_type}
                  </span>
                )}
              </div>
              <h2 id="modal-title" className="text-2xl font-bold text-white leading-tight">
                {taskData.title}
              </h2>
            </div>

            <div className="flex-1 overflow-y-auto p-6 sm:p-8 custom-scrollbar">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="md:col-span-2 space-y-8">
                  {taskData.description && (
                    <section>
                      <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">Description</h3>
                      <div className="prose prose-invert prose-slate max-w-none text-slate-300 text-sm">
                        <ReactMarkdown>{taskData.description}</ReactMarkdown>
                      </div>
                    </section>
                  )}

                  <section>
                    <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">Activity Log</h3>
                    <div className="space-y-4">
                      {taskData.activities && taskData.activities.length > 0 ? (
                        taskData.activities.map((activity) => (
                          <div key={activity.id} className="bg-slate-800/30 rounded-xl p-4 border border-slate-800/60">
                            <div className="flex justify-between items-start mb-2">
                              <span className="text-sm font-medium text-indigo-300">
                                {activity.event_type}
                              </span>
                              <span className="text-xs text-slate-500">
                                {new Date(activity.created_at).toLocaleString()}
                              </span>
                            </div>
                            <pre className="text-xs text-slate-400 bg-slate-900 p-3 rounded-lg overflow-x-auto border border-slate-800">
                              {JSON.stringify(activity.payload, null, 2)}
                            </pre>
                          </div>
                        ))
                      ) : (
                        <p className="text-sm text-slate-500 italic">No activities found for this task.</p>
                      )}
                    </div>
                  </section>
                </div>

                <div className="space-y-6">
                  <section className="bg-slate-800/20 rounded-xl p-5 border border-slate-800/50">
                    <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">Details</h3>
                    <dl className="space-y-4">
                      {taskData.project && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Project</dt>
                          <dd className="text-sm text-slate-200 font-medium">{taskData.project}</dd>
                        </div>
                      )}
                      {taskData.team && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Team</dt>
                          <dd className="text-sm text-slate-200 font-medium">{taskData.team}</dd>
                        </div>
                      )}
                      {taskData.sprint && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Sprint</dt>
                          <dd className="text-sm text-slate-200 font-medium">{taskData.sprint}</dd>
                        </div>
                      )}
                      {taskData.priority && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Priority</dt>
                          <dd className="text-sm text-slate-200 font-medium">{taskData.priority}</dd>
                        </div>
                      )}
                      {taskData.due_date && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Due Date</dt>
                          <dd className="text-sm text-slate-200 font-medium">
                            {new Date(taskData.due_date).toLocaleDateString()}
                          </dd>
                        </div>
                      )}
                      {taskData.source_name && (
                        <div>
                          <dt className="text-xs text-slate-500 mb-1">Source</dt>
                          <dd className="text-sm text-slate-200 font-medium">{taskData.source_name}</dd>
                        </div>
                      )}
                    </dl>
                  </section>
                </div>
              </div>
            </div>
          </>
        ) : null}
      </div>
    </div>
  );
}
