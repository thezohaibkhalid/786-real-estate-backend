create table if not exists admin_password_reset_tokens (
  id bigint generated always as identity primary key,
  admin_user_id bigint not null references admin_users(id) on delete cascade,
  token_hash text not null unique,
  expires_at timestamptz not null,
  used_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists admin_password_reset_tokens_user_idx
  on admin_password_reset_tokens (admin_user_id, created_at desc);

create index if not exists admin_password_reset_tokens_expiry_idx
  on admin_password_reset_tokens (expires_at)
  where used_at is null;
