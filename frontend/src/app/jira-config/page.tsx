'use client';

import React, { useState } from 'react';

// Defensive programming: strict typing for submission payload
type JiraSubmissionPayload = {
  platform: 'jira';
  target_id: string; // The Jira Project Key
};

export default function JiraConfigPage() {
  const [projectKey, setProjectKey] = useState<string>('');
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState<string>('');

  // Strict client-side validation for Jira Project Key (e.g., SAAS, PROJ, etc.)
  // Format: Starts with a letter, followed by 1 or more letters or numbers.
  const isValidJiraKey = (key: string): boolean => {
    return /^[A-Z][A-Z0-9]+$/.test(key);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage('');
    setStatus('idle');

    // Strip whitespace unconditionally as a defensive measure
    const cleanKey = projectKey.trim();

    // Validate Input Unconditionally
    if (!cleanKey) {
      setErrorMessage('Jira Project Key cannot be empty.');
      setStatus('error');
      return;
    }

    if (!isValidJiraKey(cleanKey)) {
      setErrorMessage('Invalid format. Must be uppercase letters/numbers starting with a letter (e.g., SAAS).');
      setStatus('error');
      return;
    }

    setStatus('loading');

    // Never log PII. Avoid logging user input entirely.
    // console.log("Submitting:", cleanKey); // Omitted for GDPR compliance.

    const payload: JiraSubmissionPayload = {
      platform: 'jira',
      target_id: cleanKey,
    };

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    try {
      const res = await fetch(`${apiUrl}/api/v1/integrations/targets`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        // Defensive programming: Do not expose raw server errors blindly.
        throw new Error('Failed to configure Jira project. Please try again.');
      }

      setStatus('success');
      setProjectKey(''); // Clear form on success

      // Auto-reset success state after 3 seconds for better UX
      setTimeout(() => {
        setStatus('idle');
      }, 3000);

    } catch {
      // Generalize the error message for safety
      setStatus('error');
      setErrorMessage('An unexpected network error occurred.');
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-6 font-sans">
      <div className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-3xl shadow-2xl p-8 relative overflow-hidden transition-all duration-300">

        {/* Micro-animation decorative elements */}
        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-500 to-indigo-600 opacity-80" />
        <div className="absolute -top-12 -right-12 w-32 h-32 bg-blue-500/10 rounded-full blur-2xl" />

        <div className="relative z-10">
          <h1 className="text-2xl font-bold text-slate-100 mb-2">Jira Configuration</h1>
          <p className="text-sm text-slate-400 mb-8">
            Connect your Jira project to enable bi-directional task synchronization and updates.
          </p>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label
                htmlFor="projectKey"
                className="block text-xs font-semibold text-slate-300 uppercase tracking-wider"
              >
                Jira Project Key
              </label>
              <div className="relative group">
                <input
                  id="projectKey"
                  type="text"
                  value={projectKey}
                  onChange={(e) => {
                    // Force uppercase as they type for better UX
                    setProjectKey(e.target.value.toUpperCase());
                    if (status === 'error') setStatus('idle'); // Clear error on typing
                  }}
                  placeholder="e.g. SAAS"
                  disabled={status === 'loading'}
                  className={`
                    w-full h-12 px-4 rounded-xl bg-slate-950 border text-slate-100
                    placeholder:text-slate-600 transition-all duration-200 outline-none
                    ${status === 'error'
                      ? 'border-rose-500/50 focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20'
                      : 'border-slate-800 focus:border-blue-500 focus:ring-1 focus:ring-blue-500/20 hover:border-slate-700'
                    }
                    disabled:opacity-50 disabled:cursor-not-allowed
                  `}
                  aria-invalid={status === 'error'}
                  aria-describedby={status === 'error' ? "error-message" : undefined}
                />
              </div>

              {/* Accessible Error Message */}
              {status === 'error' && (
                <p id="error-message" className="text-sm text-rose-400 animate-in fade-in slide-in-from-top-1">
                  {errorMessage}
                </p>
              )}
            </div>

            <button
              type="submit"
              disabled={status === 'loading'}
              className={`
                w-full h-12 rounded-xl font-semibold text-white transition-all duration-200
                flex items-center justify-center gap-2
                ${status === 'loading'
                  ? 'bg-blue-600/50 cursor-not-allowed'
                  : status === 'success'
                  ? 'bg-emerald-500 hover:bg-emerald-400'
                  : 'bg-blue-600 hover:bg-blue-500 active:scale-[0.98]'
                }
              `}
            >
              {status === 'loading' ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Connecting...
                </>
              ) : status === 'success' ? (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  Connected Successfully
                </>
              ) : (
                'Add Project'
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
