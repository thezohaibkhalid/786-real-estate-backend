insert into admin_users (name, email, password_hash, role) values (
  'Admin',
  'admin@786realestate.pk',
  '$argon2id$v=19$m=65536,t=1,p=4$qgJMRaWj1d7yEvvDtvAYvQ$xematO966nD6wzvmd/1jMtAo09/ZULdwmW09yLFch6g',
  'owner'
)
on conflict (email) do update set
  name = excluded.name,
  role = excluded.role;
