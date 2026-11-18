import { readState } from '#lib/server/state.ts';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () =>
	Response.json(await readState(), { headers: { 'cache-control': 'no-store' } });
