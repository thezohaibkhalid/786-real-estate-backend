-- =========================================================================
-- 0003_seed_properties.sql
-- Seed Areas, Agents, Agent Areas, Properties, and Property Media
-- =========================================================================

-- Ensure unique index on agents(email) for idempotency
create unique index if not exists agents_email_idx on agents (email);

-- ─────────────────────────────────────────────────────────────────────────
-- 1. Seed Areas
-- ─────────────────────────────────────────────────────────────────────────

insert into areas (name, city, cover_image_url, h1, meta_title, meta_desc, description) values
  -- Faisalabad Areas
  ('Canal Road', 'Faisalabad', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80',
   'Canal Road Faisalabad: Properties, Houses & Plots',
   'Properties in Canal Road Faisalabad | 786 Real Estate',
   'Browse verified luxury houses, residential plots, and commercial properties on Canal Road Faisalabad.',
   '<p>Canal Road is Faisalabad’s premier residential and commercial corridor featuring top-tier gated societies, lush surroundings, elite educational campuses, and direct highway access.</p>'),

  ('Madina Town', 'Faisalabad', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80',
   'Madina Town Faisalabad: Properties & Houses for Sale',
   'Properties in Madina Town Faisalabad | 786 Real Estate',
   'Explore prime houses and commercial spaces in Madina Town, one of Faisalabad’s most established central neighborhoods.',
   '<p>Madina Town is an established upscale residential locality in Faisalabad renowned for its spacious parks, renowned shopping markets, schools, and central connectivity.</p>'),

  ('Susan Road', 'Faisalabad', 'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?auto=format&fit=crop&w=1200&q=80',
   'Susan Road Faisalabad: Apartments, Commercial & Houses',
   'Properties on Susan Road Faisalabad | 786 Real Estate',
   'Discover luxury apartments, executive commercial offices, and houses on busy Susan Road Faisalabad.',
   '<p>Susan Road is a bustling commercial and urban living hub in Faisalabad with high rental demand, trendy cafes, corporate offices, and modern apartment developments.</p>'),

  ('Wapda City', 'Faisalabad', 'https://images.unsplash.com/photo-1570129477492-45c003edd2be?auto=format&fit=crop&w=1200&q=80',
   'Wapda City Faisalabad: Houses & Plots for Sale',
   'Properties in Wapda City Faisalabad | 786 Real Estate',
   'Find brand new modern homes and ready-to-build residential plots in Wapda City Faisalabad with uninterrupted utilities.',
   '<p>Wapda City is a modern master-planned gated housing community along the Canal Expressway, offering secure family living, green parks, and high capital growth.</p>'),

  ('Citi Housing', 'Faisalabad', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80',
   'Citi Housing Faisalabad: Gold Standard Living',
   'Properties in Citi Housing Faisalabad | 786 Real Estate',
   'Verified plots and contemporary houses in Citi Housing Faisalabad Phase 1 & 2.',
   '<p>Citi Housing Faisalabad represents gold-standard lifestyle community living with underground utilities, dancing fountains, international standard parks, and 24/7 security.</p>'),

  ('Eden Valley', 'Faisalabad', 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?auto=format&fit=crop&w=1200&q=80',
   'Eden Valley Faisalabad: Prime Commercial & Residential',
   'Properties in Eden Valley Faisalabad | 786 Real Estate',
   'Prime commercial plots, plazas, and elegant houses in Eden Valley on Canal Expressway.',
   '<p>Eden Valley is an FDA-approved community located directly on the Main Canal Road with a high-growth commercial boulevard and peaceful residential sectors.</p>'),

  ('Peoples Colony', 'Faisalabad', 'https://images.unsplash.com/photo-1564013799919-ab600027ffc6?auto=format&fit=crop&w=1200&q=80',
   'Peoples Colony Faisalabad: Historic & Central Residences',
   'Properties in Peoples Colony Faisalabad | 786 Real Estate',
   'Houses for sale and rent in Peoples Colony No. 1 and No. 2 Faisalabad.',
   '<p>Peoples Colony is one of Faisalabad’s most prestigious central sectors, known for large family homes, D-Ground market access, and top healthcare centers.</p>'),

  ('Gulberg', 'Faisalabad', 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?auto=format&fit=crop&w=1200&q=80',
   'Gulberg Faisalabad: Central Residential & Commercial',
   'Properties in Gulberg Faisalabad | 786 Real Estate',
   'Affordable and mid-range properties for sale and rent in Gulberg Faisalabad.',
   '<p>Centrally located with vibrant markets, schools, and quick transit to all major city sectors.</p>'),

  ('Kohinoor City', 'Faisalabad', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80',
   'Kohinoor City Faisalabad: Premier Commercial District',
   'Properties in Kohinoor City Faisalabad | 786 Real Estate',
   'Executive office spaces, commercial showrooms, and luxury suites in Kohinoor City.',
   '<p>Kohinoor City is Faisalabad’s premier corporate and retail center, home to major banks, corporate headquarters, and high-end restaurants.</p>'),

  ('Jhang Road', 'Faisalabad', 'https://images.unsplash.com/photo-1568605114967-8130f3a36994?auto=format&fit=crop&w=1200&q=80',
   'Jhang Road Faisalabad: Residential & Affordable Living',
   'Properties on Jhang Road Faisalabad | 786 Real Estate',
   'Family houses and commercial plots with quick highway access on Jhang Road.',
   '<p>A key transit artery of Faisalabad featuring thriving industrial zones, suburban housing schemes, and accessible public transport.</p>'),

  -- Lahore Areas
  ('DHA Phase 6', 'Lahore', 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?auto=format&fit=crop&w=1200&q=80',
   'DHA Phase 6 Lahore: Luxury Homes & Prime Plots',
   'Properties in DHA Phase 6 Lahore | 786 Real Estate',
   'Exclusive designer houses and investment plots in DHA Phase 6 Lahore.',
   '<p>DHA Phase 6 represents the pinnacle of luxury living in Lahore with wide boulevards, elite commercial hubs, and world-class sports clubs.</p>'),

  ('Bahria Town', 'Lahore', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80',
   'Bahria Town Lahore: World-Class Community Living',
   'Properties in Bahria Town Lahore | 786 Real Estate',
   'Affordable luxury homes, apartments, and plots near the Eiffel Tower and Grand Jamia Masjid.',
   '<p>A master-planned private city with uninterrupted electricity, themed monuments, international hospital networks, and 24/7 security.</p>'),

  -- Islamabad Areas
  ('E-11', 'Islamabad', 'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&w=1200&q=80',
   'Sector E-11 Islamabad: Luxury Apartments & Hill Views',
   'Properties in E-11 Islamabad | 786 Real Estate',
   'Scenic Margalla-facing apartments and penthouses in Sector E-11 Islamabad.',
   '<p>Sector E-11 is perched beneath the scenic Margalla Hills, popular among executives, expatriates, and investors for luxury high-rise living.</p>')

