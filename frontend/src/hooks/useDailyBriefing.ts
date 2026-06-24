import { useState, useEffect } from 'react';

type Message = {
  id: string;
  content: string;
  sender_id: string;
  role: 'user' | 'ai';
  created_at: string;
};

export default function useDailyBriefing() {
  const [briefing, setBriefing] = useState<Message | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBriefing = async () => {
      try {
        setIsLoading(true);
        // Using a relative URL so it works in production
        const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${API_BASE_URL}/api/v1/chat/messages`);
        if (!res.ok) {
          throw new Error('Failed to fetch messages');
        }
        const messages: Message[] = await res.json();

        // Find the latest message that is an AI role and contains "Daily Briefing"
        let foundBriefing = null;
        for (let i = messages.length - 1; i >= 0; i--) {
          const msg = messages[i];
          if (msg.role === 'ai' && msg.content.includes('Daily Briefing')) {
            // Unescape any double-escaped newlines returned from the backend
            msg.content = msg.content.replace(/\\n/g, '\n');
            foundBriefing = msg;
            break;
          }
        }

        setBriefing(foundBriefing);
      } catch (err: unknown) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError('An unknown error occurred');
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchBriefing();
  }, []);

  return { briefing, isLoading, error };
}
