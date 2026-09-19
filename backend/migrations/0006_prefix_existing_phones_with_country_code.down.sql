UPDATE users
SET phone = substring(phone from 4)
WHERE phone LIKE '+91%';
