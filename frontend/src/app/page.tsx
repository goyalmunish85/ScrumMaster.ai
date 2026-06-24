'use client';

import React, { useState, useEffect, useRef } from 'react';
import ReactMarkdown from 'react-markdown';
import IntegrationsPanel from '../components/IntegrationsPanel';
import SemanticSearchBar from '../components/SemanticSearchBar';
import MetricsCards from '../components/MetricsCards';

type Message = {
  id: string;
  content: string;
  sender_id: string;
  role: 'user' | 'ai';
  created_at: string;
};

type Task = {
  id: string;
  title: string;
  status: string;
  assignee: string | null;
  due_date: string | null;
  updated_at: string;
  client: string | null;
  team: string | null;
  task_type: string | null;
  sprint: string | null;
  source_name: string | null;
};

type SyncLog = {
  id: string;
  platform: string;
  target_id: string;
  status: string;
  message: string;
  created_at: string;
};

export default function Home() {
  const [input, setInput] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  // Reports
  const [isReportLoading, setIsReportLoading] = useState(false);
  const [weeklyReport, setWeeklyReport] = useState<string | null>(null);
  const [showReportModal, setShowReportModal] = useState(false);

  // Sync Dashboard
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [syncLogs, setSyncLogs] = useState<SyncLog[]>([]);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncPlatform, setSyncPlatform] = useState('');
  const [syncTargetId, setSyncTargetId] = useState('');
  const [syncFullSync, setSyncFullSync] = useState(false);

  const [hasMounted, setHasMounted] = useState(false);
  const [evaluatingMsgId, setEvaluatingMsgId] = useState<string | null>(null);

  const handleExportCSV = () => {
    if (tasks.length === 0) {
      alert('No tasks to export.');
      return;
    }

    // GDPR compliance: Omit 'assignee' and 'client' (PII)
    const header = [
      'Jira Key',
      'Title',
      'Status',
      'Team',
      'Task Type',
      'Sprint',
      'Source Name',
      'Due Date',
      'Updated At',
    ];

    const escapeCSV = (str: string | null | undefined) => {
      if (!str) return '';
      const stringified = String(str);
      if (stringified.includes(',') || stringified.includes('"') || stringified.includes('\n')) {
        return `"${stringified.replace(/"/g, '""')}"`;
      }
      return stringified;
    };

    const rows = tasks.map((t) => [
      escapeCSV(t.id),
      escapeCSV(t.title),
      escapeCSV(t.status),
      escapeCSV(t.team),
      escapeCSV(t.task_type),
      escapeCSV(t.sprint),
      escapeCSV(t.source_name),
      escapeCSV(t.due_date ? t.due_date.substring(0, 10) : ''),
      escapeCSV(new Date(t.updated_at).toLocaleString()),
    ]);

    const csvContent = [header, ...rows]
      .map((row) => row.join(','))
      .join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', 'tasks_export.csv');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };
  const [evalFeedback, setEvalFeedback] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setHasMounted(true);
  }, []);

  const fetchTasks = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/v1/tasks');
      const data = await res.json();
      setTasks(data || []);
    } catch (err) {
      console.error('Failed to fetch tasks:', err);
    }
  };

  const fetchSyncLogs = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/v1/integrations/logs');
      const data = await res.json();
      setSyncLogs(data || []);
    } catch (err) {
      console.error('Failed to fetch sync logs:', err);
    }
  };

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (showSyncModal) {
      fetchSyncLogs(); // fetch immediately
      interval = setInterval(fetchSyncLogs, 2000); // poll every 2 seconds
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [showSyncModal]);

  const handleTargetedSync = async (
    platform?: string,
    target_id?: string,
    full_sync?: boolean
  ) => {
    setIsSyncing(true);
    try {
      await fetch('http://localhost:8080/api/v1/integrations/sync', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          platform: platform || '',
          target_id: target_id || '',
          full_sync: full_sync || false,
        }),
      });
      // The backend immediately returns 202 Accepted.
      // Poll logs to update UI.
      setTimeout(fetchSyncLogs, 1000);
    } catch (err) {
      console.error('Failed to sync:', err);
    } finally {
      setIsSyncing(false);
    }
  };

  const handleGenerateReport = async () => {
    setIsReportLoading(true);
    setShowReportModal(true);
    try {
      const res = await fetch('http://localhost:8080/api/v1/reports/weekly');
      const data = await res.json();
      setWeeklyReport(data.report);
    } catch (err) {
      console.error('Failed to generate report:', err);
      setWeeklyReport('Failed to generate weekly report. Please try again.');
    } finally {
      setIsReportLoading(false);
    }
  };

  const submitEvaluation = async (messageId: string) => {
    if (!evalFeedback.trim()) return;
    try {
      await fetch('http://localhost:8080/api/v1/chat/evaluate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message_id: messageId, feedback: evalFeedback }),
      });
      setEvaluatingMsgId(null);
      setEvalFeedback('');
      alert('Feedback saved! The AI will use this to self-improve.');
    } catch (err) {
      console.error('Failed to submit evaluation:', err);
    }
  };

  // Fetch initial messages & tasks
  useEffect(() => {
    fetch('http://localhost:8080/api/v1/chat/messages')
      .then((res) => res.json())
      .then((data) => setMessages(data || []))
      .catch((err) => console.error('Failed to fetch messages:', err));

    fetchTasks();
  }, []);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!input.trim() || isLoading) return;

    const userText = input.trim();
    setInput('');
    setIsLoading(true);

    // Optimistically add user message
    const tempUserMsg: Message = {
      id: 'temp-' + Date.now(),
      content: userText,
      sender_id: 'user-local',
      role: 'user',
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, tempUserMsg]);

    try {
      const res = await fetch('http://localhost:8080/api/v1/chat/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: userText }),
      });

      if (res.ok) {
        const aiMsg = await res.json();
        setMessages((prev) => [...prev, aiMsg]);
        // Refresh tasks after AI processes the message
        fetchTasks();
      } else {
        setMessages((prev) => [
          ...prev,
          {
            id: 'err-' + Date.now(),
            content: 'Sorry, I had trouble reaching the backend.',
            sender_id: 'ai-system',
            role: 'ai',
            created_at: new Date().toISOString(),
          },
        ]);
      }
    } catch (err) {
      console.error('Failed to send message:', err);
      setMessages((prev) => [
        ...prev,
        {
          id: 'err-' + Date.now(),
          content: 'Network error reaching the backend.',
          sender_id: 'ai-system',
          role: 'ai',
          created_at: new Date().toISOString(),
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  if (!hasMounted) return null;

  return (
    <div className="flex h-screen overflow-hidden bg-slate-950 text-slate-200 font-sans selection:bg-indigo-500/30">
      {/* Sidebar */}
      <aside className="w-72 flex-shrink-0 border-r border-slate-800/60 bg-slate-950/50 backdrop-blur-xl flex flex-col relative z-20">
        <div className="h-16 px-6 border-b border-slate-800/60 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20">
              <span className="text-white font-bold text-sm">AI</span>
            </div>
            <h2 className="font-bold tracking-tight text-slate-100">
              Antigravity OS
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-4 py-6 space-y-8 no-scrollbar">
          <div>
            <h3 className="mb-3 px-2 text-[11px] font-semibold text-slate-500 uppercase tracking-widest">
              Active Projects
            </h3>
            <div className="space-y-1">
              <button className="w-full flex items-center gap-3 rounded-xl bg-indigo-500/10 border border-indigo-500/20 px-3 py-2.5 text-sm font-medium text-indigo-200 transition-all shadow-sm">
                <span className="flex h-2.5 w-2.5 rounded-full bg-indigo-400 shadow-[0_0_8px_rgba(129,140,248,0.8)] animate-pulse"></span>
                Core Platform
              </button>
            </div>
          </div>
        </div>

        <div className="p-4 border-t border-slate-800/60 mt-auto">
          <div className="flex items-center gap-3 px-2">
            <div className="h-9 w-9 rounded-full bg-slate-800 flex items-center justify-center border border-slate-700">
              <span className="text-slate-300 font-medium text-sm">US</span>
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-medium text-slate-200">User</span>
              <span className="text-[11px] text-slate-500">
                Engineering Lead
              </span>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Chat Area */}
      <main className="flex-1 flex flex-col relative bg-[#0B0F19]">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-indigo-500/10 blur-[120px] rounded-full pointer-events-none"></div>

        <header className="h-16 border-b border-slate-800/40 bg-[#0B0F19]/80 backdrop-blur-md flex items-center px-8 justify-between relative z-10">
          <div className="flex flex-col">
            <h1 className="font-semibold text-slate-200 text-base flex items-center gap-2">
              Core Platform
              <span className="text-slate-600">/</span>
              <span className="text-indigo-300">Operations Chat</span>
            </h1>
          </div>

          <SemanticSearchBar />

          <div className="flex items-center gap-2 bg-indigo-500/10 border border-indigo-500/20 px-3 py-1.5 rounded-full shadow-inner">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
            </span>
            <span className="text-xs font-medium text-indigo-300 tracking-wide uppercase">
              AI Online
            </span>
          </div>
        </header>

        {/* Top-Level Metrics Dashboard */}
        <MetricsCards tasks={tasks} />

        {/* Messages List */}
        <div className="flex-1 overflow-y-auto px-4 py-8 space-y-8 relative z-10 scroll-smooth">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex gap-4 max-w-4xl mx-auto group ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}
            >
              <div
                className={`h-10 w-10 rounded-2xl flex items-center justify-center flex-shrink-0 shadow-sm border mt-1 transition-transform group-hover:scale-105 ${msg.role === 'user' ? 'bg-slate-800 border-slate-700' : 'bg-gradient-to-br from-indigo-500 to-purple-600 shadow-indigo-500/20 border-indigo-400/20'}`}
              >
                {msg.role === 'user' ? (
                  <span className="text-slate-300 font-medium text-sm">US</span>
                ) : (
                  <svg
                    className="w-5 h-5 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 10V3L4 14h7v7l9-11h-7z"
                    />
                  </svg>
                )}
              </div>
              <div
                className={`flex flex-col gap-1.5 w-full ${msg.role === 'user' ? 'items-end' : ''}`}
              >
                <div className="flex items-center gap-2">
                  {msg.role === 'user' && (
                    <span className="text-[10px] text-slate-500 font-medium">
                      You
                    </span>
                  )}
                  <span className="font-semibold text-sm text-slate-200">
                    {msg.role === 'user' ? '' : 'Execution OS'}
                  </span>
                  {msg.role !== 'user' && (
                    <span className="text-[10px] text-slate-500 font-medium">
                      {new Date(msg.created_at).toLocaleTimeString([], {
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  )}
                </div>
                <div
                  className={`text-[15px] p-5 rounded-2xl shadow-sm leading-relaxed max-w-[85%] ${msg.role === 'user' ? 'text-indigo-50 bg-indigo-600/20 border border-indigo-500/30 rounded-tr-none' : 'text-slate-200 bg-slate-800/80 backdrop-blur-md rounded-tl-none border border-slate-700/60 shadow-lg'}`}
                >
                  <ReactMarkdown
                    components={{
                      p: ({ node, ...props }) => (
                        <p className="mb-4 last:mb-0" {...props} />
                      ),
                      ul: ({ node, ...props }) => (
                        <ul
                          className="list-disc pl-6 mb-4 space-y-1.5 marker:text-slate-500"
                          {...props}
                        />
                      ),
                      ol: ({ node, ...props }) => (
                        <ol
                          className="list-decimal pl-6 mb-4 space-y-1.5 marker:text-slate-500"
                          {...props}
                        />
                      ),
                      li: ({ node, ...props }) => (
                        <li className="pl-1" {...props} />
                      ),
                      strong: ({ node, ...props }) => (
                        <strong
                          className="font-semibold text-white"
                          {...props}
                        />
                      ),
                    }}
                  >
                    {msg.content}
                  </ReactMarkdown>
                </div>
                {msg.role === 'ai' && (
                  <div className="flex flex-col gap-2 w-full max-w-[85%]">
                    {!evaluatingMsgId || evaluatingMsgId !== msg.id ? (
                      <button
                        onClick={() => setEvaluatingMsgId(msg.id)}
                        className="text-[11px] text-slate-500 hover:text-indigo-400 self-start ml-2 flex items-center gap-1 transition-colors"
                      >
                        <svg
                          className="w-3.5 h-3.5"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                          />
                        </svg>
                        Improve/Fix Response
                      </button>
                    ) : (
                      <div className="flex items-center gap-2 mt-1">
                        <input
                          type="text"
                          autoFocus
                          value={evalFeedback}
                          onChange={(e) => setEvalFeedback(e.target.value)}
                          placeholder="Tell the AI how to improve this..."
                          className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
                        />
                        <button
                          onClick={() => submitEvaluation(msg.id)}
                          className="px-3 py-1.5 bg-indigo-500 hover:bg-indigo-400 text-white text-xs font-medium rounded-lg transition-colors"
                        >
                          Save
                        </button>
                        <button
                          onClick={() => {
                            setEvaluatingMsgId(null);
                            setEvalFeedback('');
                          }}
                          className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-400 text-xs font-medium rounded-lg transition-colors border border-slate-700"
                        >
                          Cancel
                        </button>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          ))}
          {isLoading && (
            <div className="flex gap-4 max-w-4xl mx-auto group">
              <div className="h-10 w-10 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center flex-shrink-0 shadow-lg shadow-indigo-500/20 border border-indigo-400/20 mt-1 animate-pulse">
                <span className="flex gap-1">
                  <span className="h-1 w-1 bg-white rounded-full animate-bounce"></span>
                  <span className="h-1 w-1 bg-white rounded-full animate-bounce delay-75"></span>
                  <span className="h-1 w-1 bg-white rounded-full animate-bounce delay-150"></span>
                </span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <div className="p-6 bg-gradient-to-t from-[#0B0F19] via-[#0B0F19] relative z-20">
          <form
            onSubmit={handleSendMessage}
            className={`max-w-4xl mx-auto relative rounded-2xl border transition-all duration-300 bg-slate-900/50 backdrop-blur-xl shadow-lg ${isFocused ? 'border-indigo-500/50 shadow-indigo-500/10' : 'border-slate-700/60'}`}
          >
            <textarea
              className="w-full bg-transparent p-5 pr-14 text-sm text-slate-100 placeholder:text-slate-500 focus:outline-none resize-none min-h-[72px] max-h-40 font-medium"
              placeholder="Send an operational update or ask a question..."
              rows={1}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              onFocus={() => setIsFocused(true)}
              onBlur={() => setIsFocused(false)}
              disabled={isLoading}
            />
            <div className="absolute right-3 bottom-3 flex items-center gap-2">
              <button
                type="submit"
                disabled={input.length === 0 || isLoading}
                className={`h-10 w-10 rounded-xl flex items-center justify-center transition-all duration-300 shadow-md ${input.length > 0 && !isLoading ? 'bg-indigo-500 text-white hover:bg-indigo-400 hover:shadow-indigo-500/25 shadow-indigo-500/20' : 'bg-slate-800 text-slate-500 cursor-not-allowed'}`}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="m22 2-7 20-4-9-9-4Z" />
                  <path d="M22 2 11 13" />
                </svg>
              </button>
            </div>
          </form>
          <div className="flex items-center justify-between max-w-4xl mx-auto mt-3 px-2">
            <p className="text-[11px] text-slate-500 font-medium flex items-center gap-1.5">
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              AI extracts events and maintains project memory automatically.
            </p>
            <p className="text-[11px] text-slate-500 font-medium tracking-wide">
              Press{' '}
              <kbd className="font-sans px-1.5 py-0.5 bg-slate-800 rounded border border-slate-700 text-slate-400 mx-1">
                Enter
              </kbd>{' '}
              to send
            </p>
          </div>
        </div>
      </main>

      {/* Right Sidebar - Live Tasks Dashboard */}
      <aside className="w-[400px] flex-shrink-0 border-l border-slate-800/60 bg-[#070b14] flex flex-col relative z-20">
        <div className="h-16 px-6 border-b border-slate-800/60 flex items-center justify-between bg-slate-900/40 backdrop-blur-md">
          <h2 className="font-bold tracking-tight text-slate-100 flex items-center gap-2">
            <svg
              className="w-4 h-4 text-indigo-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
              />
            </svg>
            Live Dashboard
          </h2>
          <div className="flex gap-2">
            <button
              onClick={handleExportCSV}
              className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 hover:bg-cyan-500/20 transition-colors shadow-sm group relative"
              title="Export Tasks as CSV"
            >
              <svg
                className="w-4 h-4 group-hover:-translate-y-0.5 transition-transform duration-300"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </button>
            <button
              onClick={() => setShowSyncModal(true)}
              className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20 transition-colors shadow-sm group relative"
              title="Sync Integrations (Jira, Sheets, Slack)"
            >
              <svg
                className="w-4 h-4 group-hover:rotate-180 transition-transform duration-500"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>
            <button
              onClick={handleGenerateReport}
              className="p-2 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 hover:bg-indigo-500/20 transition-colors shadow-sm group relative"
              title="Generate Weekly Report"
            >
              <svg
                className="w-4 h-4 group-hover:rotate-12 transition-transform"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-6 no-scrollbar">
          {['BLOCKED', 'IN_PROGRESS', 'DRAFT', 'DONE'].map((statusGroup) => {
            const groupTasks = tasks.filter((t) => t.status === statusGroup);
            if (groupTasks.length === 0) return null;

            let statusColor =
              'bg-slate-500/10 text-slate-400 border-slate-500/20';
            let dotColor = 'bg-slate-500';
            let label = 'Drafts';

            if (statusGroup === 'BLOCKED') {
              statusColor = 'bg-rose-500/10 text-rose-400 border-rose-500/20';
              dotColor = 'bg-rose-500 animate-pulse';
              label = 'Blocked';
            } else if (statusGroup === 'IN_PROGRESS') {
              statusColor =
                'bg-amber-500/10 text-amber-400 border-amber-500/20';
              dotColor = 'bg-amber-500';
              label = 'In Progress';
            } else if (statusGroup === 'DONE') {
              statusColor =
                'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
              dotColor = 'bg-emerald-500';
              label = 'Completed';
            }

            return (
              <div key={statusGroup} className="space-y-3">
                <div className="flex items-center gap-2 px-1">
                  <span className={`h-2 w-2 rounded-full ${dotColor}`}></span>
                  <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider">
                    {label}
                  </h3>
                  <span className="ml-auto text-xs font-medium text-slate-500">
                    {groupTasks.length}
                  </span>
                </div>

                <div className="space-y-2">
                  {groupTasks.map((task) => (
                    <div
                      key={task.id}
                      className="p-3.5 rounded-xl bg-slate-900/50 border border-slate-800/80 shadow-sm hover:border-slate-700 transition-colors group"
                    >
                      <p className="text-sm font-medium text-slate-200 leading-snug">
                        {task.title}
                      </p>

                      <div className="flex flex-wrap gap-1.5 mt-2.5">
                        {task.source_name && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                            {task.source_name.split(':')[0]}
                          </span>
                        )}
                        {task.client && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-fuchsia-500/10 text-fuchsia-400 border border-fuchsia-500/20">
                            {task.client}
                          </span>
                        )}
                        {task.team && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
                            {task.team}
                          </span>
                        )}
                        {task.sprint && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-violet-500/10 text-violet-400 border border-violet-500/20">
                            {task.sprint}
                          </span>
                        )}
                        {task.task_type && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                            {task.task_type}
                          </span>
                        )}
                      </div>

                      <div className="flex items-center gap-3 mt-3 pt-3 border-t border-slate-800/60">
                        {task.assignee ? (
                          <div className="flex items-center gap-1.5">
                            <div className="h-5 w-5 rounded-full bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
                              <span className="text-[9px] font-bold text-indigo-300">
                                {task.assignee.substring(0, 2).toUpperCase()}
                              </span>
                            </div>
                            <span
                              className="text-xs text-slate-400 truncate max-w-[100px]"
                              title={task.assignee}
                            >
                              {task.assignee}
                            </span>
                          </div>
                        ) : (
                          <span className="text-xs text-slate-500 italic">
                            Unassigned
                          </span>
                        )}

                        {task.due_date && (
                          <div className="flex items-center gap-1 text-xs text-slate-400 ml-auto shrink-0">
                            <svg
                              className="w-3.5 h-3.5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                              />
                            </svg>
                            {task.due_date.substring(0, 10)}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}

          {tasks.length === 0 && (
            <div className="flex flex-col items-center justify-center h-40 text-center px-4">
              <svg
                className="w-8 h-8 text-slate-600 mb-3"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                />
              </svg>
              <p className="text-sm font-medium text-slate-400">
                No active tasks
              </p>
              <p className="text-xs text-slate-500 mt-1">
                Start chatting to extract tasks
              </p>
            </div>
          )}
        </div>
      </aside>

      {/* Weekly Report Modal */}
      {showReportModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-300">
          <div className="bg-slate-900 border border-slate-800 w-full max-w-2xl max-h-[80vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
            <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-xl bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
                  <svg
                    className="w-5 h-5 text-indigo-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </div>
                <h2 className="text-xl font-bold text-slate-100">
                  Weekly Executive Summary
                </h2>
              </div>
              <button
                onClick={() => {
                  setShowReportModal(false);
                  setWeeklyReport(null);
                }}
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
              {isReportLoading ? (
                <div className="flex flex-col items-center justify-center h-64 space-y-4">
                  <div className="relative h-16 w-16">
                    <div className="absolute inset-0 rounded-full border-4 border-indigo-500/20"></div>
                    <div className="absolute inset-0 rounded-full border-4 border-t-indigo-500 animate-spin"></div>
                  </div>
                  <div className="flex flex-col items-center">
                    <p className="text-slate-200 font-semibold text-lg">
                      Synthesizing project memory...
                    </p>
                    <p className="text-slate-500 text-sm">
                      DeepSeek is analyzing your operational velocity
                    </p>
                  </div>
                </div>
              ) : weeklyReport ? (
                <div className="prose prose-invert max-w-none prose-p:leading-relaxed prose-headings:text-slate-100 prose-headings:font-bold prose-p:text-slate-300">
                  <ReactMarkdown
                    components={{
                      p: ({ node, ...props }) => (
                        <p className="mb-4 last:mb-0" {...props} />
                      ),
                      ul: ({ node, ...props }) => (
                        <ul
                          className="list-disc pl-6 mb-6 space-y-2 marker:text-indigo-500/60"
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
                    {weeklyReport}
                  </ReactMarkdown>
                </div>
              ) : (
                <p className="text-slate-500 text-center">
                  No report generated.
                </p>
              )}
            </div>

            <div className="px-6 py-4 border-t border-slate-800 bg-slate-900/50 flex items-center justify-between">
              <p className="text-xs text-slate-500 flex items-center gap-1.5">
                <svg
                  className="w-3.5 h-3.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                This report is generated by analyzing the last 7 days of audit
                logs.
              </p>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => {
                    if (weeklyReport) {
                      navigator.clipboard.writeText(weeklyReport);
                      alert('Report copied to clipboard!');
                    }
                  }}
                  className="px-4 py-2 rounded-xl bg-slate-800 text-slate-200 text-sm font-semibold border border-slate-700 hover:bg-slate-700 transition-colors flex items-center gap-2 shadow-sm"
                >
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                    />
                  </svg>
                  Copy Report
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Sync Dashboard Modal */}
      {showSyncModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-300">
          <div className="bg-slate-900 border border-slate-800 w-full max-w-4xl max-h-[85vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
            <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-xl bg-emerald-500/20 flex items-center justify-center border border-emerald-500/30">
                  <svg
                    className="w-5 h-5 text-emerald-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                </div>
                <h2 className="text-xl font-bold text-slate-100">
                  Sync Dashboard
                </h2>
              </div>
              <button
                onClick={() => setShowSyncModal(false)}
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

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar bg-[#0B0F19] flex flex-col gap-6">
              {/* Dynamic Targets UI */}
              <IntegrationsPanel />

              <div className="flex flex-col gap-4 p-5 rounded-2xl bg-slate-800/30 border border-slate-800/60 mt-4">
                <div className="flex flex-wrap items-end gap-4">
                  <div className="flex flex-col gap-1.5 flex-1 min-w-[200px]">
                    <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Platform
                    </label>
                    <select
                      value={syncPlatform}
                      onChange={(e) => setSyncPlatform(e.target.value)}
                      className="h-10 bg-slate-900 border border-slate-700 text-slate-200 rounded-xl px-3 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
                    >
                      <option value="">All Platforms</option>
                      <option value="jira">Jira</option>
                      <option value="slack">Slack</option>
                      <option value="sheets">Google Sheets</option>
                      <option value="gitlab">GitLab</option>
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5 flex-1 min-w-[200px]">
                    <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Target ID (Optional)
                    </label>
                    <input
                      type="text"
                      value={syncTargetId}
                      onChange={(e) => setSyncTargetId(e.target.value)}
                      placeholder="e.g. SAAS-123 or C0AQMS8J0P3"
                      className="h-10 bg-slate-900 border border-slate-700 text-slate-200 rounded-xl px-3 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors placeholder:text-slate-600"
                    />
                  </div>
                  <div className="flex items-center gap-2 h-10 px-2">
                    <input
                      type="checkbox"
                      id="fullSyncToggle"
                      checked={syncFullSync}
                      onChange={(e) => setSyncFullSync(e.target.checked)}
                      className="w-4 h-4 rounded border-slate-700 text-indigo-500 focus:ring-indigo-500 bg-slate-900 cursor-pointer"
                    />
                    <label
                      htmlFor="fullSyncToggle"
                      className="text-sm text-slate-300 cursor-pointer"
                    >
                      Full History Sync
                    </label>
                  </div>
                </div>

                <div className="flex gap-3 mt-2 border-t border-slate-700/50 pt-4">
                  <button
                    onClick={() => handleTargetedSync('', '', syncFullSync)}
                    disabled={isSyncing}
                    className="px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium transition-colors disabled:opacity-50 flex items-center gap-2 shadow-lg shadow-emerald-900/20"
                  >
                    {isSyncing ? (
                      <svg
                        className="animate-spin h-4 w-4 text-white"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24"
                      >
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                        ></circle>
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                      </svg>
                    ) : (
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                        />
                      </svg>
                    )}
                    Sync Everything
                  </button>
                  <button
                    onClick={() =>
                      handleTargetedSync(
                        syncPlatform,
                        syncTargetId,
                        syncFullSync
                      )
                    }
                    disabled={isSyncing || (!syncPlatform && !syncTargetId)}
                    className="px-5 py-2.5 rounded-xl bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-400 border border-indigo-500/20 font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
                  >
                    Sync Targeted Selection
                  </button>
                </div>
              </div>

              <div className="rounded-2xl border border-slate-800/80 bg-slate-900/40 overflow-hidden">
                <table className="w-full text-left text-sm text-slate-300">
                  <thead className="bg-slate-800/50 text-xs uppercase text-slate-500">
                    <tr>
                      <th className="px-6 py-4 font-semibold tracking-wider">
                        Status
                      </th>
                      <th className="px-6 py-4 font-semibold tracking-wider">
                        Platform
                      </th>
                      <th className="px-6 py-4 font-semibold tracking-wider">
                        Target
                      </th>
                      <th className="px-6 py-4 font-semibold tracking-wider">
                        Message
                      </th>
                      <th className="px-6 py-4 font-semibold tracking-wider">
                        Time
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {syncLogs.length > 0 ? (
                      syncLogs.map((log) => (
                        <tr
                          key={log.id}
                          className="hover:bg-slate-800/20 transition-colors"
                        >
                          <td className="px-6 py-4">
                            {log.status === 'SUCCESS' ? (
                              <span className="px-2 py-1 rounded-md text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                SUCCESS
                              </span>
                            ) : (
                              <span className="px-2 py-1 rounded-md text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">
                                ERROR
                              </span>
                            )}
                          </td>
                          <td className="px-6 py-4 font-medium text-slate-200 capitalize">
                            {log.platform}
                          </td>
                          <td className="px-6 py-4">
                            <span className="px-2 py-1 rounded-md bg-slate-800 text-xs">
                              {log.target_id}
                            </span>
                          </td>
                          <td
                            className="px-6 py-4 max-w-sm truncate"
                            title={log.message}
                          >
                            {log.message}
                          </td>
                          <td className="px-6 py-4 text-xs text-slate-500">
                            {new Date(log.created_at).toLocaleString()}
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td
                          colSpan={5}
                          className="px-6 py-8 text-center text-slate-500 italic"
                        >
                          No sync logs found.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
