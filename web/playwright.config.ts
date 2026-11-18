import { defineConfig, devices } from '@playwright/test';

// retries stay 0 and forbidOnly stays on: a flaky or focused test must fail the gate,
// never pass on a second try. The app starts per test from tests/app.ts, so no webServer.
export default defineConfig({
	testDir: 'tests',
	testMatch: '**/*.spec.ts',
	retries: 0,
	forbidOnly: true,
	workers: 1,
	globalSetup: './tests/global-setup.ts',
	reporter: [['list'], ['json', { outputFile: 'test-results/report.json' }]],
	use: {
		timezoneId: 'Europe/London',
		locale: 'en-US'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});
