import { readDesign, type DesignRead } from '#lib/server/design.ts';
import { readState } from '#lib/server/state.ts';
import { findRoundAt } from '#lib/sums.ts';
import type { PageServerLoad } from './$types';

/**
 * Reads the round's design file when the snapshot records one. A round with no design, or an
 * address that names no round, reads no file.
 */
export const load: PageServerLoad = async ({ params }): Promise<{ design: DesignRead | null }> => {
	const { snapshot } = await readState();
	const found = snapshot ? findRoundAt(snapshot, params.project, params.round) : undefined;
	const path = found?.round.design;
	if (!path) return { design: null };
	return { design: await readDesign(params.project, path) };
};
