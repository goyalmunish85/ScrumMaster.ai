import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test('should render the dashboard layout with timeline', async ({ page }) => {
    // We navigate to dashboard
    await page.goto('http://localhost:3000/dashboard');

    // Check if Operations Dashboard header is visible
    await expect(page.locator('text=Operations Dashboard')).toBeVisible();

    // Check if Active Tasks is visible
    await expect(page.locator('text=Active Tasks').first()).toBeVisible();

    // Check if Activity Timeline is visible
    await expect(page.locator('text=Activity Timeline').first()).toBeVisible();
  });
});