on conflict (city, name) do update set
  cover_image_url = excluded.cover_image_url,
  h1 = excluded.h1,
  meta_title = excluded.meta_title,
  meta_desc = excluded.meta_desc,
  description = excluded.description,
  updated_at = now();

-- ─────────────────────────────────────────────────────────────────────────
-- 2. Seed Agents
-- ─────────────────────────────────────────────────────────────────────────

insert into agents (name, role, phone, whatsapp, email, exp, deals, photo_url, bio, active) values
  ('Tariq Mehmood', 'Residential Sales, Canal Road', '+923000000001', '923000000001', 'tariq@786realestate.example',
   12, 240, 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&w=600&q=80',
   '<p>Senior sales advisor specializing in luxury residential properties along Canal Road and Madina Town with over 12 years of market experience.</p>', true),

  ('Hamza Malik', 'Project Bookings & Societies', '+923000000002', '923000000002', 'hamza@786realestate.example',
   8, 160, 'https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?auto=format&fit=crop&w=600&q=80',
   '<p>Dedicated consultant for master-planned communities, FDA-approved societies, and flexible payment plan projects.</p>', true),

  ('Zainab Bibi', 'Rentals & Luxury Suites Specialist', '+923000000003', '923000000003', 'zainab@786realestate.example',
   6, 110, 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?auto=format&fit=crop&w=600&q=80',
   '<p>Specialist in executive furnished apartments, luxury rental houses, and tenant vetting across Faisalabad.</p>', true),

  ('Bilal Ahmed', 'Commercial & Plaza Property', '+923000000004', '923000000004', 'bilal@786realestate.example',
   14, 310, 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&w=600&q=80',
   '<p>Commercial property veteran handling corporate offices, high-street retail shops, and commercial plots with proven ROI.</p>', true),

  ('Usman Farooq', 'Overseas Clients Desk (UK & Gulf)', '+923000000005', '923000000005', 'usman@786realestate.example',
   9, 195, 'https://images.unsplash.com/photo-1580489944761-15a19d654956?auto=format&fit=crop&w=600&q=80',
   '<p>Head of Overseas Desk providing remote video tours, secure title verification, and POA transfer services for Pakistani expatriates.</p>', true),

  ('Ahmed Raza', 'Plots, Files & Society Verification', '+923000000006', '923000000006', 'ahmed@786realestate.example',
   7, 145, 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?auto=format&fit=crop&w=600&q=80',
   '<p>Land specialist focusing on on-ground residential and commercial plot files, NOC clearance, and town-planning compliance.</p>', true)

