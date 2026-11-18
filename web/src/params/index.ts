import { defineParams } from '@sveltejs/kit/params';

export const params = defineParams({
	// The ids `cc-desk` gives a report: `rep-` and at least four digits.
	report: (param) => (/^rep-\d{4,}$/.test(param) ? param : undefined)
});
