BEGIN;

-- Delete all conferences
DELETE
FROM conferences
WHERE host_id IN (SELECT id
                  FROM users
                  WHERE email LIKE '%@seeder.nathakusuma.com');

-- Delete all seeded users
DELETE
FROM users
WHERE email LIKE '%@seeder.nathakusuma.com';

-- Drop the ULID generation function
DROP FUNCTION IF EXISTS generate_ulid_at_time(TIMESTAMP WITH TIME ZONE);

COMMIT;
