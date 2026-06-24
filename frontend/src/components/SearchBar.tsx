'use client';

import React, { useState } from 'react';

interface SearchBarProps {
  onSearch: (query: string) => void;
  placeholder?: string;
}

export default function SearchBar({
  onSearch,
  placeholder = 'Search tasks, events, and projects...',
}: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [isFocused, setIsFocused] = useState(false);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(query);
  };

  return (
    <form
      onSubmit={handleSearch}
      className={`relative w-full flex items-center transition-all duration-300 ease-in-out ${
        isFocused ? 'scale-[1.01]' : 'scale-100'
      }`}
    >
      <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
        <svg
          className={`w-5 h-5 transition-colors duration-300 ${
            isFocused ? 'text-indigo-400' : 'text-slate-500'
          }`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>
      <input
        type="search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        className={`w-full py-3.5 pl-11 pr-4 bg-slate-900/50 border rounded-2xl text-sm text-slate-200 placeholder-slate-500 focus:outline-none transition-all duration-300 ${
          isFocused
            ? 'border-indigo-500/50 shadow-[0_0_15px_rgba(99,102,241,0.2)] bg-slate-900/80'
            : 'border-slate-800/80 hover:border-slate-700/80 hover:bg-slate-900/60'
        }`}
        placeholder={placeholder}
        aria-label="Search"
      />
      <button
        type="submit"
        className="absolute inset-y-0 right-2 flex items-center px-3 my-2 text-xs font-semibold text-white bg-indigo-600 rounded-xl hover:bg-indigo-500 transition-colors duration-200 shadow-sm shadow-indigo-900/20"
        aria-label="Submit Search"
      >
        Search
      </button>
    </form>
  );
}
