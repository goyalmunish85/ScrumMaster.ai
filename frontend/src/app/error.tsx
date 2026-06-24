'use client'; // Error components must be Client Components

import { useEffect } from 'react';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error(error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-8 text-center bg-[#0B0F19] antialiased">
      <div className="flex flex-col items-center justify-center p-10 bg-slate-900/50 backdrop-blur-xl rounded-3xl border border-red-500/20 shadow-2xl shadow-red-500/10 max-w-lg w-full transition-all duration-500 ease-out animate-in fade-in slide-in-from-bottom-4">
        <div className="w-20 h-20 mb-8 rounded-full bg-gradient-to-br from-red-500/20 to-red-500/5 flex items-center justify-center relative overflow-hidden group">
          <div className="absolute inset-0 bg-red-500/20 animate-pulse"></div>
          <svg
            className="w-10 h-10 text-red-500 relative z-10 group-hover:scale-110 transition-transform duration-300"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>

        <h2 className="text-2xl font-bold text-white mb-4 tracking-tight">
          Oops! Something went wrong
        </h2>

        <p className="text-slate-400 mb-8 leading-relaxed max-w-sm">
          We&apos;ve encountered an unexpected error while loading this page. Our team has been notified.
        </p>

        <div className="flex gap-4 w-full sm:w-auto">
          <button
            onClick={() => window.location.reload()}
            className="flex-1 sm:flex-none px-6 py-3 bg-transparent border border-slate-700 hover:bg-slate-800 text-slate-300 text-sm font-semibold rounded-xl transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-slate-500 focus:ring-offset-2 focus:ring-offset-[#0B0F19]"
          >
            Reload Page
          </button>

          <button
            onClick={() => reset()}
            className="flex-1 sm:flex-none px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold rounded-xl transition-all duration-200 hover:shadow-lg hover:shadow-indigo-500/20 hover:-translate-y-0.5 active:translate-y-0 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-[#0B0F19]"
          >
            Try Again
          </button>
        </div>
      </div>
    </div>
  );
}
