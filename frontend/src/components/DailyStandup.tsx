import React from 'react';
import ReactMarkdown from 'react-markdown';
import useDailyBriefing from '../hooks/useDailyBriefing';

type DailyStandupProps = {
  onClose: () => void;
};

export default function DailyStandup({ onClose }: DailyStandupProps) {
  const { briefing, isLoading, error } = useDailyBriefing();

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-slate-900 border border-slate-800 w-full max-w-2xl max-h-[80vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-orange-500/20 flex items-center justify-center border border-orange-500/30">
              <svg
                className="w-5 h-5 text-orange-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-slate-100">
              Daily Briefing
            </h2>
          </div>
          <button
            onClick={onClose}
            className="h-10 w-10 rounded-xl flex items-center justify-center hover:bg-slate-800 text-slate-500 transition-colors"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-8 custom-scrollbar bg-[#0B0F19]">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center h-64 space-y-4">
              <div className="relative h-16 w-16">
                <div className="absolute inset-0 rounded-full border-4 border-orange-500/20"></div>
                <div className="absolute inset-0 rounded-full border-4 border-t-orange-500 animate-spin"></div>
              </div>
              <div className="flex flex-col items-center">
                <p className="text-slate-200 font-semibold text-lg">
                  Fetching latest Daily Briefing...
                </p>
              </div>
            </div>
          ) : error ? (
            <div className="flex flex-col items-center justify-center h-64 space-y-4">
              <p className="text-rose-400 text-center">{error}</p>
            </div>
          ) : briefing ? (
            <div className="prose prose-invert max-w-none prose-p:leading-relaxed prose-headings:text-slate-100 prose-headings:font-bold prose-p:text-slate-300">
              <ReactMarkdown
                components={{
                  p: ({ node, ...props }) => (
                    <p className="mb-4 last:mb-0" {...props} />
                  ),
                  ul: ({ node, ...props }) => (
                    <ul
                      className="list-disc pl-6 mb-6 space-y-2 marker:text-orange-500/60"
                      {...props}
                    />
                  ),
                  h1: ({ node, ...props }) => (
                    <h1 className="text-2xl mb-6 text-white" {...props} />
                  ),
                  h2: ({ node, ...props }) => (
                    <h2 className="text-xl mb-4 text-white" {...props} />
                  ),
                  h3: ({ node, ...props }) => (
                    <h3 className="text-lg mb-3 text-white" {...props} />
                  ),
                  strong: ({ node, ...props }) => (
                    <strong className="font-bold text-white" {...props} />
                  ),
                }}
              >
                {briefing.content}
              </ReactMarkdown>
            </div>
          ) : (
            <p className="text-slate-500 text-center">
              No daily briefing found.
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
