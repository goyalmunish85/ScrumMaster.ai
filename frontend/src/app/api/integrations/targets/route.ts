import { NextRequest, NextResponse } from 'next/server';

const API_URL = process.env.API_URL || 'http://localhost:8080';
const ADMIN_CREDENTIALS = process.env.ADMIN_CREDENTIALS || 'admin:admin';

const getHeaders = () => {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Basic ${Buffer.from(ADMIN_CREDENTIALS).toString('base64')}`,
  };
};

export async function GET(request: NextRequest) {
  try {
    const res = await fetch(`${API_URL}/api/v1/integrations/targets`, {
      method: 'GET',
      headers: getHeaders(),
    });

    if (!res.ok) {
      const text = await res.text();
      return NextResponse.json({ error: text || 'Failed to fetch targets' }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error proxying GET /api/v1/integrations/targets:', error);
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    const res = await fetch(`${API_URL}/api/v1/integrations/targets`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      const text = await res.text();
      return NextResponse.json({ error: text || 'Failed to create target' }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error proxying POST /api/v1/integrations/targets:', error);
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}

export async function DELETE(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const id = searchParams.get('id');

    if (!id) {
      return NextResponse.json({ error: 'Missing id parameter' }, { status: 400 });
    }

    const res = await fetch(`${API_URL}/api/v1/integrations/targets?id=${id}`, {
      method: 'DELETE',
      headers: getHeaders(),
    });

    if (!res.ok) {
      const text = await res.text();
      return NextResponse.json({ error: text || 'Failed to delete target' }, { status: res.status });
    }

    // Attempt to parse JSON response if any, but DELETE might return 204 No Content
    if (res.status === 204) {
      return new NextResponse(null, { status: 204 });
    }

    const text = await res.text();
    let data;
    try {
      data = JSON.parse(text);
    } catch {
      data = { message: text || 'Target deleted' };
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error('Error proxying DELETE /api/v1/integrations/targets:', error);
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}
