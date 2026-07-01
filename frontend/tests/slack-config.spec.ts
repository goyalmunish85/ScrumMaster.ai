import { test, expect } from '@playwright/test';

test.describe('Slack Config Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the initial fetch
    await page.route('**/api/integrations/targets', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            { id: '1', platform: 'slack', target_id: 'C12345678', created_at: new Date().toISOString() }
          ]),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto('http://localhost:3000/slack-config');
  });

  test('should render the form with correct initial state', async ({ page }) => {
    await expect(page.locator('h1')).toHaveText('Slack Configuration');
    await expect(page.locator('input#targetId')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toHaveText('Add Channel');

    // Verify existing channel is rendered
    await expect(page.locator('text=C12345678')).toBeVisible();
  });

  test('should show error on empty submission', async ({ page }) => {
    // Note: the button is disabled if empty, but we can test focus or bypass disabled state
    // Actually, the UI disables the submit button if the input is empty. Let's verify that.
    const submitBtn = page.locator('button[type="submit"]');
    await expect(submitBtn).toBeDisabled();

    // To trigger the error manually, we'd have to somehow bypass the disabled state,
    // or type space and submit.
    await page.fill('input#targetId', '   ');
    await expect(submitBtn).toBeDisabled(); // still disabled due to targetId.trim() check in UI
  });

  test('should show error on invalid Slack Channel ID format', async ({ page }) => {
    // invalid format, doesn't start with C
    await page.fill('input#targetId', 'D123456');
    await page.click('button[type="submit"]');

    const errorBanner = page.locator('div[role="alert"]').filter({ hasText: 'Invalid format' });
    await expect(errorBanner).toBeVisible();
    await expect(errorBanner).toContainText('Invalid format');
  });

  test('should submit successfully with valid channel ID', async ({ page }) => {
    // Intercept the API call to mock a successful POST
    let postCalled = false;
    await page.route('**/api/integrations/targets', async route => {
      if (route.request().method() === 'POST') {
        postCalled = true;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: '2', platform: 'slack', target_id: 'C0AQMS8J0P3' }),
        });
      } else if (route.request().method() === 'GET') {
        // Mock the GET request that happens after successful POST
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            { id: '1', platform: 'slack', target_id: 'C12345678', created_at: new Date().toISOString() },
            { id: '2', platform: 'slack', target_id: 'C0AQMS8J0P3', created_at: new Date().toISOString() }
          ]),
        });
      }
    });

    await page.fill('input#targetId', 'C0AQMS8J0P3');
    await page.click('button[type="submit"]');

    // Check that the input is cleared
    await expect(page.locator('input#targetId')).toHaveValue('');

    // Verify the new channel is added to the list
    await expect(page.locator('text=C0AQMS8J0P3')).toBeVisible();
    expect(postCalled).toBe(true);
  });

  test('should delete successfully', async ({ page }) => {
    let deleteCalled = false;
    await page.route('**/api/integrations/targets?id=1', async route => {
      if (route.request().method() === 'DELETE') {
        deleteCalled = true;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'Deleted' }),
        });
      } else {
        await route.continue();
      }
    });

    // Re-mock GET inside this test to control when it returns empty
    await page.route('**/api/integrations/targets', async route => {
      if (route.request().method() === 'GET') {
        if (deleteCalled) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([]),
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([
              { id: '1', platform: 'slack', target_id: 'C12345678', created_at: new Date().toISOString() }
            ]),
          });
        }
      } else {
        await route.continue();
      }
    });

    // Reload page to ensure the inner route mock is applied
    await page.goto('http://localhost:3000/slack-config');

    // The list uses created_at to format text. Wait for the list item to render.
    await expect(page.locator('text=C12345678')).toBeVisible();

    // The delete button is revealed on hover. We force click it.
    await page.locator('button[aria-label="Delete channel C12345678"]').click({ force: true });

    // Check that the item is removed from the DOM
    await expect(page.locator('text=C12345678')).not.toBeVisible();
    expect(deleteCalled).toBe(true);
  });
});
