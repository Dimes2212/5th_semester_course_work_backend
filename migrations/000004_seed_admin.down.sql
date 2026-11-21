DELETE FROM users
WHERE email = 'raduloffd@gmail.com'
  AND full_name = 'Administrator'
  AND password_hash = '$2y$12$9NuaYCJ2K4hvEJs3hYzXBOSw4Fxe1riZzVWbBmkld7aLPmmDGL8ce'
  AND role = 'admin';
