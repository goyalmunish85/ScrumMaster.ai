'use client';

import React, { useState, useEffect } from 'react';
import SearchBar from '../../components/SearchBar';
import TaskTable, { Task } from '../../components/TaskTable';
import Timeline, { TimelineEvent } from '../../components/Timeline';
import DailyBriefing from '../../components/DailyBriefing';

export default function Dashboard() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    // Fetch initial data
    const fetchDashboardData = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${apiUrl}/api/v1/tasks`);
        const data = await res.json();
        setTasks(data || []);
      } catch (err) {
        console.error('Failed to fetch tasks for dashboard:', err);
      }

      // Mocking some events since there's no dedicated endpoint for events in the brief
      // but the UI requires a Timeline.
      setEvents([
        {
          id: '1',
          title: 'Project Initiated',
          description: 'The core platform integration project was created.',
          timestamp: new Date(Date.now() - 86400000 * 2).toISOString(),
          type: 'creation',
        },
        {
          id: '2',
          title: 'Jira Sync Completed',
          description: 'Successfully pulled 24 new issues from Jira.',
          timestamp: new Date(Date.now() - 86400000).toISOString(),
          type: 'update',
        },
        {
          id: '3',
          title: 'User Feedback Added',
          description: 'Comment left on task T-12: Needs more tests.',
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          type: 'comment',
        },
        {
          id: '4',
          title: 'Deployment Successful',
          description: 'v1.2.0 deployed to production.',
          timestamp: new Date().toISOString(),
          type: 'completion',
        }
      ]);
    };

    fetchDashboardData();
  }, []);

  const handleSearch = (query: string) => {
    setSearchQuery(query.toLowerCase());
  };

  const filteredTasks = tasks.filter((task) =>
    task.title.toLowerCase().includes(searchQuery) ||
    (task.assignee && task.assignee.toLowerCase().includes(searchQuery))
  );

  return (
    <div className="flex h-screen overflow-hidden bg-slate-950 text-slate-200 font-sans selection:bg-indigo-500/30">
      <main className="flex-1 flex flex-col relative bg-[#0B0F19] overflow-y-auto custom-scrollbar">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[500px] bg-indigo-500/10 blur-[150px] rounded-full pointer-events-none z-0"></div>

        <header className="h-20 border-b border-slate-800/40 bg-[#0B0F19]/80 backdrop-blur-md flex items-center px-8 justify-between relative z-10 sticky top-0">
          <div className="flex flex-col">
            <h1 className="font-semibold text-slate-200 text-xl flex items-center gap-2">
              <svg className="w-6 h-6 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
              Operations Dashboard
            </h1>
            <p className="text-sm text-slate-500 mt-1">Holistic view of all active tasks and events</p>
          </div>
          <div className="w-1/3 min-w-[300px]">
            <SearchBar onSearch={handleSearch} />
          </div>
        </header>

        <div className="flex-1 p-8 relative z-10">
          <div className="max-w-7xl mx-auto h-full flex flex-col gap-8">
            {/* Daily Briefing Row */}
            <div className="w-full">
              <DailyBriefing />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 flex-1">
              <div className="lg:col-span-2 flex flex-col gap-8 h-full">
                <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 shadow-lg shadow-black/20">
                  <h2 className="text-lg font-bold text-slate-100 mb-6 flex items-center gap-2">
                    <svg className="w-5 h-5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                    </svg>
                    Active Tasks
                  </h2>
                  <TaskTable tasks={filteredTasks} />
                </div>
              </div>

              <div className="lg:col-span-1 h-full min-h-[500px]">
                <Timeline events={events} />
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
