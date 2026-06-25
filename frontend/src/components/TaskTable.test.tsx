import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import TaskTable, { Task } from './TaskTable';

describe('TaskTable', () => {
  const mockTasks: Task[] = [
    {
      id: '1',
      title: 'Fix login bug',
      status: 'BLOCKED',
      assignee: 'Alice',
      due_date: '2023-10-27T00:00:00Z',
      updated_at: '2023-10-26T00:00:00Z',
      client: null,
      team: null,
      task_type: null,
      sprint: null,
      source_name: 'Jira',
    },
    {
      id: '2',
      title: 'Implement dashboard',
      status: 'IN_PROGRESS',
      assignee: 'Bob',
      due_date: null,
      updated_at: '2023-10-26T00:00:00Z',
      client: null,
      team: null,
      task_type: null,
      sprint: null,
      source_name: null,
    },
    {
      id: '3',
      title: 'Update documentation',
      status: 'DONE',
      assignee: null,
      due_date: '2023-10-28T00:00:00Z',
      updated_at: '2023-10-26T00:00:00Z',
      client: null,
      team: null,
      task_type: null,
      sprint: null,
      source_name: 'Linear',
    },
    {
      id: '4',
      title: 'Draft new feature spec',
      status: 'DRAFT',
      assignee: 'Charlie',
      due_date: '2023-10-29T00:00:00Z',
      updated_at: '2023-10-26T00:00:00Z',
      client: null,
      team: null,
      task_type: null,
      sprint: null,
      source_name: 'Asana',
    },
  ];

  it('renders table headers correctly', () => {
    render(<TaskTable tasks={mockTasks} />);
    expect(screen.getByText('Title')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('Assignee')).toBeInTheDocument();
    expect(screen.getByText('Due Date')).toBeInTheDocument();
    expect(screen.getByText('Source')).toBeInTheDocument();
  });

  it('renders the empty state when no tasks are provided', () => {
    render(<TaskTable tasks={[]} />);
    expect(screen.getByText('No tasks found.')).toBeInTheDocument();
  });

  it('renders tasks with various statuses', () => {
    render(<TaskTable tasks={mockTasks} />);

    // Check titles
    expect(screen.getByText('Fix login bug')).toBeInTheDocument();
    expect(screen.getByText('Implement dashboard')).toBeInTheDocument();
    expect(screen.getByText('Update documentation')).toBeInTheDocument();
    expect(screen.getByText('Draft new feature spec')).toBeInTheDocument();

    // Check statuses
    expect(screen.getByText('BLOCKED')).toBeInTheDocument();
    expect(screen.getByText('IN PROGRESS')).toBeInTheDocument();
    expect(screen.getByText('DONE')).toBeInTheDocument();
    expect(screen.getByText('DRAFT')).toBeInTheDocument();
  });

  it('handles presence and absence of optional fields correctly', () => {
    render(<TaskTable tasks={mockTasks} />);

    // Assignee
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    expect(screen.getByText('Unassigned')).toBeInTheDocument();

    // Due Date
    // Note: the component uses toLocaleDateString() which depends on the environment
    // We check that the formatted date or dash is rendered.
    // For id=2 due_date is null, so there should be a '-'
    const dashElements = screen.getAllByText('-');
    expect(dashElements.length).toBeGreaterThan(0);

    // Source
    expect(screen.getByText('Jira')).toBeInTheDocument();
    expect(screen.getByText('Linear')).toBeInTheDocument();
    expect(screen.getByText('Asana')).toBeInTheDocument();
  });
});
