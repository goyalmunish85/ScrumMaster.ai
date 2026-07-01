import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  try {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const authCredentials = process.env.ADMIN_CREDENTIALS || 'admin:admin';
    const encodedCredentials = Buffer.from(authCredentials).toString('base64');

    const url = new URL(request.url);
    const limit = url.searchParams.get('limit') || '50';
    const offset = url.searchParams.get('offset') || '0';

    const res = await fetch(`${apiUrl}/api/activities?limit=${limit}&offset=${offset}`, {
      headers: {
        'Authorization': `Basic ${encodedCredentials}`,
        'Content-Type': 'application/json',
      },
    });

    if (!res.ok) {
      return NextResponse.json({ error: 'Failed to fetch activities' }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (err) {
    console.error('API Route Error:', err);
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}
