// markdown-it 15 ships no typings. This declares only what lib/server/design.ts calls, so the
// page takes no second package for types.
declare module 'markdown-it' {
	interface Options {
		html?: boolean;
		linkify?: boolean;
	}

	class MarkdownIt {
		constructor(options?: Options);
		/** Rules to turn off; an unknown name throws. */
		disable(rules: string | string[], ignoreInvalid?: boolean): this;
		/** Replaceable: says whether a link's address may become an href. */
		validateLink: (url: string) => boolean;
		render(source: string): string;
	}

	export default MarkdownIt;
}