on conflict (email) do update set
  name = excluded.name,
  role = excluded.role,
  phone = excluded.phone,
  whatsapp = excluded.whatsapp,
  exp = excluded.exp,
  deals = excluded.deals,
  photo_url = excluded.photo_url,
  bio = excluded.bio,
  active = excluded.active,
  updated_at = now();

-- ─────────────────────────────────────────────────────────────────────────
-- 3. Seed Agent Areas (M:M Junction)
-- ─────────────────────────────────────────────────────────────────────────

insert into agent_areas (agent_id, area_id)
select a.id, ar.id
from agents a, areas ar
where (a.name = 'Tariq Mehmood' and ar.city = 'Faisalabad' and ar.name in ('Canal Road', 'Madina Town', 'Susan Road'))
   or (a.name = 'Hamza Malik' and ar.city = 'Faisalabad' and ar.name in ('Wapda City', 'Citi Housing', 'Eden Valley'))
   or (a.name = 'Zainab Bibi' and ar.city = 'Faisalabad' and ar.name in ('Susan Road', 'Peoples Colony', 'Kohinoor City', 'Jhang Road'))
   or (a.name = 'Bilal Ahmed' and ar.city = 'Faisalabad' and ar.name in ('Eden Valley', 'Kohinoor City', 'Gulberg'))
   or (a.name = 'Usman Farooq' and ((ar.city = 'Lahore' and ar.name = 'DHA Phase 6') or (ar.city = 'Islamabad' and ar.name = 'E-11')))
   or (a.name = 'Ahmed Raza' and ((ar.city = 'Faisalabad' and ar.name in ('Wapda City', 'Citi Housing')) or (ar.city = 'Lahore' and ar.name = 'Bahria Town')))
on conflict do nothing;

-- ─────────────────────────────────────────────────────────────────────────
-- 4. Seed Properties
-- ─────────────────────────────────────────────────────────────────────────

