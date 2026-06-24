'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';

type Target = {
  id: string;
  platform: string;
  target_id: string;
  created_at: string;
};

export default function JiraConfigPage() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [newProjectKey, setNewProjectKey] = useState('');

  const fetchTargets = async () => {
    try {
      setIsLoading(true);
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const res = await fetch(`${apiUrl}/api/v1/integrations/targets`);
      const data = await res.json();

      // Filter for Jira platform only
      if (data && Array.isArray(data)) {
        setTargets(data.filter((t: Target) => t.platform === 'jira'));
      } else {
        setTargets([]);
      }
    } catch (err) {
      console.error('Failed to fetch Jira targets:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;
    const fetchInit = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${apiUrl}/api/v1/integrations/targets`);
        const data = await res.json();

        if (isMounted) {
          if (data && Array.isArray(data)) {
            setTargets(data.filter((t: Target) => t.platform === 'jira'));
          } else {
            setTargets([]);
          }
        }
      } catch (err) {
        console.error('Failed to fetch Jira targets:', err);
      } finally {
        if (isMounted) setIsLoading(false);
      }
    };
    fetchInit();
    return () => { isMounted = false; };
  }, []);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const keyToSubmit = newProjectKey.trim().toUpperCase();
    if (!keyToSubmit) return;

    setIsSubmitting(true);
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      await fetch(`${apiUrl}/api/v1/integrations/targets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ platform: 'jira', target_id: keyToSubmit }),
      });
      setNewProjectKey('');
      await fetchTargets();
    } catch (err) {
      console.error('Failed to add Jira target:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      await fetch(`${apiUrl}/api/v1/integrations/targets?id=${id}`, {
        method: 'DELETE',
      });
      // Optimistic UI update
      setTargets((prev) => prev.filter((t) => t.id !== id));
    } catch (err) {
      console.error('Failed to delete target:', err);
    }
  };

  return (
    <div className="min-h-screen bg-[#06080D] text-slate-200 p-8 flex flex-col items-center">
      <div className="w-full max-w-4xl flex flex-col gap-8">
        {/* Header */}
        <header className="flex items-center justify-between border-b border-slate-800 pb-6">
          <div className="flex items-center gap-4">
            <Link
              href="/"
              className="p-2 rounded-xl hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500"
              aria-label="Back to Dashboard"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="m15 18-6-6 6-6"/>
              </svg>
            </Link>
            <div>
              <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-indigo-400">
                Jira Configuration
              </h1>
              <p className="text-slate-400 text-sm mt-1">
                Manage your Atlassian Jira project integrations for AI-OS.
              </p>
            </div>
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex flex-col gap-8">
          {/* Add Target Form */}
          <section className="bg-slate-900/50 border border-slate-800 rounded-2xl p-6 transition-all duration-300 hover:border-slate-700">
            <h2 className="text-xl font-semibold mb-4 text-slate-200">Add New Project</h2>
            <form onSubmit={handleAdd} className="flex flex-wrap items-end gap-4">
              <div className="flex flex-col gap-2 flex-1 min-w-[200px]">
                <label htmlFor="projectKey" className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  Jira Project Key
                </label>
                <input
                  id="projectKey"
                  type="text"
                  value={newProjectKey}
                  onChange={(e) => setNewProjectKey(e.target.value)}
                  placeholder="e.g. ENG, OPS, PROD"
                  className="h-12 bg-slate-950 border border-slate-700 text-slate-200 rounded-xl px-4 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors uppercase placeholder:normal-case placeholder:text-slate-600"
                  aria-required="true"
                />
              </div>
              <button
                type="submit"
                disabled={isSubmitting || !newProjectKey.trim()}
                className="h-12 px-6 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 group shadow-lg shadow-blue-900/20"
                aria-label="Add Jira Project Key"
              >
                {isSubmitting ? (
                  <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : (
                  <>
                    <svg className="w-5 h-5 group-hover:scale-110 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    <span>Add Project</span>
                  </>
                )}
              </button>
            </form>
          </section>

          {/* Targets List */}
          <section className="bg-slate-900/50 border border-slate-800 rounded-2xl p-6 transition-all duration-300 hover:border-slate-700">
            <h2 className="text-xl font-semibold mb-6 text-slate-200">Active Jira Projects</h2>

            {isLoading ? (
              <div className="flex justify-center items-center py-12">
                <svg className="animate-spin h-8 w-8 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
            ) : targets.length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                {targets.map((t) => (
                  <div
                    key={t.id}
                    className="group flex items-center justify-between p-4 rounded-xl bg-slate-800/40 border border-slate-700/50 hover:bg-slate-800/60 hover:border-blue-500/30 transition-all duration-200"
                  >
                    <div className="flex items-center gap-3">
                      <span className="px-2.5 py-1 rounded text-xs font-bold tracking-wider bg-blue-500/10 text-blue-400 border border-blue-500/20">
                        {t.platform.toUpperCase()}
                      </span>
                      <span className="text-base font-semibold text-slate-200">
                        {t.target_id}
                      </span>
                    </div>
                    <button
                      onClick={() => handleDelete(t.id)}
                      className="p-2 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-rose-500/20 text-slate-500 hover:text-rose-400 transition-all duration-200 focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-rose-500"
                      aria-label={`Delete ${t.target_id} project`}
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M3 6h18" />
                        <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                        <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                      </svg>
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-12 px-4 rounded-xl bg-slate-800/20 border border-slate-800 border-dashed">
                <svg className="mx-auto h-12 w-12 text-slate-600 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
                <p className="text-slate-400">No Jira projects configured yet.</p>
                <p className="text-sm text-slate-500 mt-1">Add a project key above to get started.</p>
              </div>
            )}
          </section>
        </main>
      </div>
    </div>
  );
}
