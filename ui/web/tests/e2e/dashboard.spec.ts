import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('loads dashboard page', async ({ page }) => {
    await page.goto('/');

    // Check for main heading
    await expect(page.locator('h1')).toContainText('Dashboard');

    // Check for connection status
    await expect(page.locator('text=/Connected|Connecting|Disconnected/')).toBeVisible();
  });

  test('navigates between pages', async ({ page }) => {
    await page.goto('/');

    // Navigate to positions
    await page.click('text=Positions');
    await expect(page).toHaveURL('/positions');
    await expect(page.locator('h1')).toContainText('Positions');

    // Navigate to orders
    await page.click('text=Orders');
    await expect(page).toHaveURL('/orders');
    await expect(page.locator('h1')).toContainText('Orders');
  });

  test('displays account stats', async ({ page }) => {
    await page.goto('/');

    // Check for stat cards
    await expect(page.locator('text=Balance')).toBeVisible();
    await expect(page.locator('text=Equity')).toBeVisible();
    await expect(page.locator('text=Unrealized P&L')).toBeVisible();
  });
});
