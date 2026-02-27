-- Drop old columns
ALTER TABLE flash_sales DROP COLUMN IF EXISTS start_date;
ALTER TABLE flash_sales DROP COLUMN IF EXISTS end_date;
ALTER TABLE flash_sales DROP COLUMN IF EXISTS days_of_week;

-- Add new column
ALTER TABLE flash_sales ADD COLUMN flash_date DATE;
