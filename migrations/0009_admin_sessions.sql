create table admin_sessions (
  token_hash text primary key,
  admin_user_id bigint not null references admin_users(id) on delete cascade,
  created_at timestamptz not null default now(),
  last_seen_at timestamptz not null default now(),
  expires_at timestamptz not null
);

create index admin_sessions_user_idx on admin_sessions(admin_user_id);
create index admin_sessions_expiry_idx on admin_sessions(expires_at);
