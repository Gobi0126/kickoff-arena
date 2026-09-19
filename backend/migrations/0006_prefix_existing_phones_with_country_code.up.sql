-- Login/signup now combine a country-code dropdown (defaulting to +91) with
-- the number, so the stored phone is always in "+<code><number>" form. Any
-- account created before that change has a bare number (e.g. "9876543210"),
-- which no longer matches what login sends ("+919876543210") — backfill
-- those rows so existing subadmins can still log in.
UPDATE users
SET phone = '+91' || phone
WHERE phone NOT LIKE '+%';
