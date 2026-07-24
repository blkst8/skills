# Deployment Notification Agent

## Purpose
Notify team members when a deployment finishes.

## Steps
1. Listen for the `deployment_complete` event.
2. Extract the environment (staging or production) and the commit hash.
3. Format a message: "Deployment to `{environment}` finished: `{commit_hash}`".
4. Post the message to the general channel.
5. Log the event details to the console at every step for debugging.
