import { test, expect } from '@playwright/test';

test('click on task row opens detail modal', async ({ page }) => {
  await page.goto('http://localhost:3000/dashboard');

  // Wait for the tasks to load
  await page.waitForSelector('table tbody tr:not(:has-text("No tasks found"))', { timeout: 15000 });

  // Take screenshot before click
  await page.screenshot({ path: '/home/jules/verification/screenshots/dashboard-with-data.png' });

  // Click the first task row
  const firstTaskRow = page.locator('table tbody tr').first();
  await firstTaskRow.click();

  // Wait a moment for modal to animate in
  await page.waitForSelector('role=dialog');
  await page.waitForTimeout(1000);

  // Take screenshot after click
  await page.screenshot({ path: '/home/jules/verification/screenshots/task-modal.png' });
});