-- Property 1: 10 Marla Modern House (Canal Road)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '10 Marla Modern House',
  '10-marla-modern-house-canal-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
  'House', 'sale', 38500000, 10, 'Marla',
  5, 6, 2, 2, 'Double', 'Semi-furnished', 'FDA',
  array['Solar system', 'Gated community', 'Lawn', 'Drawing room', 'TV lounge', 'Electricity', 'Gas']::property_feature[],
  'Brand new 10 Marla architectural masterpiece situated on prime Canal Road. Built with imported Spanish porcelain tiles, bespoke Turkish fittings, and solid ash-wood doors. Features two designer kitchens with Italian appliances, double-height lobby, and private terrace garden overlooking landscaped greenery.',
  'Published', true, 1,
  '10 Marla Modern House for Sale on Canal Road Faisalabad',
  'Brand new 10 Marla luxury home on Canal Road Faisalabad. 5 bed, 6 bath, 2 kitchens, solar ready.',
  (select id from agents where name = 'Tariq Mehmood')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 2: 1 Kanal Corner House (Madina Town)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '1 Kanal Corner House',
  '1-kanal-corner-house-madina-town',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Madina Town'),
  'House', 'sale', 72000000, 1, 'Kanal',
  6, 7, 2, 3, 'Double', 'Unfurnished', 'FDA',
  array['Corner', 'Park facing', 'Lawn', 'Drawing room', 'TV lounge', 'Security guards', 'Gas', 'Electricity']::property_feature[],
  'Exclusive 1 Kanal corner residence in the prestigious heart of Madina Town. Features expansive 60-foot front facing lush public park, separate servant quarters, basement home theater space, and premium Italian marble throughout.',
  'Published', true, 2,
  '1 Kanal Corner Luxury House for Sale in Madina Town Faisalabad',
  'Exclusive 1 Kanal corner house in Madina Town Faisalabad facing central park with servant quarters.',
  (select id from agents where name = 'Tariq Mehmood')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 3: 5 Marla Brand New House (Wapda City)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '5 Marla Brand New House',
  '5-marla-brand-new-house-wapda-city',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Wapda City'),
  'House', 'sale', 17500000, 5, 'Marla',
  3, 4, 1, 1, 'Double', 'Unfurnished', 'FDA',
  array['Gas', 'Electricity', 'Near masjid', 'Near park', 'Drawing room', 'TV lounge']::property_feature[],
  'Priced to sell quickly: newly constructed 5 Marla contemporary double-story home in Wapda City. Excellent layout with spacious bedrooms, attached designer baths, powder room, and dedicated laundry area.',
  'Published', false, 0,
  '5 Marla Brand New House for Sale in Wapda City Faisalabad',
  'Contemporary 5 Marla double-story home in Wapda City Faisalabad with gas, electricity, and attached baths.',
  (select id from agents where name = 'Hamza Malik')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 4: 2 Bed Apartment (Susan Road)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '2 Bed Apartment',
  '2-bed-apartment-susan-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Susan Road'),
  'Apartment', 'rent', 85000, 1150, 'Sq. Ft.',
  2, 2, 1, 1, '4th floor', 'Furnished', 'FDA',
  array['Balcony', 'Gated community', 'Security guards', 'CCTV', 'Electricity']::property_feature[],
  'Fully furnished luxury executive apartment located on Susan Road. Equipped with inverter air conditioning, modern sofa set, LED TV, king-size orthopedic beds, and high-speed fiber internet connection.',
  'Published', false, 0,
  'Luxury 2 Bed Furnished Apartment for Rent on Susan Road Faisalabad',
  'Move-in ready 2 bed furnished apartment on Susan Road Faisalabad with elevator, backup generator, and CCTV.',
  (select id from agents where name = 'Zainab Bibi')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 5: 10 Marla Residential Plot (Citi Housing)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '10 Marla Residential Plot',
  '10-marla-residential-plot-citi-housing',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Citi Housing'),
  'Plot', 'sale', 9800000, 10, 'Marla',
  0, 0, 0, 0, '', 'Unfurnished', 'FDA',
  array['Gated community', 'Electricity', 'Gas', 'Security guards']::property_feature[],
  'Ready for immediate construction: 10 Marla residential plot in Citi Housing Phase 1, Block B. Completely clear possession, direct transfer, and located adjacent to the international standard commercial center and school.',
  'Published', false, 0,
  '10 Marla Residential Plot for Sale in Citi Housing Faisalabad',
  'Ready-to-build 10 Marla plot in Citi Housing Phase 1 Faisalabad with underground electricity.',
  (select id from agents where name = 'Ahmed Raza')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 6: 8 Marla Commercial Plot (Eden Valley)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '8 Marla Commercial Plot',
  '8-marla-commercial-plot-eden-valley',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Eden Valley'),
  'Commercial', 'sale', 45000000, 8, 'Marla',
  0, 0, 0, 0, '', 'Unfurnished', 'FDA',
  array['Corner', 'Electricity', 'CCTV', 'Security guards']::property_feature[],
  'High ROI investment opportunity: 8 Marla commercial plot on the 100-foot Main Boulevard of Eden Valley. Approved for multi-story plaza construction (Basement + Ground + 5 Floors) with guaranteed rental yields.',
  'Published', true, 3,
  '8 Marla Main Boulevard Commercial Plot in Eden Valley Faisalabad',
  'Prime 8 Marla commercial plot on 100ft boulevard in Eden Valley Faisalabad. Multi-story plaza approved.',
  (select id from agents where name = 'Bilal Ahmed')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 7: 1 Kanal Upper Portion (Peoples Colony)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '1 Kanal Upper Portion',
  '1-kanal-upper-portion-peoples-colony',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Peoples Colony'),
  'House', 'rent', 120000, 1, 'Kanal',
  3, 3, 1, 2, 'Upper portion', 'Unfurnished', 'FDA',
  array['Terrace', 'Balcony', 'Near park', 'Near masjid', 'Electricity', 'Gas']::property_feature[],
  'Spacious and airy 1 Kanal upper portion with completely independent entrance and private staircase. Features 3 master bedrooms, expansive living hall, separate dining room, modern kitchen, and scenic open rooftop.',
  'Published', false, 0,
  '1 Kanal Upper Portion for Rent in Peoples Colony Faisalabad',
  'Independent 1 Kanal upper portion with private entrance in Peoples Colony Faisalabad.',
  (select id from agents where name = 'Zainab Bibi')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 8: 10 Marla House (DHA Phase 6 Lahore)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '10 Marla House',
  '10-marla-house-dha-phase-6-lahore',
  'Lahore',
  (select id from areas where city = 'Lahore' and name = 'DHA Phase 6'),
  'House', 'sale', 65000000, 10, 'Marla',
  5, 6, 2, 2, 'Double', 'Unfurnished', 'LDA',
  array['Solar system', 'Gated community', 'Security guards', 'Lawn', 'Drawing room', 'TV lounge']::property_feature[],
  'Ultra-modern designer 10 Marla home in Sector C, DHA Phase 6 Lahore. Featuring striking double-glazed glass elevation, Grohe sanitary fittings, full solar array, and expansive basement lounge.',
  'Published', false, 0,
  '10 Marla Designer House for Sale in DHA Phase 6 Lahore',
  'Brand new 10 Marla modern house in DHA Phase 6 Lahore with solar system and basement.',
  (select id from agents where name = 'Usman Farooq')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 9: 1 Kanal Residential Plot (Bahria Town Lahore)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '1 Kanal Residential Plot',
  '1-kanal-residential-plot-bahria-town-lahore',
  'Lahore',
  (select id from areas where city = 'Lahore' and name = 'Bahria Town'),
  'Plot', 'sale', 28000000, 1, 'Kanal',
  0, 0, 0, 0, '', 'Unfurnished', 'LDA',
  array['Near masjid', 'Near park', 'Electricity', 'Gas', 'Gated community']::property_feature[],
  'Prime 1 Kanal residential plot in Sector F, Bahria Town Lahore. Level plot with zero filling required, located within walking distance from Eiffel Tower park and Grand Jamia Masjid.',
  'Published', false, 0,
  '1 Kanal Residential Plot for Sale in Bahria Town Lahore',
  'Prime 1 Kanal plot in Sector F Bahria Town Lahore near Grand Jamia Masjid and Eiffel Tower.',
  (select id from agents where name = 'Ahmed Raza')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 10: 3 Bed Apartment (E-11 Islamabad)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '3 Bed Apartment',
  '3-bed-apartment-e-11-islamabad',
  'Islamabad',
  (select id from areas where city = 'Islamabad' and name = 'E-11'),
  'Apartment', 'sale', 42000000, 1850, 'Sq. Ft.',
  3, 3, 1, 1, '7th floor', 'Semi-furnished', 'CDA',
  array['Balcony', 'Gated community', 'Security guards', 'CCTV', 'Electricity', 'Gas']::property_feature[],
  'Breathtaking Margalla Hills panoramic view from the 7th floor in Sector E-11 Islamabad. Featuring expansive balconies, open-concept island kitchen, high-speed passenger & cargo lifts, and 24/7 security surveillance.',
  'Published', false, 0,
  '3 Bed Luxury Apartment for Sale in Sector E-11 Islamabad',
  'Panoramic Margalla Hills view 3 bed luxury apartment in E-11 Islamabad with 24/7 security.',
  (select id from agents where name = 'Usman Farooq')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 11: Office Space (Kohinoor City)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  'Office Space',
  'office-space-kohinoor-city',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Kohinoor City'),
  'Commercial', 'rent', 150000, 900, 'Sq. Ft.',
  0, 2, 0, 1, '', 'Furnished', 'FDA',
  array['CCTV', 'Security guards', 'Electricity']::property_feature[],
  'Turnkey executive office space in Kohinoor City commercial center. Pre-fitted with modern glass cubicles, director executive room, conference room, reception counter, and dedicated server room wiring.',
  'Published', false, 0,
  'Turnkey Executive Office Space for Rent in Kohinoor City Faisalabad',
  'Ready office space with glass cabins and conference room in Kohinoor City Faisalabad.',
  (select id from agents where name = 'Bilal Ahmed')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 12: 5 Marla House (Jhang Road)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '5 Marla House',
  '5-marla-house-jhang-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Jhang Road'),
  'House', 'rent', 55000, 5, 'Marla',
  3, 3, 1, 1, 'Double', 'Unfurnished', 'FDA',
  array['Electricity', 'Gas', 'Near masjid', 'Near park', 'Drawing room']::property_feature[],
  'Clean and well-kept 5 Marla double-story home on Jhang Road for rent. Family-friendly quiet neighborhood with nearby schools, shopping bazaar, and quick access to public transport.',
  'Published', false, 0,
  '5 Marla Double Story House for Rent on Jhang Road Faisalabad',
  'Affordable 5 Marla rental house on Jhang Road Faisalabad with car porch and marble flooring.',
  (select id from agents where name = 'Zainab Bibi')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 13: 5 Marla Designer Home (Canal Road)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '5 Marla Designer Home',
  '5-marla-designer-home-canal-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
  'House', 'sale', 21500000, 5, 'Marla',
  3, 4, 1, 1, 'Double', 'Semi-furnished', 'FDA',
  array['Solar system', 'Gated community', 'Balcony', 'Drawing room', 'TV lounge', 'Electricity', 'Gas']::property_feature[],
  'Chic 5 Marla contemporary double-story home situated inside an exclusive gated community on Canal Road. Features Spanish tiling, false ceilings with ambient LED cove lights, and a cozy private terrace.',
  'Published', true, 4,
  '5 Marla Designer House for Sale on Canal Road Faisalabad',
  'Stylish 5 Marla modern home in secure gated society on Canal Road Faisalabad.',
  (select id from agents where name = 'Tariq Mehmood')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 14: 1 Kanal Luxury Villa (Canal Road - Rent)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '1 Kanal Luxury Villa',
  '1-kanal-luxury-villa-canal-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
  'House', 'rent', 220000, 1, 'Kanal',
  5, 6, 2, 3, 'Double', 'Furnished', 'FDA',
  array['Lawn', 'Terrace', 'Drawing room', 'TV lounge', 'Solar system', 'Security guards', 'Electricity', 'Gas']::property_feature[],
  'Magnificent 1 Kanal furnished villa for corporate or family lease along Canal Road. Features sprawling manicured lawn, separate servant quarter, 15kVA solar backup, and imported designer furniture.',
  'Published', false, 0,
  '1 Kanal Fully Furnished Luxury Villa for Rent on Canal Road Faisalabad',
  'Executive 1 Kanal furnished residence on Canal Road Faisalabad with lawn and solar backup.',
  (select id from agents where name = 'Zainab Bibi')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 15: 10 Marla Residential Plot (Canal Road)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '10 Marla Residential Plot',
  '10-marla-residential-plot-canal-road',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
  'Plot', 'sale', 16500000, 10, 'Marla',
  0, 0, 0, 0, '', 'Unfurnished', 'FDA',
  array['Gated community', 'Electricity', 'Gas', 'Near park', 'Near masjid']::property_feature[],
  'Prime 10 Marla residential plot located on a 50-foot wide street in an FDA-approved society off Canal Road. Clear title, possession available, ready for immediate house construction.',
  'Published', false, 0,
  '10 Marla Plot for Sale on Canal Road Faisalabad',
  'Possession ready 10 Marla plot on 50ft road off Canal Road Faisalabad.',
  (select id from agents where name = 'Hamza Malik')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- Property 16: 10 Marla Spanish House (Madina Town)
