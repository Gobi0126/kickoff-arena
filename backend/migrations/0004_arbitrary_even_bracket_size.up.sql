ALTER TABLE spin_wheel_settings DROP CONSTRAINT spin_wheel_settings_bracket_size_check;
ALTER TABLE spin_wheel_settings ADD CONSTRAINT spin_wheel_settings_bracket_size_check
    CHECK (bracket_size >= 2 AND bracket_size <= 64 AND bracket_size % 2 = 0);
