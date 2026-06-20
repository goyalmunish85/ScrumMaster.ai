import React, { useState, useEffect } from 'react';

type Target = {
  id: string;
  platform: string;
  target_id: string;
  created_at: string;
};

export default function IntegrationsPanel() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [platform, setPlatform] = useState('slack');
  const [targetId, setTargetId] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const fetchTargets = async () => {
    try {
      const res = await fetch(
        'http://localhost:8080/api/v1/integrations/targets'
      );
      const data = await res.json();
      setTargets(data || []);
    } catch (err) {
      console.error('Failed to fetch targets:', err);
    }
  };

  useEffect(() => {
    fetchTargets();
  }, []);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!platform || !targetId) return;

    setIsLoading(true);
    try {
      await fetch('http://localhost:8080/api/v1/integrations/targets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ platform, target_id: targetId }),
      });
      setTargetId('');
      fetchTargets();
    } catch (err) {
      console.error('Failed to add target:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await fetch(
        `http://localhost:8080/api/v1/integrations/targets?id=${id}`,
        {
          method: 'DELETE',
        }
      );
      fetchTargets();
    } catch (err) {
      console.error('Failed to delete target:', err);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-bold text-slate-200">
          Active Sync Targets
        </h3>
      </div>

      <form
        onSubmit={handleAdd}
        className="flex flex-wrap items-end gap-3 p-4 rounded-2xl bg-slate-800/30 border border-slate-700/60"
      >
        <div className="flex flex-col gap-1.5 flex-1 min-w-[150px]">
          <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
            Platform
          </label>
          <select
            value={platform}
            onChange={(e) => setPlatform(e.target.value)}
            className="h-10 bg-slate-900 border border-slate-700 text-slate-200 rounded-xl px-3 focus:outline-none focus:border-indigo-500"
          >
            <option value="slack">Slack</option>
            <option value="jira">Jira</option>
            <option value="sheets">Google Sheets</option>
            <option value="gitlab">GitLab</option>
          </select>
        </div>
        <div className="flex flex-col gap-1.5 flex-[2] min-w-[200px]">
          <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
            Target ID (Channel, Sheet, Project)
          </label>
          <input
            type="text"
            value={targetId}
            onChange={(e) => setTargetId(e.target.value)}
            placeholder="e.g. C0AQMS8J0P3"
            className="h-10 bg-slate-900 border border-slate-700 text-slate-200 rounded-xl px-3 focus:outline-none focus:border-indigo-500"
          />
        </div>
        <button
          type="submit"
          disabled={isLoading || !targetId}
          className="h-10 px-5 rounded-xl bg-indigo-500 hover:bg-indigo-400 text-white font-medium transition-colors disabled:opacity-50"
        >
          Add Target
        </button>
      </form>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {targets.length > 0 ? (
          targets.map((t) => (
            <div
              key={t.id}
              className="flex items-center justify-between p-3 rounded-xl bg-slate-800/50 border border-slate-700/50"
            >
              <div className="flex items-center gap-3">
                <span
                  className={`px-2 py-1 rounded text-[10px] font-bold uppercase tracking-wider ${
                    t.platform === 'slack'
                      ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                      : t.platform === 'jira'
                        ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                        : t.platform === 'sheets'
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : 'bg-orange-500/10 text-orange-400 border border-orange-500/20'
                  }`}
                >
                  {t.platform}
                </span>
                <span className="text-sm font-medium text-slate-300">
                  {t.target_id}
                </span>
              </div>
              <button
                onClick={() => handleDelete(t.id)}
                className="p-1.5 rounded-lg hover:bg-rose-500/20 text-slate-500 hover:text-rose-400 transition-colors"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M3 6h18" />
                  <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                  <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                </svg>
              </button>
            </div>
          ))
        ) : (
          <p className="text-sm text-slate-500 italic">
            No sync targets configured. Add one above.
          </p>
        )}
      </div>
    </div>
  );
}
