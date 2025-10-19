-- Add 'booking' to the source_website enum type
ALTER TYPE source_website ADD VALUE 'booking';

-- Verify the change
SELECT unnest(enum_range(NULL::source_website)) as valid_sources;
