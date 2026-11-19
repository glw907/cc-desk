# Refuse a replacing ruling once a close answer has begun

When Geoff answers a merge question with "close", the check closes the pull request and then
ends the session. If ending the session fails, the pass stays building with a closed pull
request until the next check. In that window a replacing ruling can record "merge" for the
same question, and the pass then carries a merge answer for a pull request that is already closed.

Decide whether the replace step refuses a pass whose close answer has begun, or whether the next
check treats the closed pull request as the stronger record and drops the merge answer.
