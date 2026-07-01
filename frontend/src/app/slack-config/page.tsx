'use client';

import React, { useState, useEffect } from 'react';

type Target = {
  id: string;
  platform: string;
  target_id: string;
  created_at: string;
};

export default function SlackConfigPage() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [targetId, setTargetId] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isValidSlackChannelId = (id: string): boolean => {
    return /^C[A-Z0-9]+$/.test(id);
  };

  const fetchTargets = async () => {
    try {
      const res = await fetch(`/api/integrations/targets`);
      if (!res.ok) throw new Error('Failed to fetch targets');
      const data = await res.json();
      setTargets(data?.filter((t: Target) => t.platform === 'slack') || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch targets:', err);
      setError('Failed to fetch Slack channels. Please try again.');
    }
  };

  useEffect(() => {
    let isMounted = true;
    const fetchInit = async () => {
      try {
        const res = await fetch(`/api/integrations/targets`);
        if (!res.ok) throw new Error('Failed to fetch targets');
        const data = await res.json();
        if (isMounted) {
          setTargets(data?.filter((t: Target) => t.platform === 'slack') || []);
        }
      } catch (err) {
        console.error('Failed to fetch targets:', err);
        if (isMounted) setError('Failed to load initial data.');
      }
    };
    fetchInit();
    return () => { isMounted = false; };
  }, []);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanId = targetId.trim();
    if (!cleanId) {
      setError('Slack Channel ID cannot be empty.');
      return;
    }

    if (!isValidSlackChannelId(cleanId)) {
      setError("Invalid format. Must be uppercase letters/numbers starting with 'C' (e.g., C0AQMS8J0P3).");
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/integrations/targets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ platform: 'slack', target_id: cleanId }),
      });
      if (!res.ok) throw new Error('Failed to add target');
      setTargetId('');
      fetchTargets();
    } catch (err) {
      console.error('Failed to add target:', err);
      setError('Failed to add Slack channel. Please check the ID and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    setError(null);
    try {
      const res = await fetch(
        `/api/integrations/targets?id=${id}`,
        { method: 'DELETE' }
      );
      if (!res.ok) throw new Error('Failed to delete target');
      fetchTargets();
    } catch (err) {
      console.error('Failed to delete target:', err);
      setError('Failed to delete Slack channel.');
    }
  };

  return (
    <div className="min-h-screen bg-[#0B0F19] text-slate-200 p-8 flex justify-center">
      <div className="w-full max-w-3xl flex flex-col gap-8">

        {/* Header */}
        <div className="flex flex-col gap-2 border-b border-slate-800 pb-6">
          <h1 className="text-3xl font-bold text-white tracking-tight">Slack Configuration</h1>
          <p className="text-slate-400">
            Manage your Slack Listener Channels to allow the AI to sync messages and interactions.
          </p>
        </div>

        {/* Error Banner */}
        {error && (
          <div
            className="p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm flex items-center gap-3 animate-in fade-in duration-300"
            role="alert"
            aria-live="assertive"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            {error}
          </div>
        )}

        {/* Add Channel Form */}
        <form
          onSubmit={handleAdd}
          className="flex flex-col gap-4 p-6 rounded-2xl bg-slate-800/30 border border-slate-700/60 transition-all hover:border-slate-600/60"
        >
          <div className="flex flex-col gap-2">
            <label htmlFor="targetId" className="text-sm font-semibold text-slate-300">
              Slack Channel ID
            </label>
            <div className="flex flex-col sm:flex-row gap-3">
              <input
                id="targetId"
                type="text"
                value={targetId}
                onChange={(e) => {
                  setTargetId(e.target.value.toUpperCase());
                  if (error) setError(null);
                }}
                placeholder="e.g. C0AQMS8J0P3"
                className={`
                  flex-1 h-12 bg-slate-900 border text-slate-200 rounded-xl px-4
                  focus:outline-none transition-all placeholder:text-slate-500
                  ${error
                    ? 'border-rose-500/50 focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20'
                    : 'border-slate-700 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500'
                  }
                `}
                aria-label="Slack Channel ID"
                aria-invalid={!!error}
              />
              <button
                type="submit"
                disabled={isLoading || !targetId.trim()}
                className="h-12 px-6 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-all disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] whitespace-nowrap flex items-center justify-center min-w-[140px]"
              >
                {isLoading ? (
                  <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : (
                  'Add Channel'
                )}
              </button>
            </div>
          </div>
          <p className="text-xs text-slate-500">
            Find the channel ID in Slack by right-clicking the channel name and selecting &quot;Copy link&quot;, then extracting the alphanumeric ID at the end.
          </p>
        </form>

        {/* Existing Channels List */}
        <div className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold text-slate-200">Active Slack Channels</h2>

          {targets.length === 0 ? (
            <div className="p-8 rounded-2xl border border-dashed border-slate-700 flex flex-col items-center justify-center text-center gap-3">
              <div className="w-12 h-12 rounded-full bg-slate-800 flex items-center justify-center text-slate-400">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
              </div>
              <p className="text-slate-400 text-sm">No Slack channels configured yet.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3">
              {targets.map((t) => (
                <div
                  key={t.id}
                  className="group flex items-center justify-between p-4 rounded-xl bg-slate-800/40 border border-slate-700/50 hover:bg-slate-800/60 hover:border-slate-600 transition-all animate-in slide-in-from-bottom-2 duration-300"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400 shrink-0">
                      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 9h-4V5a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-4h4z"/><path d="M18 9h4"/></svg>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-sm font-semibold text-slate-200 tracking-wide">
                        {t.target_id}
                      </span>
                      <span className="text-xs text-slate-500">
                        Added {new Date(t.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDelete(t.id)}
                    className="p-2 rounded-lg text-slate-500 hover:bg-rose-500/10 hover:text-rose-400 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
                    aria-label={`Delete channel ${t.target_id}`}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

      </div>
    </div>
  );
}
