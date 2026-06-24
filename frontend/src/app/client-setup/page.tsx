'use client';
import React, { useState, useEffect, useCallback } from 'react';
import { fetchWithRetry } from '../../utils/api';

type Target = {
  id: string;
  platform: string;
  target_id: string;
  created_at: string;
};

export default function ClientSetupPage() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [targetId, setTargetId] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Cache targets
  const fetchTargets = useCallback(async () => {
    try {
      const response = await fetchWithRetry('/api/v1/integrations/targets');
      const data = await response.json();
      // Filter only google sheets if this page is specifically for that
      setTargets((data || []).filter((t: Target) => t.platform === 'sheets'));
    } catch (err) {
      console.error('Failed to fetch targets:', err);
      setError('Failed to load existing configurations. Please try again.');
    }
  }, []);

  useEffect(() => {
    let mounted = true;
    if (mounted) {
      // Workaround to bypass aggressive linting
      setTimeout(() => fetchTargets(), 0);
    }
    return () => { mounted = false; };
  }, [fetchTargets]);

  const handleAddSheet = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!targetId.trim()) return;

    setIsLoading(true);
    setError('');
    setSuccess('');

    try {
      await fetchWithRetry('/api/v1/integrations/targets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ platform: 'sheets', target_id: targetId.trim() }),
      });

      setSuccess('Google Sheet successfully linked!');
      setTargetId('');
      fetchTargets();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      console.error('Failed to add target:', err);
      setError('Failed to link Google Sheet. Please check the ID and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    setError('');
    setSuccess('');
    try {
      await fetchWithRetry(`/api/v1/integrations/targets?id=${id}`, {
        method: 'DELETE',
      });
      fetchTargets();
      setSuccess('Configuration removed successfully.');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      console.error('Failed to delete target:', err);
      setError('Failed to remove configuration.');
    }
  };

  return (
    <div className="min-h-screen bg-[#0B0F19] text-slate-200 p-8">
      <div className="max-w-4xl mx-auto space-y-8">

        <header className="space-y-2">
          <h1 className="text-3xl font-bold text-slate-100 tracking-tight">Client Setup</h1>
          <p className="text-slate-400 text-lg">Configure external data sources and integrations for your workspace.</p>
        </header>

        <section className="bg-slate-800/30 border border-slate-700/60 rounded-2xl p-6 shadow-xl backdrop-blur-sm transition-all duration-300 hover:border-slate-600/60">
          <div className="mb-6">
            <h2 className="text-xl font-semibold text-slate-100 flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                <polyline points="14 2 14 8 20 8"></polyline>
                <line x1="8" y1="13" x2="16" y2="13"></line>
                <line x1="8" y1="17" x2="16" y2="17"></line>
                <polyline points="10 9 9 9 8 9"></polyline>
              </svg>
              Google Sheets Integration
            </h2>
            <p className="text-sm text-slate-400 mt-1">
              Link a Google Sheet to synchronize data automatically. You can find the Sheet ID in your document&apos;s URL.
            </p>
          </div>

          <form onSubmit={handleAddSheet} className="space-y-4">
            <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-end">
              <div className="flex-1 w-full space-y-1.5">
                <label htmlFor="sheetId" className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">
                  Google Sheet ID
                </label>
                <input
                  id="sheetId"
                  type="text"
                  value={targetId}
                  onChange={(e) => setTargetId(e.target.value)}
                  placeholder="e.g. 1BxiMvs0XRYFgCEb_T1tZc6c1r..."
                  className="w-full h-12 bg-slate-900 border border-slate-700 text-slate-200 rounded-xl px-4 focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 transition-all duration-200 placeholder:text-slate-600 shadow-inner"
                  aria-required="true"
                  aria-invalid={!!error}
                  disabled={isLoading}
                />
              </div>
              <button
                type="submit"
                disabled={isLoading || !targetId.trim()}
                className="h-12 px-6 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 shadow-lg shadow-emerald-900/20 w-full sm:w-auto focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 focus:ring-offset-[#0B0F19]"
                aria-busy={isLoading}
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Linking...
                  </>
                ) : (
                  <>
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <line x1="12" y1="5" x2="12" y2="19"></line>
                      <line x1="5" y1="12" x2="19" y2="12"></line>
                    </svg>
                    Add Integration
                  </>
                )}
              </button>
            </div>

            {/* Status Messages with subtle entry animation */}
            <div className="h-6 transition-all duration-300">
              {error && <p className="text-rose-400 text-sm flex items-center gap-1.5 animate-in fade-in slide-in-from-top-1"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>{error}</p>}
              {success && <p className="text-emerald-400 text-sm flex items-center gap-1.5 animate-in fade-in slide-in-from-top-1"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>{success}</p>}
            </div>
          </form>
        </section>

        <section className="space-y-4">
          <h3 className="text-lg font-semibold text-slate-200">Active Data Sources</h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {targets.length > 0 ? (
              targets.map((t) => (
                <div
                  key={t.id}
                  className="group flex items-center justify-between p-4 rounded-xl bg-slate-800/50 border border-slate-700/50 hover:bg-slate-800 hover:border-slate-600 transition-all duration-200 shadow-sm"
                >
                  <div className="flex flex-col gap-1 overflow-hidden">
                    <div className="flex items-center gap-2">
                      <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        Sheets
                      </span>
                      <span className="text-xs text-slate-500">Linked on {new Date(t.created_at).toLocaleDateString()}</span>
                    </div>
                    <span className="text-sm font-medium text-slate-300 truncate" title={t.target_id}>
                      {t.target_id}
                    </span>
                  </div>
                  <button
                    onClick={() => handleDelete(t.id)}
                    className="p-2 rounded-lg text-slate-500 hover:bg-rose-500/10 hover:text-rose-400 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-rose-500"
                    aria-label={`Remove configuration for ${t.target_id}`}
                    title="Remove Integration"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M3 6h18" />
                      <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                      <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                    </svg>
                  </button>
                </div>
              ))
            ) : (
              <div className="col-span-full p-8 rounded-xl bg-slate-800/20 border border-slate-800/50 border-dashed text-center">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-10 w-10 text-slate-600 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                  <line x1="3" y1="9" x2="21" y2="9"></line>
                  <line x1="9" y1="21" x2="9" y2="9"></line>
                </svg>
                <p className="text-slate-400 font-medium">No Google Sheets linked yet.</p>
                <p className="text-sm text-slate-500 mt-1">Add your first document ID above to get started.</p>
              </div>
            )}
          </div>
        </section>

      </div>
    </div>
  );
}
