const { test, expect } = require("@playwright/test");

const extensionPath = process.env.MBR_EXTENSION_PATH || "/extensions/sample-ops-extension";
const publicPath = process.env.MBR_PUBLIC_PATH || "/sample-ops-extension";

test("extension admin page renders", async ({ page }) => {
  await page.goto(extensionPath);
  await expect(page.locator("body")).toBeVisible();
});

test("extension public page renders", async ({ page }) => {
  await page.goto(publicPath);
  await expect(page.locator("body")).toBeVisible();
});
