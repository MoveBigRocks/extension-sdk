const { defineConfig } = require("@playwright/test");

module.exports = defineConfig({
  testDir: "./tests",
  use: {
    baseURL: process.env.MBR_BASE_URL || "http://127.0.0.1:8080",
    trace: "retain-on-failure"
  }
});
