import React, { useState, useEffect } from 'react';

type EventRecord = {
  id: string;
  task_id: string | null;
  event_type: string;
  payload: string;
  created_at: string;
};

export default function GlobalTimeline() {
  const [events, setEvents] = useState<EventRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let interval: NodeJS.Timeout;
    const fetchEvents = async () => {
      try {
        const res = await fetch(process.env.NEXT_PUBLIC_API_URL ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1/events` : 'http://localhost:8080/api/v1/events');
        const data = await res.json();
        setEvents(data || []);
      } catch (err) {
        console.error('Failed to fetch events:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchEvents();
    interval = setInterval(fetchEvents, 5000); // Poll every 5s

    return () => clearInterval(interval);
  }, []);

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'TASK_CREATED':
        return (
          <div className="w-8 h-8 rounded-full bg-blue-500/10 flex items-center justify-center border border-blue-500/20 text-blue-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 5v14M5 12h14"/></svg>
          </div>
        );
      case 'TASK_COMPLETED':
        return (
          <div className="w-8 h-8 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20 text-emerald-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
          </div>
        );
      case 'TASK_BLOCKED':
        return (
          <div className="w-8 h-8 rounded-full bg-rose-500/10 flex items-center justify-center border border-rose-500/20 text-rose-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
          </div>
        );
      case 'TASK_ASSIGNED':
        return (
          <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center border border-amber-500/20 text-amber-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          </div>
        );
      case 'BULK_TASKS':
        return (
          <div className="w-8 h-8 rounded-full bg-purple-500/10 flex items-center justify-center border border-purple-500/20 text-purple-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
          </div>
        );
      default:
        return (
          <div className="w-8 h-8 rounded-full bg-slate-500/10 flex items-center justify-center border border-slate-500/20 text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
          </div>
        );
    }
  };

  const parsePayload = (payload: string) => {
    try {
      const parsed = JSON.parse(payload);
      return (
        <div className="mt-2 text-xs text-slate-400 bg-slate-800/30 rounded-lg p-3 border border-slate-700/50">
          <pre className="whitespace-pre-wrap font-mono text-[10px] leading-relaxed">
            {JSON.stringify(parsed, null, 2)}
          </pre>
        </div>
      );
    } catch {
      return <span className="text-slate-400 text-sm">{payload}</span>;
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-end pb-4 border-b border-slate-800">
        {isLoading && (
          <span className="flex h-2 w-2 relative">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
          </span>
        )}
      </div>

      <div className="relative pl-4 md:pl-0 space-y-6 before:absolute before:inset-0 before:ml-8 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-slate-800 before:to-transparent">
        {events.map((event, index) => (
          <div key={event.id} className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
            <div className="flex items-center justify-center w-8 h-8 rounded-full shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 bg-slate-900 border-2 border-slate-800 z-10 transition-transform duration-300 group-hover:scale-110">
               {getEventIcon(event.event_type)}
            </div>

            <div className="w-[calc(100%-3rem)] md:w-[calc(50%-2rem)] p-4 rounded-xl bg-slate-800/20 border border-slate-800/60 shadow-lg backdrop-blur-sm transition-all duration-300 hover:bg-slate-800/40 hover:border-slate-700 hover:shadow-indigo-500/5">
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-bold uppercase tracking-wider text-indigo-400">
                  {event.event_type.replace(/_/g, ' ')}
                </span>
                <span className="text-[10px] text-slate-500 font-medium">
                  {new Date(event.created_at).toLocaleString()}
                </span>
              </div>
              {parsePayload(event.payload)}
            </div>
          </div>
        ))}

        {events.length === 0 && !isLoading && (
          <div className="text-center py-10 text-slate-500 text-sm italic relative z-10 bg-slate-900/50 rounded-xl border border-slate-800/50 backdrop-blur-sm mx-auto w-[calc(100%-3rem)] md:w-1/2">
            No activity events recorded yet.
          </div>
        )}
      </div>
    </div>
  );
}
