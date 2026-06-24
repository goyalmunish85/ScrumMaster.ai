import { NextResponse } from 'next/server';

export async function GET() {
  // Use a secure server-side env variable.
  const basicAuth = process.env.ADMIN_CREDENTIALS;
  if (!basicAuth) {
    return NextResponse.json({ error: 'Server misconfiguration: credentials missing' }, { status: 500 });
  }

  // Define backend host from env or default to localhost
  const backendHost = process.env.BACKEND_HOST || 'http://localhost:8080';

  try {
    const res = await fetch(`${backendHost}/api/activities`, {
      headers: {
        'Authorization': `Basic ${Buffer.from(basicAuth).toString('base64')}`
      },
      // Avoid caching to always get fresh data
      cache: 'no-store'
    });

    if (!res.ok) {
      return NextResponse.json({ error: 'Failed to fetch activities' }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}
