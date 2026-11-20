# Using cc-desk

cc-desk runs your background Claude sessions and tells you when one needs you. You read its
status page to see how the work is going, and you answer its questions at the desk, a Claude
session in your terminal. This guide covers what you type, what the page shows, and what to do
when something is wrong.

## Open the desk

Type `cc-desk` in a terminal. A Claude session opens and tells you what is waiting:

- When nothing waits, it says so in one sentence and gives the time of the next question.
- When something waits, it says how many questions there are and what changed since your last
  visit. Then it asks the first question.

Each question comes with its background, what each answer would cause, and a recommendation.
Answer in your words. The desk records your words as your ruling and tells you the time it
recorded them. When your words don't settle the choice, it asks you to pick one answer.

A question about a pull request has three answers:

- Merge: it merges it on the next check, if every check still passes and nothing was pushed since you
  were asked.
- Fix again: it sends it back with your words for one more try.
- Close: it closes it, and the work that comes after it doesn't start.

## Read the status page

The page shows the records. The one thing you can change there is a report's "Mark as read"
button, which takes the report off your list.

The sidebar holds four parts, with a "Rulings" link above them:

- "Next decision session" says when you are next needed. It reads "Needs you now", "Ready
  now", a time, or "Nothing waiting". "Rough" beside the time means the estimate rests on little
  data.
- "Waiting for you" lists each round with questions ready now. Each line links to the round
  and can open the desk for it in a terminal.
- "Reports to read" lists the reports the sessions sent you. Each opens its report.
- "Running" shows each release as a bar, one segment per round, with how much of the planned
  work is done.

The main area repeats none of that:

- "Your next decision session" lists the rounds and questions the next session will cover.
- "Guesses waiting on an answer" lists work that started on a guess at your answer.
- "Since the last check" says what the background check did on its last run.
- "Sessions" lists each session that is running or stopped.
- Below them, one board per release shows each round and its passes.

## When the page says "Needs you now"

"Needs you now" means a session stopped on something only you can do. The page lists each
reason with the one thing to do. The common ones:

| The page says | What to do |
| --- | --- |
| A permission prompt | Open the desk and say yes or no. The desk passes on your answer and restarts the session. |
| The sign-in ran out | Sign in to Claude Code again. The check sees the new sign-in and starts the work again by itself. After it has done that twice for one pass, the page names the command to give the desk instead. |
| Usage limit reached | Wait for the reset, or reset your usage early. The work starts again after the reset. |
| The session would not start, The stop could not be read, or Restarted twice already | Read the error or the session's log, then tell the desk to run the command the page names, such as `cc-desk pass start proj a`. |
| An irreversible step, a secret, a sign-in or data change, or a change of scope | Read the session's log, deal with it, then tell the desk to start the pass again. |

## Fix a wrong ruling

If the desk recorded the wrong answer, tell it at the desk, in your words. It records the new
ruling in place of the old one, and both stay in the log. You can see every ruling under
"Rulings" in the page's sidebar.

If the work that waits on the ruling hasn't started, it starts on the new ruling. If it is being
built, its session restarts with the new ruling. Once that work has merged, a new change is
needed instead.
