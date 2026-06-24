'use client';

import React from 'react';

export type TimelineEvent = {
  id: string;
  title: string;
  description: string;
  timestamp: string;
  type: 'creation' | 'update' | 'completion' | 'comment';
};

interface TimelineProps {
  events: TimelineEvent[];
}

export default function Timeline({ events }: TimelineProps) {
  const getEventIcon = (type: TimelineEvent['type']) => {
    switch (type) {
      case 'creation':
        return (
          <svg className="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        );
      case 'completion':
        return (
          <svg className="w-4 h-4 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        );
      case 'comment':
        return (
          <svg className="w-4 h-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        );
      case 'update':
      default:
        return (
          <svg className="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        );
    }
  };

  const getEventBg = (type: TimelineEvent['type']) => {
    switch (type) {
      case 'creation': return 'bg-emerald-500/10 border-emerald-500/20';
      case 'completion': return 'bg-indigo-500/10 border-indigo-500/20';
      case 'comment': return 'bg-amber-500/10 border-amber-500/20';
      case 'update':
      default: return 'bg-slate-500/10 border-slate-500/20';
    }
  };

  if (!events || events.length === 0) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center p-8 text-center bg-slate-900/40 border border-slate-800/80 rounded-2xl shadow-lg shadow-black/20">
        <svg className="w-8 h-8 text-slate-600 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p className="text-sm font-medium text-slate-400">No events to display</p>
        <p className="text-xs text-slate-500 mt-1">Activity will appear here</p>
      </div>
    );
  }

  return (
    <div className="w-full bg-slate-900/40 border border-slate-800/80 rounded-2xl shadow-lg shadow-black/20 p-6 overflow-hidden flex flex-col h-full">
      <h3 className="text-sm font-bold text-slate-200 uppercase tracking-wider mb-6 flex items-center gap-2">
        <svg className="w-4 h-4 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        Activity Timeline
      </h3>

      <div className="relative flex-1 overflow-y-auto pr-2 custom-scrollbar ml-4">
        <div className="absolute top-0 bottom-0 left-[3px] border-l border-slate-700/50"></div>
        <div className="space-y-6 mt-1">
        {events.map((event, index) => (
          <div key={event.id} className="relative pl-8 group">
            <span className={`absolute left-[-13px] top-1 h-8 w-8 rounded-full border flex items-center justify-center bg-slate-900 z-10 transition-transform duration-300 group-hover:scale-110 ${getEventBg(event.type)}`}>
              {getEventIcon(event.type)}
            </span>
            <div className="flex flex-col gap-1">
              <div className="flex items-center justify-between gap-4">
                <h4 className="text-sm font-semibold text-slate-200">{event.title}</h4>
                <time className="text-xs text-slate-500 font-medium whitespace-nowrap">
                  {new Date(event.timestamp).toLocaleString(undefined, {
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                  })}
                </time>
              </div>
              <p className="text-sm text-slate-400 leading-relaxed">{event.description}</p>
            </div>
          </div>
        ))}
        </div>
      </div>
    </div>
  );
}
