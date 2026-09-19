ALTER TABLE spin_wheel_settings DROP CONSTRAINT spin_wheel_settings_bracket_size_check;
ALTER TABLE spin_wheel_settings ADD CONSTRAINT spin_wheel_settings_bracket_size_check
    CHECK (bracket_size IN (8, 16, 32));
