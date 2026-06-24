import { test, expect } from '@playwright/test';

test.describe('Jira Config Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:3000/jira-config');
  });

  test('should render the form with correct initial state', async ({ page }) => {
    await expect(page.locator('h1')).toHaveText('Jira Configuration');
    await expect(page.locator('input#projectKey')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toHaveText('Add Project');
  });

  test('should show error on empty submission', async ({ page }) => {
    await page.click('button[type="submit"]');
    const errorMessage = page.locator('#error-message');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toHaveText('Jira Project Key cannot be empty.');
  });

  test('should show error on invalid Jira Project Key format', async ({ page }) => {
    // lowercase first letter
    await page.fill('input#projectKey', 'sAAS');
    // NOTE: Our component auto-capitalizes on change! So typing 'sAAS' will actually become 'SAAS'.
    // We should test an invalid character instead, like a symbol.
    await page.fill('input#projectKey', 'SAAS!');
    await page.click('button[type="submit"]');

    const errorMessage = page.locator('#error-message');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('Invalid format');
  });

  test('should submit successfully with valid key', async ({ page }) => {
    // Intercept the API call to mock a successful response
    await page.route('**/api/v1/integrations/targets', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: '123', platform: 'jira', target_id: 'TESTPROJECT' }),
      });
    });

    await page.fill('input#projectKey', 'TESTPROJECT');
    await page.click('button[type="submit"]');

    // Wait for the success state
    const submitBtn = page.locator('button[type="submit"]');
    await expect(submitBtn).toContainText('Connected Successfully');

    // Check that the input is cleared
    await expect(page.locator('input#projectKey')).toHaveValue('');
  });

  test('should show generic error message on API failure', async ({ page }) => {
    // Intercept the API call to mock a server error
    await page.route('**/api/v1/integrations/targets', async route => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      });
    });

    await page.fill('input#projectKey', 'PROJ123');
    await page.click('button[type="submit"]');

    const errorMessage = page.locator('#error-message');
    await expect(errorMessage).toBeVisible();
    // It should be a generic error, not exposing server details
    await expect(errorMessage).toHaveText('An unexpected network error occurred.');
  });
});
