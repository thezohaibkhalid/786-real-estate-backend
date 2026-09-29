-- =========================================================================
-- 786 Real Estate — PostgreSQL schema
--
-- Source of truth: 786-real-estate-admin/src/lib/types.ts (the admin panel's
-- domain model was explicitly designed to map 1:1 onto this schema) plus the
-- public site's data shape in 786-real-estate-frontend/src/lib/data.ts.
--
-- Conventions:
--   - snake_case identifiers, singular enum type names, plural table names.
--   - Every content table gets created_at; anything an admin actively edits
--     also gets updated_at (auto-maintained by set_updated_at()), even where
--     the current admin-panel mock layer didn't bother — real editorial
--     content needs edit history.
--   - `city` columns are plain text with a natural-key FK to cities.name
--     rather than cities.id, mirroring the admin type layer's own choice to
--     keep city denormalized (low cardinality, rarely renamed) while giving
--     areas a full surrogate-key FK (areas are rich SEO content, not a
--     closed list). This still gets full referential integrity.
--   - Array-typed fields with no dedicated child-table interface in
--     types.ts (Property.features, SeoPage.tags) are native Postgres arrays.
--     Fields that DID get their own interface (ProjectAmenityTag,
--     SeoPagePin, AgentArea) are real junction tables.
--   - `active`/`published`/status enums replace the design mockup's bare
--     booleans where the mockup was clearly a placeholder for a real
--     workflow (property_status upgrades a boolean to Draft/Published/
--     Sold/Rented/Archived, matching the admin type layer's own upgrade).
-- =========================================================================


create extension if not exists citext;

-- ─────────────────────────────────────────────────────────────────────────
-- Helper: auto-maintained updated_at
-- ─────────────────────────────────────────────────────────────────────────

create or replace function set_updated_at() returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

-- ─────────────────────────────────────────────────────────────────────────
-- Enums
-- ─────────────────────────────────────────────────────────────────────────

create type property_type as enum ('House', 'Plot', 'Apartment', 'Commercial', 'Farm House');
create type purpose as enum ('sale', 'rent');
create type size_unit as enum ('Marla', 'Kanal', 'Sq. Ft.');
create type furnished_state as enum ('Furnished', 'Semi-furnished', 'Unfurnished');

-- Upgrades the design mockup's bare `active` boolean.
create type property_status as enum ('Draft', 'Published', 'Sold', 'Rented', 'Archived');

create type project_status as enum ('Launching', 'Under Construction', 'Ready to Move');

create type blog_post_status as enum ('Draft', 'Published', 'Scheduled');
create type blog_category as enum (
  'Buying Guide', 'Society Reviews', 'Market Updates',
  'Commercial', 'Legal & Documents', 'Overseas'
);

create type lead_status as enum ('New', 'Contacted', 'Closed');

-- The 12 public-site forms that can produce a lead.
create type lead_form as enum (
  'Property Inquiry', 'Project Inquiry', 'Contact Form', 'Requirement Form',
  'Sell / Valuation', 'Rent Out', 'Site Visit', 'Brochure Download',
  'Call Back', 'Exit Pop-up', 'Saved Search', 'Newsletter'
);

-- SEO pages can target a single property type or every type.
create type seo_page_type as enum ('All', 'House', 'Plot', 'Apartment', 'Commercial', 'Farm House');
create type seo_page_tag as enum (
  'corner', 'installments', 'possession', 'furnished', 'overseas-pick', 'park-facing'
);

-- Polymorphic FAQ owner types.
create type faq_entity_type as enum ('project', 'seo_page', 'area', 'blog_post', 'home', 'overseas');

create type legal_document_key as enum ('privacy', 'terms', 'cookies', 'disclaimer');
create type currency_code as enum ('GBP', 'USD', 'AED');
create type admin_role as enum ('owner', 'manager', 'editor');

-- Controlled vocabulary for property features (design ships exactly these 15).
create type property_feature as enum (
  'Lawn', 'Terrace', 'Balcony', 'Drawing room', 'TV lounge',
  'Gated community', 'Security guards', 'CCTV', 'Gas', 'Electricity',
  'Solar system', 'Corner', 'Park facing', 'Near masjid', 'Near park'
);

-- ─────────────────────────────────────────────────────────────────────────
-- Geography
-- ─────────────────────────────────────────────────────────────────────────

create table cities (
  id          bigint generated always as identity primary key,
  name        citext not null unique,
  is_primary  boolean not null default false
  -- Property/project counts are COMPUTED (see v_city_stats below), never stored.
);

-- Enforce at most one primary city.
create unique index one_primary_city on cities (is_primary) where is_primary;

create table areas (
  id             bigint generated always as identity primary key,
  name           text not null,
  city           citext not null references cities(name) on update cascade,
  cover_image_url text not null default '',
  h1             text not null default '',
  meta_title     text not null default '',
  meta_desc      text not null default '',
  description    text not null default '',   -- rich text (HTML)
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  unique (city, name)
);
create trigger areas_set_updated_at before update on areas
  for each row execute function set_updated_at();
create index areas_city_idx on areas (city);

-- ─────────────────────────────────────────────────────────────────────────
-- Agents
-- ─────────────────────────────────────────────────────────────────────────

create table agents (
  id         bigint generated always as identity primary key,
  name       text not null,
  role       text not null default '',
  phone      text not null,
  whatsapp   text not null,
  email      citext not null,
  exp        integer not null default 0 check (exp >= 0),   -- years of experience
  deals      integer not null default 0 check (deals >= 0),
  photo_url  text not null default '',
  bio        text not null default '',        -- rich text (HTML)
  active     boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
create trigger agents_set_updated_at before update on agents
  for each row execute function set_updated_at();

-- M:M between agents and the areas they cover.
create table agent_areas (
  agent_id bigint not null references agents(id) on delete cascade,
  area_id  bigint not null references areas(id) on delete cascade,
  primary key (agent_id, area_id)
);

-- ─────────────────────────────────────────────────────────────────────────
-- Properties
-- ─────────────────────────────────────────────────────────────────────────

create table properties (
  id            bigint generated always as identity primary key,
  title         text not null,
  city          citext not null references cities(name) on update cascade,
  area_id       bigint not null references areas(id) on delete restrict,
  type          property_type not null,
  purpose       purpose not null,
  price         numeric(14, 0) not null check (price >= 0),   -- PKR
  size          numeric(10, 2) not null check (size > 0),
  unit          size_unit not null,
  beds          smallint not null default 0 check (beds >= 0),
  baths         smallint not null default 0 check (baths >= 0),
  kitchens      smallint not null default 0 check (kitchens >= 0),
  parking       smallint not null default 0 check (parking >= 0),
  floors        text not null default '',      -- free text, e.g. "Double storey"
  furnished     furnished_state not null default 'Unfurnished',
  approved      text not null default '',      -- approving authority, e.g. FDA/LDA/CDA/MDA
  youtube       text not null default '',
  maps          text not null default '',
  features      property_feature[] not null default '{}',
  description   text not null default '',      -- rich text (HTML)
  status        property_status not null default 'Draft',
  home          boolean not null default false,  -- "Featured" flag for the homepage
  home_sort_order integer not null default 0,
  slug          citext not null unique,
  meta_title    text not null default '',
  meta_desc     text not null default '',
  agent_id      bigint references agents(id) on delete set null,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
create trigger properties_set_updated_at before update on properties
  for each row execute function set_updated_at();
create index properties_search_idx on properties (city, area_id, type, purpose, status);
create index properties_price_idx on properties (purpose, price);
create index properties_home_idx on properties (home, home_sort_order) where home;

create table property_media (
  id         bigint generated always as identity primary key,
  property_id bigint not null references properties(id) on delete cascade,
  category   text not null,   -- e.g. Exterior/Interior/Bedroom/Kitchen/Bathroom, or custom
  url        text not null,
  sort_order integer not null default 0
);
create index property_media_property_idx on property_media (property_id, category, sort_order);

-- ─────────────────────────────────────────────────────────────────────────
-- Projects
-- ─────────────────────────────────────────────────────────────────────────

create table projects (
  id              bigint generated always as identity primary key,
  name            text not null,
  tagline         text not null default '',
  developer       text not null default '',
  city            citext not null references cities(name) on update cascade,
  area_id         bigint not null references areas(id) on delete restrict,
  address         text not null default '',
  status          project_status not null default 'Launching',
  from_price      numeric(14, 0) not null check (from_price >= 0),  -- starting price, PKR
  possession      text not null default '',
  size            text not null default '',
  approval        text not null default '',
  noc             text not null default '',
  types           text not null default '',   -- free-text summary, e.g. "Plots · Villas · Apartments"
  down_pct        smallint not null default 20 check (down_pct between 0 and 100),
  plan_years      smallint not null default 3 check (plan_years >= 0),
  agent_id        bigint references agents(id) on delete set null,
  brochure_url    text not null default '',
  plan_pdf_url    text not null default '',
  price_list_url  text not null default '',
  master_pdf_url  text not null default '',
  gate_downloads  boolean not null default false,  -- lead-gates the PDF downloads
  description     text not null default '',   -- rich text (HTML)
  dev_about       text not null default '',   -- rich text (HTML)
  dev_logo_url    text not null default '',
  dev_projects    integer not null default 0,
  dev_years       integer not null default 0,
  dev_families    integer not null default 0,
  price_note      text not null default '',
  progress_date   date,
  youtube         text not null default '',
  maps            text not null default '',
  custom_amenities text not null default '',  -- free-text overflow beyond the controlled vocab
  home            boolean not null default false,
  active          boolean not null default true,
  slug            citext not null unique,
  meta_title      text not null default '',
  meta_desc       text not null default '',
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now()
);
create trigger projects_set_updated_at before update on projects
  for each row execute function set_updated_at();
create index projects_search_idx on projects (city, area_id, status);
create index projects_home_idx on projects (home) where home;

create table project_units (
  id         bigint generated always as identity primary key,
  project_id bigint not null references projects(id) on delete cascade,
  category   text not null,   -- tab on the public project page, e.g. "Residential Plots"
  name       text not null,
  size       text not null,
  price      numeric(14, 0) not null check (price >= 0),
  sort_order integer not null default 0
);
create index project_units_project_idx on project_units (project_id, category, sort_order);

create table project_payment_stages (
  id         bigint generated always as identity primary key,
  project_id bigint not null references projects(id) on delete cascade,
  label      text not null,
  pct        numeric(5, 2) not null check (pct >= 0 and pct <= 100),  -- share of total price
  count      integer not null default 1 check (count >= 1),          -- number of payments in this stage
  sort_order integer not null default 0
);
create index project_payment_stages_project_idx on project_payment_stages (project_id, sort_order);

create table project_floor_plans (
  id            bigint generated always as identity primary key,
  project_id    bigint not null references projects(id) on delete cascade,
  label         text not null,
  covered_area  text not null,
  beds          smallint not null default 0,
  baths         smallint not null default 0,
  balcony       smallint not null default 0,
  price         numeric(14, 0) not null check (price >= 0),
  sort_order    integer not null default 0
);
create index project_floor_plans_project_idx on project_floor_plans (project_id, sort_order);

create table project_progress_stages (
  id         bigint generated always as identity primary key,
  project_id bigint not null references projects(id) on delete cascade,
  label      text not null,
  pct        smallint not null check (pct between 0 and 100),
  note       text not null default '',
  sort_order integer not null default 0
);
create index project_progress_stages_project_idx on project_progress_stages (project_id, sort_order);

create table project_nearby_places (
  id         bigint generated always as identity primary key,
  project_id bigint not null references projects(id) on delete cascade,
  place      text not null,
  minutes    smallint not null check (minutes >= 0),
  sort_order integer not null default 0
);
create index project_nearby_places_project_idx on project_nearby_places (project_id, sort_order);

-- Explicit junction table (types.ts modeled this as its own ProjectAmenityTag
-- interface, unlike Property.features which stayed a plain array).
create table project_amenities (
  project_id bigint not null references projects(id) on delete cascade,
  amenity    text not null,
  primary key (project_id, amenity)
);

create table project_media (
  id         bigint generated always as identity primary key,
  project_id bigint not null references projects(id) on delete cascade,
  category   text not null,   -- Exterior/Amenities/Floor Plans/Master Plan/Construction Progress, or custom
  url        text not null,
  sort_order integer not null default 0
);
create index project_media_project_idx on project_media (project_id, category, sort_order);

-- ─────────────────────────────────────────────────────────────────────────
-- SEO pages (programmatic-SEO engine: one row per area × type × purpose combo)
-- ─────────────────────────────────────────────────────────────────────────

create table seo_pages (
  id          bigint generated always as identity primary key,
  h1          text not null,
  slug        citext not null unique,
  meta_title  text not null default '',   -- supports a `{count}` template token
  meta_desc   text not null default '',
  city        citext not null references cities(name) on update cascade,
  area_id     bigint references areas(id) on delete cascade,  -- null = citywide collection page
  type        seo_page_type not null default 'All',
  purpose     purpose not null,
  min_price   numeric(14, 0),
  max_price   numeric(14, 0),
  tags        seo_page_tag[] not null default '{}',
  intro       text not null default '',   -- rich text (HTML)
  content     text not null default '',   -- rich text (HTML)
  published   boolean not null default false,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now(),
  check (min_price is null or max_price is null or min_price <= max_price)
);
create trigger seo_pages_set_updated_at before update on seo_pages
  for each row execute function set_updated_at();
create index seo_pages_lookup_idx on seo_pages (city, area_id, type, purpose);

-- Manually pinned listings on an SEO page.
create table seo_page_pins (
  seo_page_id bigint not null references seo_pages(id) on delete cascade,
  property_id bigint not null references properties(id) on delete cascade,
  sort_order  integer not null default 0,
  primary key (seo_page_id, property_id)
);

-- ─────────────────────────────────────────────────────────────────────────
-- Blog
-- ─────────────────────────────────────────────────────────────────────────

create table blog_posts (
  id            bigint generated always as identity primary key,
  title         text not null,
  category      blog_category not null,
  status        blog_post_status not null default 'Draft',
  publish_date  date,
  cover_image_url text not null default '',
  author_id     bigint references agents(id) on delete set null,
  area_id       bigint references areas(id) on delete set null,
  excerpt       text not null default '',
  body          text not null default '',   -- rich text (HTML)
  slug          citext not null unique,
  meta_title    text not null default '',
  meta_desc     text not null default '',
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
create trigger blog_posts_set_updated_at before update on blog_posts
  for each row execute function set_updated_at();
create index blog_posts_status_idx on blog_posts (status, publish_date desc);
create index blog_posts_category_idx on blog_posts (category);

create table blog_post_sections (
  id         bigint generated always as identity primary key,
  post_id    bigint not null references blog_posts(id) on delete cascade,
  heading    text not null,
  body       text not null,
  sort_order integer not null default 0
);
create index blog_post_sections_post_idx on blog_post_sections (post_id, sort_order);

-- ─────────────────────────────────────────────────────────────────────────
-- FAQs (polymorphic — reused by projects/seo_pages/areas/blog_posts/home/overseas)
-- ─────────────────────────────────────────────────────────────────────────

create table faqs (
  id          bigint generated always as identity primary key,
  entity_type faq_entity_type not null,
  entity_id   bigint,   -- null for the site-wide "home"/"overseas" FAQ sets
  question    text not null,
  answer      text not null,
  sort_order  integer not null default 0,
  check ((entity_type in ('home', 'overseas')) = (entity_id is null))
);
create index faqs_owner_idx on faqs (entity_type, entity_id, sort_order);

-- ─────────────────────────────────────────────────────────────────────────
-- CRM
-- ─────────────────────────────────────────────────────────────────────────

create table leads (
  id          bigint generated always as identity primary key,
  name        text not null,
  phone       text not null,
  whatsapp    text not null default '',
  email       citext not null default '',
  form        lead_form not null,
  source_page text not null default '',    -- URL the lead was submitted from
  interest    text not null default '',    -- free-text summary of what they want
  message     text not null default '',
  details     jsonb not null default '{}', -- form-specific structured payload
  status      lead_status not null default 'New',
  agent_id    bigint references agents(id) on delete set null,
  notes       text not null default '',    -- internal, never shown publicly
  created_at  timestamptz not null default now()
  -- Leads only ever originate from the public site (no admin "create"); the
  -- admin panel may only update status/agent_id/notes, or delete.
);
create index leads_status_idx on leads (status, created_at desc);
create index leads_form_idx on leads (form);
create index leads_details_gin_idx on leads using gin (details);

-- An ongoing subscription — distinct from a one-off "Saved Search" lead.
create table property_alerts (
  id         bigint generated always as identity primary key,
  phone      text not null,
  city       citext references cities(name) on update cascade,
  area_id    bigint references areas(id) on delete set null,
  type       seo_page_type not null default 'All',
  purpose    purpose not null,
  budget     text not null default '',
  active     boolean not null default true,
  created_at timestamptz not null default now()
);
create index property_alerts_active_idx on property_alerts (active, city);

-- ─────────────────────────────────────────────────────────────────────────
-- Site content & settings
-- ─────────────────────────────────────────────────────────────────────────

create table testimonials (
  id         bigint generated always as identity primary key,
  name       text not null,
  role       text not null default '',
  photo_url  text not null default '',
  quote      text not null,
  sort_order integer not null default 0
);

create table legal_documents (
  key             legal_document_key primary key,
  title           text not null,
  description     text not null default '',
  last_updated_at timestamptz not null default now()
);

create table legal_document_sections (
  id            bigint generated always as identity primary key,
  document_key  legal_document_key not null references legal_documents(key) on delete cascade,
  heading       text not null,
  body          text not null,
  sort_order    integer not null default 0
);
create index legal_document_sections_doc_idx on legal_document_sections (document_key, sort_order);

create table offices (
  id      bigint generated always as identity primary key,
  name    text not null,
  address text not null,
  phone   text not null,
  hours   text not null default '',
  lat     numeric(9, 6),
  lng     numeric(9, 6)
);

-- Singleton row (id is always 1 — enforced below).
create table site_settings (
  id                     boolean primary key default true check (id),
  name                   text not null,
  phone                  text not null,
  whatsapp               text not null,
  email                  citext not null,
  hero_title             text not null default '',
  hero_sub               text not null default '',
  hero_image_url         text not null default '',
  exit_intent_enabled    boolean not null default true,
  whatsapp_float_enabled boolean not null default true,
  updated_at             timestamptz not null default now()
);
create trigger site_settings_set_updated_at before update on site_settings
  for each row execute function set_updated_at();

create table currency_rates (
  code          currency_code primary key,
  rate_to_pkr   numeric(12, 6) not null check (rate_to_pkr > 0),
  updated_at    timestamptz not null default now()
);

-- CMS visibility toggles for the 13 named homepage sections.
create table homepage_sections (
  id         bigint generated always as identity primary key,
  key        text not null unique,
  label      text not null,
  sort_order integer not null default 0,
  is_visible boolean not null default true
);

-- Not present in the design mockup (it assumes a single implicit login);
-- included so the auth seam is real from day one.
create table admin_users (
  id            bigint generated always as identity primary key,
  name          text not null,
  email         citext not null unique,
  password_hash text not null,
  role          admin_role not null default 'editor',
  created_at    timestamptz not null default now()
);


-- =========================================================================
-- Computed views (mirrors admin-panel data-access.ts's computed helpers —
-- these values are deliberately NOT stored columns, to avoid drift)
-- =========================================================================

-- City-level counts (design's CITIES seed hardcoded these; here they're live).
create or replace view v_city_stats as
select
  c.name as city,
  c.is_primary,
  count(distinct p.id) filter (where p.status = 'Published') as property_count,
  count(distinct pr.id) filter (where pr.active) as project_count
from cities c
left join properties p on p.city = c.name
left join projects pr on pr.city = c.name
group by c.name, c.is_primary;

-- Area-level listing counts, for the "Popular areas in {city}" cards.
create or replace view v_area_stats as
select
  a.id as area_id,
  a.name,
  a.city,
  count(p.id) filter (where p.status = 'Published') as listing_count
from areas a
left join properties p on p.area_id = a.id
group by a.id, a.name, a.city;

-- SEO page display status: Indexed / Hidden (< 3 listings) / Unpublished.
-- Never stored — computed from `published` + the live matching-listing count.
create or replace view v_seo_page_status as
select
  sp.id as seo_page_id,
  sp.slug,
  sp.published,
  count(p.id) filter (
    where p.status = 'Published'
      and p.city = sp.city
      and (sp.area_id is null or p.area_id = sp.area_id)
      and (sp.type = 'All' or p.type::text = sp.type::text)
      and p.purpose = sp.purpose
      and (sp.min_price is null or p.price >= sp.min_price)
      and (sp.max_price is null or p.price <= sp.max_price)
  ) as matching_listing_count,
  case
    when not sp.published then 'Unpublished'
    when count(p.id) filter (
      where p.status = 'Published'
        and p.city = sp.city
        and (sp.area_id is null or p.area_id = sp.area_id)
        and (sp.type = 'All' or p.type::text = sp.type::text)
        and p.purpose = sp.purpose
        and (sp.min_price is null or p.price >= sp.min_price)
        and (sp.max_price is null or p.price <= sp.max_price)
    ) < 3 then 'Hidden (< 3 listings)'
    else 'Indexed'
  end as display_status
from seo_pages sp
left join properties p on p.city = sp.city
group by sp.id, sp.slug, sp.published;
