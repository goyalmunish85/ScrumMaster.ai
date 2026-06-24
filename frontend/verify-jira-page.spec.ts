import { test, expect } from '@playwright/test';

test('Jira Config Page verification', async ({ page }) => {
  await page.goto('http://localhost:3000');

  // Navigate to Jira Config page
  await page.click('a[href="/jira-config"]');

  // Wait for the page to load
  await page.waitForSelector('text=Jira Configuration');

  // Verify main elements
  await expect(page.locator('text=Add New Project')).toBeVisible();
  await expect(page.locator('input[id="projectKey"]')).toBeVisible();

  // Take screenshot
  await page.screenshot({ path: 'jira-config-page.png', fullPage: true });
});
