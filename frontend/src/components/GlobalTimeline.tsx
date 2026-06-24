import React, { useState, useEffect } from 'react';

type OperationalEventRecord = {
  id: string;
  task_id: string | null;
  event_type: string;
  payload: string;
  created_at: string;
};

export default function GlobalTimeline() {
  const [events, setEvents] = useState<OperationalEventRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;
    const fetchActivities = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${apiUrl}/api/activities`, {
          headers: {
            'Authorization': `Bearer ${process.env.NEXT_PUBLIC_MVP_TOKEN || 'mvp-token'}`
          }
        });
        if (!res.ok) {
          throw new Error('Failed to fetch activities');
        }
        const data = await res.json();
        if (isMounted) {
          setEvents(data || []);
          setIsLoading(false);
        }
      } catch (err: unknown) {
        if (isMounted) {
          console.error(err);
          if (err instanceof Error) {
            setError(err.message);
          } else {
            setError('An unknown error occurred');
          }
          setIsLoading(false);
        }
      }
    };

    fetchActivities();

    // Poll every 5 seconds
    const interval = setInterval(fetchActivities, 5000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case 'TASK_CREATED':
        return (
          <div className="h-8 w-8 rounded-full bg-emerald-500/20 flex items-center justify-center border border-emerald-500/30">
            <svg className="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </div>
        );
      case 'TASK_COMPLETED':
        return (
          <div className="h-8 w-8 rounded-full bg-blue-500/20 flex items-center justify-center border border-blue-500/30">
            <svg className="w-4 h-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
        );
      case 'TASK_BLOCKED':
        return (
          <div className="h-8 w-8 rounded-full bg-rose-500/20 flex items-center justify-center border border-rose-500/30">
            <svg className="w-4 h-4 text-rose-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        );
      case 'TASK_ASSIGNED':
        return (
          <div className="h-8 w-8 rounded-full bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
            <svg className="w-4 h-4 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
        );
      case 'TASK_STATUS_CHANGED':
        return (
          <div className="h-8 w-8 rounded-full bg-amber-500/20 flex items-center justify-center border border-amber-500/30">
            <svg className="w-4 h-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
        );
      default:
        return (
          <div className="h-8 w-8 rounded-full bg-slate-500/20 flex items-center justify-center border border-slate-500/30">
            <svg className="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        );
    }
  };

  const formatPayload = (payloadStr: string) => {
    try {
      const payload = JSON.parse(payloadStr);
      // Format specifically for task payloads to extract useful info
      if (payload.task_name) {
        return (
          <div className="mt-1 flex flex-col gap-1">
            <span className="font-semibold text-slate-200">{payload.task_name}</span>
            {payload.status && <span className="text-xs text-slate-400">Status: <span className="text-slate-300 font-medium">{payload.status}</span></span>}
            {payload.assignee && <span className="text-xs text-slate-400">Assignee: <span className="text-slate-300 font-medium">{payload.assignee}</span></span>}
          </div>
        );
      }
      if (payload.tasks && Array.isArray(payload.tasks)) {
         return (
             <div className="mt-1">
                 <span className="font-semibold text-slate-200">Bulk sync of {payload.tasks.length} tasks</span>
             </div>
         )
      }
      // Redact potential PII keys (GDPR compliance)
      const safePayload = { ...payload };
      const piiKeys = ['email', 'phone', 'address', 'name', 'password', 'token', 'ssn'];
      for (const key in safePayload) {
        if (piiKeys.some(pii => key.toLowerCase().includes(pii))) {
          safePayload[key] = '***REDACTED***';
        }
      }
      return <pre className="mt-2 text-xs text-slate-400 whitespace-pre-wrap font-mono bg-slate-900/50 p-2 rounded border border-slate-800/50">{JSON.stringify(safePayload, null, 2)}</pre>;
    } catch {
      return <div className="mt-1 text-sm text-slate-400">{payloadStr}</div>;
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-full p-8">
        <svg className="animate-spin h-8 w-8 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm">
        Error loading timeline: {error}
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full w-full custom-scrollbar overflow-y-auto p-2">
      {events.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-slate-500">
          <svg className="w-12 h-12 mb-3 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p>No activity recorded yet.</p>
        </div>
      ) : (
        <div className="relative pl-4 border-l border-slate-700/50 space-y-8 py-4">
          {events.map((event) => (
            <div key={event.id} className="relative group">
              <div className="absolute -left-[37px] top-0 bg-[#0B0F19] py-1">
                {getEventIcon(event.event_type)}
              </div>
              <div className="pl-4">
                <div className="flex flex-col sm:flex-row sm:items-baseline sm:justify-between gap-1 mb-1">
                  <h4 className="text-sm font-bold text-slate-300 flex items-center gap-2">
                    {event.event_type.replace(/_/g, ' ')}
                  </h4>
                  <time className="text-[11px] font-medium text-slate-500 whitespace-nowrap">
                    {new Date(event.created_at).toLocaleString()}
                  </time>
                </div>
                <div className="text-sm">
                  {formatPayload(event.payload)}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