insert into properties (
  title, slug, city, area_id, type, purpose, price, size, unit,
  beds, baths, kitchens, parking, floors, furnished, approved,
  features, description, status, home, home_sort_order, meta_title, meta_desc, agent_id
) values (
  '10 Marla Spanish House',
  '10-marla-spanish-house-madina-town',
  'Faisalabad',
  (select id from areas where city = 'Faisalabad' and name = 'Madina Town'),
  'House', 'sale', 48000000, 10, 'Marla',
  4, 5, 2, 2, 'Double', 'Semi-furnished', 'FDA',
  array['Terrace', 'Balcony', 'Drawing room', 'TV lounge', 'Lawn', 'Electricity', 'Gas']::property_feature[],
  'Authentic Spanish colonial architecture in prime sector of Madina Town. Curved archways, clay roof tiles, solid deodar woodwork, imported sanitary ware, and manicured side lawn.',
  'Published', true, 5,
  '10 Marla Spanish Villa for Sale in Madina Town Faisalabad',
  'Luxury Spanish style 10 Marla house in Madina Town Faisalabad with 4 beds and servant room.',
  (select id from agents where name = 'Tariq Mehmood')
)
on conflict (slug) do update set
  title = excluded.title, price = excluded.price, size = excluded.size, unit = excluded.unit,
  beds = excluded.beds, baths = excluded.baths, kitchens = excluded.kitchens, parking = excluded.parking,
  floors = excluded.floors, furnished = excluded.furnished, approved = excluded.approved,
  features = excluded.features, description = excluded.description, status = excluded.status,
  home = excluded.home, home_sort_order = excluded.home_sort_order, agent_id = excluded.agent_id,
  updated_at = now();

