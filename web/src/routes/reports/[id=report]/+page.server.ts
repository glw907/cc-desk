import { fail } from '@sveltejs/kit';
import { markRead } from '#lib/server/command.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	// The page's only write. The matcher has checked the id, and markRead checks it again.
	read: async ({ params }) => {
		const result = await markRead(params.id);
		if (!result.ok) return fail(500, { message: result.message });
	}
};
