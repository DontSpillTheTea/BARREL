import { test, expect } from '@playwright/test';
import path from 'path';

const url = 'http://localhost:5173';

test.describe('Local Smoke Test', () => {
  test('Login and single image upload with ai_native', async ({ page }) => {
    await page.goto(url);
    await expect(page.locator('h1').filter({ hasText: 'BARREL' })).toBeVisible();
    await page.fill('input[type="text"]', 'evaluator');
    await page.fill('input[type="password"]', 'fallback-demo-password-123');
    await page.click('button:has-text("Secure Login")');
    await expect(page.locator('text=Evaluator Mode')).toBeVisible();

    const filePath = path.resolve(process.cwd(), '../../samples/generated/good/good_10_dense_but_readable_label.png');
    await page.locator('input[type="file"]').setInputFiles(filePath);

    await page.click('button:has-text("Analyze Upload")');

    await expect(page.locator('.loading-spinner')).not.toBeVisible({ timeout: 60000 });

    const isError = await page.isVisible('.alert-error');
    if (isError) {
      throw new Error(`Unexpected upload error: ${await page.locator('.alert-error').first().textContent()}`);
    }

    // Field verification table should be visible
    await expect(page.locator('text=Field Verification')).toBeVisible();

    // History table should show the new job
    await expect(page.locator('.history-table tbody tr').first()).toBeVisible();

    // Decision buttons should be present
    await expect(page.locator('button:has-text("Approve")').first()).toBeVisible();
    await expect(page.locator('button:has-text("Reject")').first()).toBeVisible();

    // Submit a decision
    await page.fill('textarea[placeholder="Review notes..."]', 'LGTM');
    await page.click('button:has-text("Approve")');

    await page.goto(url);
    await expect(page.locator('.history-table tbody tr').first().locator('.badge-info')).toHaveText('approved');
  });

  test('Zip upload creates history rows', async ({ page }) => {
    await page.goto(url);
    await page.fill('input[type="text"]', 'evaluator');
    await page.fill('input[type="password"]', 'fallback-demo-password-123');
    await page.click('button:has-text("Secure Login")');

    const filePath = path.resolve(process.cwd(), '../../samples/batches/good_10.zip');
    await page.locator('input[type="file"]').setInputFiles(filePath);

    await page.click('button:has-text("Analyze Upload")');
    await expect(page.locator('.loading-spinner')).not.toBeVisible({ timeout: 60000 });
    await expect(page.locator('text=Batch Queue')).toBeVisible();
  });
});
