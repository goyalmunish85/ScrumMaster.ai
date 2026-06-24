import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import GlobalTimeline from './GlobalTimeline';

// Mock fetch
global.fetch = jest.fn(() =>
  Promise.resolve({
    json: () => Promise.resolve([
      {
        id: '1',
        task_id: null,
        event_type: 'TASK_CREATED',
        payload: '{"task_name": "Test Task"}',
        created_at: new Date().toISOString()
      }
    ]),
  })
) as jest.Mock;

describe('GlobalTimeline', () => {
  it('renders loading state initially', () => {
    const { container } = render(<GlobalTimeline />);
    // Loading indicator relies on a span with animate-ping class
    expect(container.querySelector('.animate-ping')).toBeInTheDocument();
  });

  it('renders events after fetch', async () => {
    render(<GlobalTimeline />);
    await waitFor(() => {
      expect(screen.getByText('TASK CREATED')).toBeInTheDocument();
    });
    expect(screen.getByText(/"task_name": "Test Task"/)).toBeInTheDocument();
  });
});
