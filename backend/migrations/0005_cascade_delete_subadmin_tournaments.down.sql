ALTER TABLE tournaments DROP CONSTRAINT tournaments_created_by_fkey;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);
