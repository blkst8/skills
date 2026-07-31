# Email Approval Workflow

## Purpose
Fetch incoming emails and route them to managers for approval.

## Steps
1. Poll the Gmail inbox every 5 minutes.
2. Extract the sender, subject, and body.
3. If the subject contains `[APPROVE]`, forward the email to the assigned manager.
4. Wait for the manager to reply with "APPROVED" or "REJECTED".
5. If no reply is received within 1 hour, retry sending the approval request up to 3 times.
6. On approval, mark the original email as processed and trigger the downstream webhook.
7. On rejection, move the email to the "Rejected" folder.