import { test, expect } from '@playwright/test';

test('Jira Config Page Add Project', async ({ page }) => {
  await page.goto('http://localhost:3000/jira-config');

  // Wait for the page to load
  await page.waitForSelector('text=Jira Configuration');

  // Add a project
  await page.fill('input[id="projectKey"]', 'TESTPROJ');
  await page.click('button[aria-label="Add Jira Project Key"]');

  // Wait for the project to appear
  await page.waitForSelector('text=TESTPROJ');

  // Take screenshot
  await page.screenshot({ path: 'jira-config-page-added.png', fullPage: true });
});
