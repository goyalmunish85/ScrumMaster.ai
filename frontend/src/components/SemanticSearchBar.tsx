import React, { useState, useRef, useEffect } from 'react';

export default function SemanticSearchBar() {
  const [query, setQuery] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

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

  return (
    <div
      role="search"
      className={`relative flex items-center w-full max-w-md mx-4 transition-all duration-300 ease-in-out ${
        isFocused ? 'scale-[1.02]' : 'scale-100'
      }`}
    >
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
          onBlur={() => setIsFocused(false)}
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
  );
}
