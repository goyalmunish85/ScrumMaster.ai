import { NextResponse } from 'next/server';

export async function GET() {
  const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  const credentials = process.env.ADMIN_CREDENTIALS || 'admin:admin';

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 8000); // Strict 8s timeout

  try {
    const res = await fetch(`${backendUrl}/api/v1/tasks`, {
      method: 'GET',
      headers: {
        'Authorization': `Basic ${Buffer.from(credentials).toString('base64')}`,
        'Content-Type': 'application/json',
      },
      next: { revalidate: 60 }, // Aggressively cache static or slow-moving data
      signal: controller.signal,
    });

    clearTimeout(timeoutId);

    if (!res.ok) {
      console.error(`Backend returned ${res.status}: ${res.statusText}`);
      return NextResponse.json({ error: 'Failed to fetch tasks from backend' }, { status: res.status });
    }

    const data = await res.json();

    // Optimize payload size (omit unnecessary fields if any, though frontend expects most of them).
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const optimizedTasks = (data || []).map((task: any) => ({
      id: task.id,
      title: task.title,
      status: task.status,
      assignee: task.assignee,
      due_date: task.due_date,
      updated_at: task.updated_at,
      client: task.client,
      team: task.team,
      task_type: task.task_type,
      sprint: task.sprint,
      source_name: task.source_name,
    }));

    return NextResponse.json(optimizedTasks);

  } catch (err: unknown) {
    clearTimeout(timeoutId);
    if (err instanceof Error) {
      console.error('Error fetching tasks via proxy:', err.message);
      if (err.name === 'AbortError') {
        return NextResponse.json({ error: 'Request to backend timed out' }, { status: 504 });
      }
    } else {
        console.error('Error fetching tasks via proxy:', err);
    }
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}
