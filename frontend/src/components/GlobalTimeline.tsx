import React, { useState, useEffect } from 'react';

type OperationalEventRecord = {
  id: string;
  task_id: string | null;
  event_type: string;
  payload: string;
  created_at: string;
};

export default function GlobalTimeline() {
  const [activities, setActivities] = useState<OperationalEventRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchActivities = async () => {
      try {
        const res = await fetch(`/api/activities`);
        if (res.ok) {
          const data = await res.json();
          setActivities(data);
        }
      } catch (error) {
        console.error('Failed to fetch activities:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchActivities();
  }, []);

  if (isLoading) {
    return <div className="p-4 text-slate-400">Loading timeline...</div>;
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex flex-col gap-3">
        {activities.map((activity) => {
          let payloadData: Record<string, unknown> = {};
          try {
            payloadData = JSON.parse(activity.payload);
          } catch (e) {
            // handle parse error
            console.error('Error parsing activity payload:', e);
          }

          return (
            <div key={activity.id} className="p-4 rounded-xl bg-slate-800/50 border border-slate-700/50 flex flex-col gap-2 relative overflow-hidden group hover:bg-slate-800/80 transition-colors">
              <div className="absolute left-0 top-0 bottom-0 w-1 bg-indigo-500/50 group-hover:bg-indigo-400 transition-colors"></div>
              <div className="flex items-center justify-between">
                <span className="font-semibold text-indigo-400 text-sm flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-indigo-400 animate-pulse"></span>
                  {activity.event_type}
                </span>
                <span className="text-xs text-slate-500">{new Date(activity.created_at).toLocaleString()}</span>
              </div>
              <div className="text-xs text-slate-300 font-mono bg-slate-900/80 p-3 rounded-lg break-all border border-slate-800 whitespace-pre-wrap">
                {JSON.stringify(payloadData, null, 2)}
              </div>
            </div>
          );
        })}
        {activities.length === 0 && (
          <div className="text-slate-500 italic p-4 text-center bg-slate-800/20 rounded-xl border border-slate-700/30">No activities found.</div>
        )}
      </div>
    </div>
  );
}
