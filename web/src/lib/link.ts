const name = /^[a-z0-9][a-z0-9._-]*$/;

/**
 * The address that opens a desk in a terminal, or `null` when a name fails the rule.
 *
 * @remarks
 * The rule is the desktop handler's (`web/desktop/cc-desk-open`): a name starts with a lower-case
 * letter or digit and holds only those, `.`, `_`, and `-`. That is the state's project rule. The handler refuses every other address, so the
 * page shows no link for a name it would refuse.
 */
export function terminalLink(project: string, round: string): string | null {
	return name.test(project) && name.test(round) ? `cc-desk://${project}/${round}` : null;
}
