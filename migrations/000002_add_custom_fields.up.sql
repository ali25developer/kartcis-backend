-- Add custom fields to events
ALTER TABLE events ADD COLUMN custom_fields TEXT;

-- Add custom field responses to tickets
ALTER TABLE tickets ADD COLUMN custom_field_responses TEXT;