-- ─────────────────────────────────────────────────────────────────────────
-- 5. Seed Property Media
-- ─────────────────────────────────────────────────────────────────────────

-- Clean existing media for seeded properties so we don't duplicate on re-run
delete from property_media where property_id in (
  select id from properties where slug in (
    '10-marla-modern-house-canal-road',
    '1-kanal-corner-house-madina-town',
    '5-marla-brand-new-house-wapda-city',
    '2-bed-apartment-susan-road',
    '10-marla-residential-plot-citi-housing',
    '8-marla-commercial-plot-eden-valley',
    '1-kanal-upper-portion-peoples-colony',
    '10-marla-house-dha-phase-6-lahore',
    '1-kanal-residential-plot-bahria-town-lahore',
    '3-bed-apartment-e-11-islamabad',
    'office-space-kohinoor-city',
    '5-marla-house-jhang-road',
    '5-marla-designer-home-canal-road',
    '1-kanal-luxury-villa-canal-road',
    '10-marla-residential-plot-canal-road',
    '10-marla-spanish-house-madina-town'
  )
);

-- Media for 10 Marla Modern House (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 1),
  ('Exterior', 'https://images.unsplash.com/photo-1583608205776-bfd35f0d9f83?auto=format&fit=crop&w=1200&q=80', 2),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 3),
  ('Bedroom', 'https://images.unsplash.com/photo-1493809842364-78817add7ffb?auto=format&fit=crop&w=1200&q=80', 4),
  ('Kitchen', 'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order)
where p.slug = '10-marla-modern-house-canal-road';

-- Media for 1 Kanal Corner House (Madina Town)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 1),
  ('Exterior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 3),
  ('Bedroom', 'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 4)
) as m(category, url, sort_order)
where p.slug = '1-kanal-corner-house-madina-town';

-- Media for 5 Marla Brand New House (Wapda City)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1570129477492-45c003edd2be?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '5-marla-brand-new-house-wapda-city';

-- Media for 2 Bed Apartment (Susan Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '2-bed-apartment-susan-road';

-- Media for 10 Marla Plot (Citi Housing)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 1)
) as m(category, url, sort_order)
where p.slug = '10-marla-residential-plot-citi-housing';

