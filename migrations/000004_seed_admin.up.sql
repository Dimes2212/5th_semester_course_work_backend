INSERT INTO users (
    full_name,
    email,
    password_hash,
    role
)
VALUES (
    'Administrator',
    'raduloffd@gmail.com',
    '$2y$12$9NuaYCJ2K4hvEJs3hYzXBOSw4Fxe1riZzVWbBmkld7aLPmmDGL8ce',
    'admin'
)
ON CONFLICT DO NOTHING;
