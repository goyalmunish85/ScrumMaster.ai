import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const query = searchParams.get('query');
  const limit = searchParams.get('limit') || '10';

  if (!query) {
    return NextResponse.json({ error: 'Missing query parameter' }, { status: 400 });
  }

  // Use environment variables for the backend URL and credentials
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const adminCredentials = process.env.ADMIN_CREDENTIALS;

  if (!adminCredentials) {
    console.error('ADMIN_CREDENTIALS environment variable is not set');
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }

  try {
    const backendResponse = await fetch(`${backendUrl}/api/search?query=${encodeURIComponent(query)}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Authorization': `Basic ${Buffer.from(adminCredentials).toString('base64')}`,
        'Content-Type': 'application/json',
      },
      // Do not cache search queries aggressively, or implement a shorter stale-while-revalidate
      cache: 'no-store',
    });

    if (!backendResponse.ok) {
      const errorText = await backendResponse.text();
      console.error(`Backend search error (${backendResponse.status}):`, errorText);
      return NextResponse.json(
        { error: 'Search failed' },
        { status: backendResponse.status }
      );
    }

    const data = await backendResponse.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Proxy search error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}
