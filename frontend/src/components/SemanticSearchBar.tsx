import React, { useState, useRef, useEffect } from 'react';

type Task = {
  id: string;
  title: string;
  status: string;
  priority: string;
  labels: string;
  project: string;
  jira_key: string;
  source_name: string;
};

export default function SemanticSearchBar() {
  const [query, setQuery] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const [results, setResults] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Handle click outside to close dropdown
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsFocused(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Handle keyboard shortcut (⌘K or Ctrl+K) to focus the input
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    if (!query.trim()) {
      // Use setTimeout to avoid synchronous setState inside useEffect to fix linting errors
      const timer = setTimeout(() => {
        setResults([]);
      }, 0);
      return () => clearTimeout(timer);
    }

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    debounceTimerRef.current = setTimeout(async () => {
      setIsLoading(true);
      setError(null);
      try {
        const res = await fetch(`/api/search?query=${encodeURIComponent(query)}`);
        if (!res.ok) {
          throw new Error('Failed to fetch search results');
        }
        const data = await res.json();
        setResults(data || []);
      } catch (err: unknown) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError('An unknown error occurred');
        }
      } finally {
        setIsLoading(false);
      }
    }, 400);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [query]);

  return (
    <div
      ref={containerRef}
      role="search"
      className={`relative flex flex-col w-full max-w-md mx-4 transition-all duration-300 ease-in-out ${
        isFocused ? 'scale-[1.02] z-50' : 'scale-100 z-10'
      }`}
    >
      <div className="relative flex items-center w-full">
      <div
        className={`absolute inset-0 rounded-full transition-opacity duration-300 pointer-events-none ${
          isFocused ? 'opacity-100 bg-indigo-500/10 blur-md' : 'opacity-0'
        }`}
      ></div>

      <div
        className={`relative flex items-center w-full bg-slate-900/80 backdrop-blur-sm border rounded-full overflow-hidden transition-colors duration-300 ${
          isFocused
            ? 'border-indigo-500/50 shadow-[0_0_15px_rgba(99,102,241,0.2)]'
            : 'border-slate-700/60 hover:border-slate-600/80'
        }`}
      >
        <div className="pl-4 pr-2 flex items-center justify-center text-slate-400">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={`transition-colors duration-300 ${isFocused ? 'text-indigo-400' : 'text-slate-500'}`}
          >
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.3-4.3"></path>
          </svg>
        </div>

        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsFocused(true)}
          placeholder="Ask anything or search..."
          aria-label="Search or ask anything"
          className="w-full py-2 bg-transparent text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-0"
        />

        {query && (
          <button
            type="button"
            onClick={() => {
              setQuery('');
              inputRef.current?.focus();
            }}
            aria-label="Clear search"
            className="pr-3 pl-2 flex items-center justify-center text-slate-400 hover:text-slate-200 transition-colors"
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
              <path d="M18 6 6 18"></path>
              <path d="m6 6 12 12"></path>
            </svg>
          </button>
        )}

        {!query && (
          <div className="pr-4 pl-2 flex items-center pointer-events-none">
            <kbd className="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 rounded border border-slate-700/80 bg-slate-800 text-[10px] font-medium text-slate-400">
              <span className="text-xs">⌘</span>K
            </kbd>
          </div>
        )}
      </div>
      </div>

      {isFocused && query.trim() && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-slate-900/95 backdrop-blur-xl border border-slate-700/80 rounded-xl shadow-2xl overflow-hidden z-50 transition-all duration-300 animate-in fade-in slide-in-from-top-2">
          {isLoading ? (
            <div className="p-4 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
              <svg className="animate-spin h-4 w-4 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Searching our semantic database...
            </div>
          ) : error ? (
            <div className="p-4 text-center text-red-400 text-sm">
              {error}
            </div>
          ) : results.length > 0 ? (
            <ul className="max-h-80 overflow-y-auto divide-y divide-slate-800/60 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
              {results.map((task) => (
                <li key={task.id} className="p-3 hover:bg-slate-800/60 cursor-pointer transition-colors group">
                  <div className="flex items-start justify-between">
                    <div className="flex flex-col gap-1">
                      <span className="text-sm font-medium text-slate-200 group-hover:text-indigo-300 transition-colors line-clamp-1">
                        {task.title}
                      </span>
                      <div className="flex items-center gap-2 text-xs text-slate-500">
                        {task.jira_key && (
                          <span className="bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded border border-blue-500/20">
                            {task.jira_key}
                          </span>
                        )}
                        <span className="capitalize">{task.status.replace('_', ' ')}</span>
                      </div>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <div className="p-4 text-center text-slate-400 text-sm">
              No results found for &quot;{query}&quot;
            </div>
          )}
        </div>
      )}
    </div>
  );
}
