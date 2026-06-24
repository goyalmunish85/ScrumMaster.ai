import { test, expect } from '@playwright/test';

test('Slack Config UI tests', async ({ page }) => {
  await page.goto('http://localhost:3000/slack-config');

  await expect(page.locator('text=Slack Configuration')).toBeVisible();

  // Create a channel
  const testId = `TESTCHANNEL${Date.now()}`;
  await page.fill('input[placeholder="e.g. C0AQMS8J0P3"]', testId);
  await page.click('button:has-text("Add Channel")');

  // Verify the channel gets added to the list
  await expect(page.locator(`text=${testId}`).first()).toBeVisible({ timeout: 5000 });

  await page.screenshot({ path: 'slack-config-screenshot.png', fullPage: true });
});
