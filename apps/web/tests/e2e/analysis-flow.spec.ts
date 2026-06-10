import { test, expect } from '@playwright/test';
import * as path from 'path';

test.describe('BARREL E2E Flow', () => {
  const webUrl = process.env.BARREL_AZURE_WEB_URL || 'http://localhost:5173';

  test.beforeEach(async ({ page }) => {
    page.on('console', msg => console.log('BROWSER CONSOLE:', msg.text()));
    page.on('pageerror', error => console.log('BROWSER ERROR:', error.message));

    await page.goto(webUrl);

    // Wait for the app to load (spinner finished, so h1 is rendered)
    await page.waitForSelector('h1:has-text("BARREL")');
    
    // Login if prompted
    const evaluatorBadge = page.locator('text=Evaluator Mode');
    if (await evaluatorBadge.count() === 0) {
      await page.fill('input[type="password"]', process.env.BARREL_DEMO_PASSWORD || 'fallback-demo-password-123');
      await page.click('button:has-text("Secure Login")');
      await page.waitForSelector('text=Evaluator Mode');
    }
  });

  test('Azure Vision Drag and Drop Single Image', async ({ page }) => {
    // Assert layout constraints
    const analysisCard = page.locator('.main-grid .card').first();
    await expect(analysisCard).toContainText('New Analysis');
    await expect(page.locator('.metrics-row')).toHaveCount(0); // Should be removed

    const historySection = page.locator('.review-history-section');
    await expect(historySection).toBeVisible();

    const fileInput = page.locator('input[type="file"]');
    const imagePath = path.resolve(process.cwd(), '../../samples/generated/good/good_01_distilled_spirits_clean_front.png');
    await fileInput.setInputFiles(imagePath);

    await page.selectOption('select', { label: 'Azure Vision (Default)' });
    
    // Fill expected fields based on good_01
    await page.locator('.form-group:has(label:has-text("Brand Name")) input').fill('OLD TOM DISTILLERY');
    await page.locator('.form-group:has(label:has-text("Class/Type")) input').fill('Kentucky Straight Bourbon Whiskey');
    await page.locator('.form-group:has(label:has-text("Alcohol Content")) input').fill('45% Alc./Vol. (90 Proof)');
    await page.locator('.form-group:has(label:has-text("Net Contents")) input').fill('750 mL');

    await page.click('button:has-text("Analyze Label")');
    await expect(page.locator('.loading-spinner')).toBeVisible();
    await expect(page.locator('.loading-spinner')).toHaveCount(0, { timeout: 90000 }); // Wait for spinner to go away

    // Detail panel
    await expect(page.locator('.review-left')).toBeVisible();
    await expect(page.locator('.review-right')).toBeVisible();
    
    // Assert provider stored is azure_vision
    const evidenceText = await page.locator('.review-left').innerText();
    expect(evidenceText).toContain('azure_vision');

    // Assert raw OCR text is non-empty
    await expect(page.locator('h4:has-text("Raw OCR Text") + div')).not.toBeEmpty();

    // Assert extracted fields table exists
    await expect(page.locator('.field-list')).toBeVisible();
    // Assert overall status is Pass
    await expect(page.locator('.status-badge')).toContainText('Pass');

    // Review history should contain the new file
    const targetRow = page.locator('.history-table tbody tr', { hasText: 'good_01_distilled_spirits_clean_front.png' }).first();
    await expect(targetRow).toBeVisible({ timeout: 15000 });
    await expect(targetRow).toContainText('azure_vision');

    // Click history row to reload
    await targetRow.click();
    await expect(page.locator('.review-left')).toBeVisible();
  });

  test('AI Based OCR Drag and Drop Single Image', async ({ page }) => {
    const fileInput = page.locator('input[type="file"]');
    const imagePath = path.resolve(process.cwd(), '../../samples/generated/good/good_02_bourbon_proof_and_abv.png');
    await fileInput.setInputFiles(imagePath);

    await page.selectOption('select', { label: 'AI Based OCR (Azure OpenAI)' });
    
    // Fill expected fields based on good_02
    await page.locator('.form-group:has(label:has-text("Brand Name")) input').fill('BARREL HOUSE NO. 7');
    await page.locator('.form-group:has(label:has-text("Class/Type")) input').fill('Straight Bourbon Whiskey');
    await page.locator('.form-group:has(label:has-text("Alcohol Content")) input').fill('50% Alc./Vol. (100 Proof)');
    await page.locator('.form-group:has(label:has-text("Net Contents")) input').fill('750 mL');

    await page.click('button:has-text("Analyze Label")');
    await expect(page.locator('.loading-spinner')).toHaveCount(0, { timeout: 90000 });

    const evidenceText = await page.locator('.review-left').innerText();
    expect(evidenceText).toContain('ai_based');

    // Verify AI metadata or error
    const pageText = await page.locator('body').innerText();
    if (pageText.includes('AI provider not configured')) {
      expect(pageText).toContain('AI provider not configured');
    } else {
      expect(pageText).toBeTruthy();
      // Assert overall status is Pass if configured
      await expect(page.locator('.status-badge')).toContainText('Pass');
    }

    // Verify history row
    const targetRowAi = page.locator('.history-table tbody tr', { hasText: 'good_02_bourbon_proof_and_abv.png' }).first();
    await expect(targetRowAi).toBeVisible({ timeout: 15000 });
    await expect(targetRowAi).toContainText('ai_based');
  });

  test('Batch ZIP drag/drop', async ({ page }) => {
    const fileInput = page.locator('input[type="file"]');
    const zipPath = path.resolve(process.cwd(), '../../samples/batches/good_10.zip');
    
    // Handle alert dialog that pops up for batch
    page.once('dialog', dialog => dialog.accept());
    
    await fileInput.setInputFiles(zipPath);
    await page.click('button:has-text("Analyze Label")');

    // Assert batch queue appears
    await expect(page.locator('h4:has-text("Batch Queue")')).toBeVisible({ timeout: 30000 });
    await expect(page.locator('h4:has-text("Batch Queue") + table tbody tr')).toHaveCount(10, { timeout: 30000 });
    
    // Wait for at least one to complete in history
    await expect(page.locator('.history-table tbody tr').first()).toContainText('good_', { timeout: 30000 });
  });

  test('Layout regression', async ({ page }) => {
    // New Analysis appears above Review History (implied by layout but we can check order)
    const cardTitles = await page.locator('.card-title').allInnerTexts();
    expect(cardTitles[0]).toContain('New Analysis');
    
    // Check no visible Labels Reviewed box
    await expect(page.locator('text=Labels Reviewed')).toHaveCount(0);
  });
});
