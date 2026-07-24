# Data Cleanup Pipeline

## Purpose
Clean and normalize CSV files dropped into the shared drive.

## Steps
1. Watch `~/shared/inbox/` for new `.csv` files.
2. Read the file and drop any rows with empty values in the `email` column.
3. Normalize phone numbers to the `+1-XXX-XXX-XXXX` format.
4. Write the cleaned data to `/tmp/output.csv`.
5. Send a Slack message to `#data-ops` with the row count.