-- Media for 8 Marla Commercial Plot (Eden Valley)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?auto=format&fit=crop&w=1200&q=80', 1)
) as m(category, url, sort_order)
where p.slug = '8-marla-commercial-plot-eden-valley';

-- Media for 1 Kanal Upper Portion (Peoples Colony)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1564013799919-ab600027ffc6?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '1-kanal-upper-portion-peoples-colony';

-- Media for 10 Marla House (DHA Phase 6 Lahore)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '10-marla-house-dha-phase-6-lahore';

-- Media for 1 Kanal Plot (Bahria Town Lahore)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80', 1)
) as m(category, url, sort_order)
where p.slug = '1-kanal-residential-plot-bahria-town-lahore';

-- Media for 3 Bed Apartment (E-11 Islamabad)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '3-bed-apartment-e-11-islamabad';

-- Media for Office Space (Kohinoor City)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1497366811353-6870744d04b2?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = 'office-space-kohinoor-city';

-- Media for 5 Marla House (Jhang Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1568605114967-8130f3a36994?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '5-marla-house-jhang-road';

-- Media for 5 Marla Designer Home (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '5-marla-designer-home-canal-road';

-- Media for 1 Kanal Luxury Villa (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600566753376-12c8ab7fb75b?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '1-kanal-luxury-villa-canal-road';

-- Media for 10 Marla Plot (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 1)
) as m(category, url, sort_order)
where p.slug = '10-marla-residential-plot-canal-road';

-- Media for 10 Marla Spanish House (Madina Town)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p
cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order)
where p.slug = '10-marla-spanish-house-madina-town';
