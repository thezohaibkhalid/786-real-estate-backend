-- =========================================================================
-- 0004_seed_complete_catalog.sql
-- Seed complete photos for all properties, full projects, blogs,
-- testimonials, offices, FAQs, and programmatic SEO pages.
-- =========================================================================

-- ─────────────────────────────────────────────────────────────────────────
-- 1. Full Gallery Media for All Properties
-- ─────────────────────────────────────────────────────────────────────────

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

-- Property 1: 10 Marla Modern House (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 1),
  ('Exterior', 'https://images.unsplash.com/photo-1583608205776-bfd35f0d9f83?auto=format&fit=crop&w=1200&q=80', 2),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 3),
  ('Interior', 'https://images.unsplash.com/photo-1493809842364-78817add7ffb?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 5),
  ('Bedroom',  'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 6),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 7),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 8)
) as m(category, url, sort_order) where p.slug = '10-marla-modern-house-canal-road';

-- Property 2: 1 Kanal Corner House (Madina Town)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 1),
  ('Exterior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Interior', 'https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80', 3),
  ('Bedroom',  'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 4),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 5),
  ('Bathroom', 'https://images.unsplash.com/photo-1584622650111-993a426fbf0a?auto=format&fit=crop&w=1200&q=80', 6),
  ('Lawn',     'https://images.unsplash.com/photo-1558997519-83ea9252def8?auto=format&fit=crop&w=1200&q=80', 7)
) as m(category, url, sort_order) where p.slug = '1-kanal-corner-house-madina-town';

-- Property 3: 5 Marla Brand New House (Wapda City)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1570129477492-45c003edd2be?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '5-marla-brand-new-house-wapda-city';

-- Property 4: 2 Bed Apartment (Susan Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '2-bed-apartment-susan-road';

-- Property 5: 10 Marla Plot (Citi Housing)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 1),
  ('Location', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = '10-marla-residential-plot-citi-housing';

-- Property 6: 8 Marla Commercial Plot (Eden Valley)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?auto=format&fit=crop&w=1200&q=80', 1),
  ('Location', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = '8-marla-commercial-plot-eden-valley';

-- Property 7: 1 Kanal Upper Portion (Peoples Colony)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1564013799919-ab600027ffc6?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5),
  ('Rooftop',  'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 6)
) as m(category, url, sort_order) where p.slug = '1-kanal-upper-portion-peoples-colony';

-- Property 8: 10 Marla House (DHA Phase 6 Lahore)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '10-marla-house-dha-phase-6-lahore';

-- Property 9: 1 Kanal Plot (Bahria Town Lahore)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80', 1),
  ('Location', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = '1-kanal-residential-plot-bahria-town-lahore';

-- Property 10: 3 Bed Apartment (E-11 Islamabad)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '3-bed-apartment-e-11-islamabad';

-- Property 11: Office Space (Kohinoor City)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1497366811353-6870744d04b2?auto=format&fit=crop&w=1200&q=80', 2),
  ('Conference', 'https://images.unsplash.com/photo-1431540015161-0bf868a2d407?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = 'office-space-kohinoor-city';

-- Property 12: 5 Marla House (Jhang Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1568605114967-8130f3a36994?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '5-marla-house-jhang-road';

-- Property 13: 5 Marla Designer Home (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '5-marla-designer-home-canal-road';

-- Property 14: 1 Kanal Luxury Villa (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600566753376-12c8ab7fb75b?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5),
  ('Lawn',     'https://images.unsplash.com/photo-1558997519-83ea9252def8?auto=format&fit=crop&w=1200&q=80', 6)
) as m(category, url, sort_order) where p.slug = '1-kanal-luxury-villa-canal-road';

-- Property 15: 10 Marla Plot (Canal Road)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 1),
  ('Location', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = '10-marla-residential-plot-canal-road';

-- Property 16: 10 Marla Spanish House (Madina Town)
insert into property_media (property_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from properties p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Bedroom',  'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80', 3),
  ('Kitchen',  'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 4),
  ('Bathroom', 'https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?auto=format&fit=crop&w=1200&q=80', 5)
) as m(category, url, sort_order) where p.slug = '10-marla-spanish-house-madina-town';

-- ─────────────────────────────────────────────────────────────────────────
-- 2. Seed Projects & Related Child Records
-- ─────────────────────────────────────────────────────────────────────────

insert into projects (
  name, slug, developer, city, area_id, status, from_price, types,
  possession, size, approval, description, home, active, meta_title, meta_desc, agent_id
) values
  ('Canal Vista Residencia', 'canal-vista-residencia', 'Vista Group & Developers', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
   'Under Construction', 4500000, 'Plots · Villas · Apartments', 'Dec 2028', '850 Kanal', 'FDA Approved',
   'Canal Vista Residencia is Faisalabad''s flagship waterfront master-planned community along West Canal Road. Spanning 850 Kanals of lush gated enclave, it offers underground cabling, private water filtration, grand central mosque, recreational golf greens, and world-class retail boulevard.',
   true, true,
   'Canal Vista Residencia Faisalabad: Plots & Apartments on Installments',
   'FDA approved master-planned community on Canal Road Faisalabad. 5, 10 Marla and 1 Kanal plots on 4-year installment plan.',
   (select id from agents where name = 'Hamza Malik')),

  ('Grand Avenue Heights', 'grand-avenue-heights', 'Avenue Builders & Architects', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Susan Road'),
   'Launching', 6800000, 'Apartments · Shops', 'Dec 2029', '22 Floors', 'FDA Approved',
   'Soaring 22 floors above vibrant Susan Road, Grand Avenue Heights sets a new benchmark in vertical luxury living. Features a sky lounge with panoramic city skyline views, infinity temperature-controlled pool, high-speed Otis elevators, and smart home provisions.',
   true, true,
   'Grand Avenue Heights Susan Road Faisalabad: Luxury Apartments & Retail',
   'High-rise 22-floor luxury apartment project on Susan Road Faisalabad with rooftop pool and gym.',
   (select id from agents where name = 'Hamza Malik')),

  ('Citi Executive Enclave', 'citi-executive-enclave', 'Citi Housing Developers', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Citi Housing'),
   'Ready to Move', 11500000, 'Villas · Plots', 'Ready', '250 Kanal', 'FDA Approved',
   'An exclusive low-density enclave situated within Citi Housing Phase 1. Ready-for-possession designer villas and fully developed residential plots with underground utilities, paved roads, and lush central community park.',
   false, true,
   'Citi Executive Enclave Faisalabad: Ready to Move Villas & Plots',
   'Ready to move luxury villas and immediate possession plots in Citi Housing Phase 1 Faisalabad.',
   (select id from agents where name = 'Ahmed Raza')),

  ('Grand Mall & Corporate Tower', 'grand-mall-corporate-tower', 'Eden Builders', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Eden Valley'),
   'Under Construction', 8500000, 'Commercial Shops · Offices', 'June 2028', '14 Floors', 'FDA Approved',
   'Prime commercial epicenter located on the main 100-foot boulevard of Eden Valley. Offering high-yield anchor retail outlets, corporate headquarter floors, and a dedicated rooftop food court.',
   true, true,
   'Grand Mall & Corporate Tower Eden Valley Faisalabad',
   'High ROI commercial shops and corporate offices in Eden Valley Faisalabad on easy quarterly installments.',
   (select id from agents where name = 'Bilal Ahmed')),

  ('Golf View Estates', 'golf-view-estates', 'DHA Developers', 'Lahore',
   (select id from areas where city = 'Lahore' and name = 'DHA Phase 6'),
   'Under Construction', 35000000, 'Luxury Villas · Plots', 'Dec 2027', '500 Kanal', 'LDA Approved',
   'Ultra-luxurious golf course community in DHA Phase 6 Lahore featuring 1 & 2 Kanal fairway-facing villas with smart home automation and private clubhouse membership.',
   false, true,
   'Golf View Estates DHA Phase 6 Lahore: Fairway Villas',
   'Exclusive golf view villas and residential plots in DHA Phase 6 Lahore with modern lifestyle amenities.',
   (select id from agents where name = 'Usman Farooq')),

  ('Kohinoor Heights', 'kohinoor-heights', 'Kohinoor Group', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Kohinoor City'),
   'Ready to Move', 14000000, 'Executive Suites · Retail', 'Immediate', '12 Floors', 'FDA Approved',
   'Ready-to-move executive corporate suites and premium street-facing retail outlets in the bustling heart of Kohinoor City Faisalabad.',
   false, true,
   'Kohinoor Heights Commercial & Executive Suites Faisalabad',
   'Ready to move corporate offices and luxury suites in Kohinoor City Faisalabad with instant rental return.',
   (select id from agents where name = 'Bilal Ahmed'))
on conflict (slug) do update set
  name = excluded.name, developer = excluded.developer, city = excluded.city,
  area_id = excluded.area_id, status = excluded.status, from_price = excluded.from_price,
  types = excluded.types, possession = excluded.possession, size = excluded.size,
  approval = excluded.approval, description = excluded.description, home = excluded.home,
  active = excluded.active, meta_title = excluded.meta_title, meta_desc = excluded.meta_desc,
  agent_id = excluded.agent_id, updated_at = now();

-- Clean and Seed Project Media
delete from project_media where project_id in (select id from projects where slug in (
  'canal-vista-residencia', 'grand-avenue-heights', 'citi-executive-enclave',
  'grand-mall-corporate-tower', 'golf-view-estates', 'kohinoor-heights'
));

-- Project 1 Media: Canal Vista Residencia
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80', 1),
  ('Amenities', 'https://images.unsplash.com/photo-1571896349842-33c89424de2d?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?auto=format&fit=crop&w=1200&q=80', 3),
  ('Construction Progress', 'https://images.unsplash.com/photo-1541888946425-d0fbb18086f6?auto=format&fit=crop&w=1200&q=80', 4)
) as m(category, url, sort_order) where p.slug = 'canal-vista-residencia';

-- Project 2 Media: Grand Avenue Heights
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=1200&q=80', 1),
  ('Amenities', 'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?auto=format&fit=crop&w=1200&q=80', 2),
  ('Floor Plans', 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80', 3),
  ('Construction Progress', 'https://images.unsplash.com/photo-1590381105924-c72589b9ef3f?auto=format&fit=crop&w=1200&q=80', 4)
) as m(category, url, sort_order) where p.slug = 'grand-avenue-heights';

-- Project 3 Media: Citi Executive Enclave
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?auto=format&fit=crop&w=1200&q=80', 1),
  ('Amenities', 'https://images.unsplash.com/photo-1558997519-83ea9252def8?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = 'citi-executive-enclave';

-- Project 4 Media: Grand Mall & Corporate Tower
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80', 2),
  ('Construction Progress', 'https://images.unsplash.com/photo-1541888946425-d0fbb18086f6?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = 'grand-mall-corporate-tower';

-- Project 5 Media: Golf View Estates
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?auto=format&fit=crop&w=1200&q=80', 1),
  ('Amenities', 'https://images.unsplash.com/photo-1583608205776-bfd35f0d9f83?auto=format&fit=crop&w=1200&q=80', 2),
  ('Master Plan', 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80', 3)
) as m(category, url, sort_order) where p.slug = 'golf-view-estates';

-- Project 6 Media: Kohinoor Heights
insert into project_media (project_id, category, url, sort_order)
select p.id, m.category, m.url, m.sort_order
from projects p cross join (values
  ('Exterior', 'https://images.unsplash.com/photo-1497366811353-6870744d04b2?auto=format&fit=crop&w=1200&q=80', 1),
  ('Interior', 'https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=1200&q=80', 2)
) as m(category, url, sort_order) where p.slug = 'kohinoor-heights';

-- Seed Project Units
delete from project_units where project_id in (select id from projects where slug in ('canal-vista-residencia', 'grand-avenue-heights'));
insert into project_units (project_id, category, name, size, price, sort_order)
select p.id, u.category, u.name, u.size, u.price, u.sort_order
from projects p cross join (values
  ('Residential Plots', '5 Marla Plot', '25×45 ft', 4500000::numeric, 1),
  ('Residential Plots', '7 Marla Plot', '30×52 ft', 6200000::numeric, 2),
  ('Residential Plots', '10 Marla Plot', '35×65 ft', 8500000::numeric, 3),
  ('Residential Plots', '1 Kanal Plot', '50×90 ft', 16000000::numeric, 4),
  ('Apartments', '1 Bed Suite', '650 sq. ft.', 5800000::numeric, 5),
  ('Apartments', '2 Bed Luxury', '1,050 sq. ft.', 9200000::numeric, 6),
  ('Commercial', '4 Marla Shop', '20×45 ft', 12000000::numeric, 7)
) as u(category, name, size, price, sort_order)
where p.slug = 'canal-vista-residencia';

insert into project_units (project_id, category, name, size, price, sort_order)
select p.id, u.category, u.name, u.size, u.price, u.sort_order
from projects p cross join (values
  ('Apartments', '1 Bed Executive', '620 sq. ft.', 6800000::numeric, 1),
  ('Apartments', '2 Bed Panoramic', '1,120 sq. ft.', 11500000::numeric, 2),
  ('Apartments', '3 Bed Penthouse', '1,850 sq. ft.', 18000000::numeric, 3),
  ('Commercial', 'Ground Floor Retail', '450 sq. ft.', 14500000::numeric, 4)
) as u(category, name, size, price, sort_order)
where p.slug = 'grand-avenue-heights';

-- Seed Project Amenities
delete from project_amenities where project_id in (select id from projects where slug in ('canal-vista-residencia', 'grand-avenue-heights'));
insert into project_amenities (project_id, amenity)
select p.id, a.amenity
from projects p cross join (values
  ('24/7 Gated Security & CCTV'),
  ('100% Underground Electrification'),
  ('Grand Jamia Mosque'),
  ('40-Kanal Central Theme Park'),
  ('Commercial Promenade & Dining'),
  ('International Standard Daycare')
) as a(amenity) where p.slug = 'canal-vista-residencia';

insert into project_amenities (project_id, amenity)
select p.id, a.amenity
from projects p cross join (values
  ('Rooftop Infinity Pool'),
  ('High-Speed Otis Elevators'),
  ('4-Level Underground Parking'),
  ('24/7 Standby Generator'),
  ('Modern Health Club & Gym')
) as a(amenity) where p.slug = 'grand-avenue-heights';

-- Seed Project Floor Plans
delete from project_floor_plans where project_id in (select id from projects where slug = 'canal-vista-residencia');
insert into project_floor_plans (project_id, label, covered_area, beds, baths, balcony, price, sort_order)
select p.id, f.label, f.covered_area, f.beds, f.baths, f.balcony, f.price, f.sort_order
from projects p cross join (values
  ('1 Bed Suite', '650 sq. ft.', 1, 1, 1, 5800000::numeric, 1),
  ('2 Bed Luxury', '1,050 sq. ft.', 2, 2, 1, 9200000::numeric, 2),
  ('3 Bed Sky Penthouse', '1,450 sq. ft.', 3, 3, 2, 13000000::numeric, 3)
) as f(label, covered_area, beds, baths, balcony, price, sort_order)
where p.slug = 'canal-vista-residencia';

-- Seed Project Progress Stages
delete from project_progress_stages where project_id in (select id from projects where slug = 'canal-vista-residencia');
insert into project_progress_stages (project_id, label, pct, note, sort_order)
select p.id, s.label, s.pct, s.note, s.sort_order
from projects p cross join (values
  ('Main Boulevard & Ring Road Paving', 95, 'Near completion', 1),
  ('Underground Sewerage & Water Lines', 85, 'Final connections', 2),
  ('Sector A & B Earthwork & Demarcation', 70, 'Plots pegged', 3),
  ('Civic Center & Commercial Hub Foundation', 40, 'Raft pouring underway', 4)
) as s(label, pct, note, sort_order)
where p.slug = 'canal-vista-residencia';

-- Seed Project Payment Stages
delete from project_payment_stages where project_id in (select id from projects where slug = 'canal-vista-residencia');
insert into project_payment_stages (project_id, label, pct, count, sort_order)
select p.id, s.label, s.pct, s.count, s.sort_order
from projects p cross join (values
  ('Down Payment on Booking', 20.00::numeric, 1, 1),
  ('Quarterly Installments (3.5 Years)', 65.00::numeric, 14, 2),
  ('Final Balance on Possession', 15.00::numeric, 1, 3)
) as s(label, pct, count, sort_order)
where p.slug = 'canal-vista-residencia';

-- Seed Project Nearby Places
delete from project_nearby_places where project_id in (select id from projects where slug = 'canal-vista-residencia');
insert into project_nearby_places (project_id, place, minutes, sort_order)
select p.id, n.place, n.minutes, n.sort_order
from projects p cross join (values
  ('Faisalabad Ring Road Interchange', 4, 1),
  ('University of Agriculture Campus', 8, 2),
  ('National Hospital & Medical College', 12, 3),
  ('Susan Road Commercial Market', 15, 4),
  ('Faisalabad International Airport', 25, 5)
) as n(place, minutes, sort_order)
where p.slug = 'canal-vista-residencia';

-- ─────────────────────────────────────────────────────────────────────────
-- 3. Seed Testimonials (All 7 from public data)
-- ─────────────────────────────────────────────────────────────────────────

delete from testimonials where name in (
  'Chaudhry Nadeem', 'Dr. Farooq Sattar', 'Imran Qureshi',
  'Salman Butt', 'Majid Rasheed', 'Khadija Bibi', 'Waseem Akram'
);

insert into testimonials (name, role, photo_url, quote, sort_order) values
  ('Chaudhry Nadeem', 'Bought in Canal Road', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=300&q=80',
   'They shortlisted three houses that matched our budget and arranged all visits in one afternoon. The transfer was done smoothly in two weeks.', 1),

  ('Dr. Farooq Sattar', 'Booked a project plot', 'https://images.unsplash.com/photo-1539571696357-5a69c17a67c6?auto=format&fit=crop&w=300&q=80',
   'The payment plan was explained clearly and they checked the NOC before we paid anything. Honest advice that saved us from unapproved schemes.', 2),

  ('Imran Qureshi', 'Overseas buyer, Dubai', 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=300&q=80',
   'Bought a plot in Wapda City from Dubai. Video tours, documentation checks and title transfer were completed on time without me having to travel.', 3),

  ('Salman Butt', 'Commercial buyer', 'https://images.unsplash.com/photo-1492562080023-ab3db95bfbce?auto=format&fit=crop&w=300&q=80',
   'They found a shop on the main boulevard with a tenant already in place. Generating over 8% rental yield from month one.', 4),

  ('Majid Rasheed', 'Sold a plot', 'https://images.unsplash.com/photo-1527980965255-d3b416303d12?auto=format&fit=crop&w=300&q=80',
   'Got a fair valuation and a serious buyer within three weeks. Transparent deal with zero hidden cuts.', 5),

  ('Khadija Bibi', 'Family of five', 'https://images.unsplash.com/photo-1508214751196-bcfd4ca60f91?auto=format&fit=crop&w=300&q=80',
   'We moved into a 1 Kanal home near a school and park, exactly what we asked for. The whole family loves it.', 6),

  ('Waseem Akram', 'Rented out a house', 'https://images.unsplash.com/photo-1502685104226-ee32379fefbe?auto=format&fit=crop&w=300&q=80',
   'Tenant screening and police verification were handled without any hassle. Rent comes in on time every month.', 7);

-- ─────────────────────────────────────────────────────────────────────────
-- 4. Seed Offices (For Contact page & /api/offices)
-- ─────────────────────────────────────────────────────────────────────────

delete from offices where name in ('Head Office - Susan Road', 'Canal Road Advisory Branch', 'Lahore Regional Desk');
insert into offices (name, address, phone, hours, lat, lng) values
  ('Head Office - Susan Road', 'Office 12, Main Boulevard, Susan Road, Faisalabad', '+923000000000', 'Mon - Sat: 9:00 AM - 8:00 PM', 31.4187, 73.0791),
  ('Canal Road Advisory Branch', 'Plaza 4, Gate 1, West Canal Road, Faisalabad', '+923000000001', 'Mon - Sat: 10:00 AM - 7:00 PM', 31.4352, 73.1120),
  ('Lahore Regional Desk', 'Sector C Commercial, DHA Phase 6, Lahore', '+923000000005', 'Mon - Fri: 10:00 AM - 6:00 PM', 31.4700, 74.4300);

-- ─────────────────────────────────────────────────────────────────────────
-- 5. Seed Blog Posts & Sections
-- ─────────────────────────────────────────────────────────────────────────

insert into blog_posts (
  slug, title, category, publish_date, cover_image_url, excerpt, body,
  status, author_id, meta_title, meta_desc
) values
  ('how-to-verify-a-housing-society-faisalabad',
   'How to verify a housing society in Faisalabad before you book a plot',
   'Buying Guide', current_date - interval '3 days',
   'https://images.unsplash.com/photo-1560518883-ce09059eeffa?auto=format&fit=crop&w=1200&q=80',
   'A practical step-by-step checklist to confirm FDA approval, layout sanction, land ownership, and utility connections before investing.',
   '<p>Investing in real estate requires diligence. Before parting with your hard-earned savings, verify whether the developer owns the land registry or merely holds a memorandum of understanding.</p>',
   'Published', (select id from agents where name = 'Hamza Malik'),
   'How to Verify Housing Societies in Faisalabad (FDA Checklist)',
   'Learn how to check FDA NOC approval, master layout plans, and legal ownership before booking a plot in Faisalabad.'),

  ('canal-road-property-price-trends',
   'Canal Road property price trends: 5-year historical analysis & forecast',
   'Market Updates', current_date - interval '7 days',
   'https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?auto=format&fit=crop&w=1200&q=80',
   'Historical capital growth numbers for Canal Road residential plots, commercial corridors, and rental yields across 2021-2026.',
   '<p>Canal Road continues to demonstrate unmatched liquidity and steady annual appreciation between 14% and 18%, driven by elite educational institutions and express road connectivity.</p>',
   'Published', (select id from agents where name = 'Tariq Mehmood'),
   'Canal Road Property Price Trends & Investment Forecast',
   'Detailed 5-year historical appreciation and rental yield trends along Canal Road Faisalabad.'),

  ('overseas-pakistanis-property-buying-guide',
   'Complete property buying guide for overseas Pakistanis (UK, UAE, USA)',
   'Overseas', current_date - interval '12 days',
   'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80',
   'How to use Roshan Digital Accounts (RDA), embassy-attested Special Power of Attorney (SPA), and virtual inspections safely from abroad.',
   '<p>Thousands of overseas Pakistanis purchase property every year without traveling back home. With verified agents and RDA bank channels, remote buying is now transparent and legally secure.</p>',
   'Published', (select id from agents where name = 'Usman Farooq'),
   'Overseas Pakistanis Guide to Buying Real Estate in Pakistan',
   'Step-by-step process for purchasing verified property using Special Power of Attorney and RDA from abroad.')
on conflict (slug) do update set
  title = excluded.title, category = excluded.category, publish_date = excluded.publish_date,
  cover_image_url = excluded.cover_image_url, excerpt = excluded.excerpt, body = excluded.body,
  status = excluded.status, author_id = excluded.author_id, meta_title = excluded.meta_title,
  meta_desc = excluded.meta_desc, updated_at = now();

-- Clean and seed blog post sections
delete from blog_post_sections where post_id in (select id from blog_posts where slug = 'how-to-verify-a-housing-society-faisalabad');
insert into blog_post_sections (post_id, heading, body, sort_order)
select p.id, s.heading, s.body, s.sort_order
from blog_posts p cross join (values
  ('1. Check the Official FDA Registry', '<p>Visit the Faisalabad Development Authority (FDA) portal or visit their headquarters on Jail Road to confirm the society is on the list of approved private housing schemes.</p>', 1),
  ('2. Demand the Sanctioned Layout Plan (LOP)', '<p>An approved NOC is tied to a specific Layout Plan. Verify that the plot number you are purchasing is actually marked on the approved blueprint and not on designated public parkland or graveyard zones.</p>', 2),
  ('3. Confirm Utility Access & On-ground Work', '<p>Ensure electricity demand notices and Sui Gas connections are approved or already installed before relying on promised delivery dates.</p>', 3)
) as s(heading, body, sort_order)
where p.slug = 'how-to-verify-a-housing-society-faisalabad';

-- ─────────────────────────────────────────────────────────────────────────
-- 6. Seed Site-wide FAQs (Home, Project, Overseas)
-- ─────────────────────────────────────────────────────────────────────────

delete from faqs where entity_type in ('home', 'overseas');
insert into faqs (entity_type, entity_id, question, answer, sort_order) values
  ('home', null, 'How do I know a listing is genuine?',
   'Every listing is checked by our in-house verification team. We confirm ownership documents with the relevant authority (FDA, LDA, CDA) and physically inspect the property before it goes live.', 10),
  ('home', null, 'Do you charge buyers a commission?',
   'Commission terms depend on the property deal and are agreed in writing upfront. Standard buyer advisory is 1% on residential sales and transparently communicated before any site visit.', 20),
  ('home', null, 'Can overseas Pakistanis buy through you?',
   'Yes. Over 40% of our clients reside in the UK, UAE, USA and Saudi Arabia. We arrange live video inspections, verify land titles, and handle legal transfer via Power of Attorney.', 30),
  ('home', null, 'How long does a property transfer take in Faisalabad?',
   'For private housing societies with clear NOCs, transfers typically conclude in 7 to 14 days. Registry and mutation through the revenue department takes approximately 2 to 3 weeks.', 40),

  ('overseas', null, 'Can I buy property in Faisalabad while living abroad?',
   'Yes. We conduct HD live video tours, verify registry and NOCs, and complete title transfer through your nominated representative or an attested Power of Attorney.', 10),
  ('overseas', null, 'How do I pay from the UK, USA or UAE?',
   'Most overseas clients pay securely through their State Bank Roshan Digital Account (RDA) or via direct SWIFT transfer to the seller or developer. We never handle your purchase funds directly.', 20),
  ('overseas', null, 'Do I need to travel to Pakistan for the transfer?',
   'No. A Special Power of Attorney (SPA) attested at your nearest Pakistani Embassy or High Commission enables legal transfer without travel.', 30),
  ('overseas', null, 'Can you manage my property after purchase?',
   'Yes. Our property management service vets tenants, executes police verification, handles repairs, and deposits monthly rent into your designated account.', 40);

-- ─────────────────────────────────────────────────────────────────────────
-- 7. Seed Programmatic SEO Pages & Pins
-- ─────────────────────────────────────────────────────────────────────────

insert into seo_pages (
  slug, city, area_id, type, purpose, h1, meta_title, meta_desc, content, published
) values
  ('faisalabad/canal-road/houses-for-sale', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Canal Road'),
   'House', 'sale',
   'Houses for Sale on Canal Road Faisalabad',
   'Houses for Sale on Canal Road Faisalabad | Verified Listings',
   'Browse verified 5 Marla, 10 Marla and 1 Kanal modern designer houses for sale on Canal Road Faisalabad.',
   '<p>Canal Road offers the most sought-after houses in Faisalabad, featuring master-planned security, green belts, and proximity to the city’s top educational institutions.</p>',
   true),

  ('faisalabad/madina-town/houses-for-sale', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Madina Town'),
   'House', 'sale',
   'Houses for Sale in Madina Town Faisalabad',
   'Houses for Sale in Madina Town Faisalabad | 786 Real Estate',
   'Find 10 Marla and 1 Kanal houses for sale in Madina Town Faisalabad with verified title documents.',
   '<p>Madina Town is an established residential hub offering premier family living, central markets, and park-facing locations.</p>',
   true),

  ('faisalabad/susan-road/apartments-for-rent', 'Faisalabad',
   (select id from areas where city = 'Faisalabad' and name = 'Susan Road'),
   'Apartment', 'rent',
   'Apartments for Rent on Susan Road Faisalabad',
   'Apartments for Rent on Susan Road Faisalabad | Furnished & Luxury',
   'Explore luxury furnished 2 and 3 bed executive apartments for rent on Susan Road Faisalabad.',
   '<p>Susan Road is Faisalabad’s premier urban avenue with high rental demand and executive apartments with standby generators and 24/7 security.</p>',
   true)
on conflict (slug) do update set
  h1 = excluded.h1, meta_title = excluded.meta_title, meta_desc = excluded.meta_desc,
  content = excluded.content, published = excluded.published, updated_at = now();

-- Seed SEO Page Pins
delete from seo_page_pins where seo_page_id in (
  select id from seo_pages where slug in (
    'faisalabad/canal-road/houses-for-sale',
    'faisalabad/madina-town/houses-for-sale'
  )
);

insert into seo_page_pins (seo_page_id, property_id, sort_order)
select sp.id, p.id, 1
from seo_pages sp, properties p
where sp.slug = 'faisalabad/canal-road/houses-for-sale'
  and p.slug = '10-marla-modern-house-canal-road'
on conflict do nothing;

insert into seo_page_pins (seo_page_id, property_id, sort_order)
select sp.id, p.id, 2
from seo_pages sp, properties p
where sp.slug = 'faisalabad/canal-road/houses-for-sale'
  and p.slug = '5-marla-designer-home-canal-road'
on conflict do nothing;

insert into seo_page_pins (seo_page_id, property_id, sort_order)
select sp.id, p.id, 1
from seo_pages sp, properties p
where sp.slug = 'faisalabad/madina-town/houses-for-sale'
  and p.slug = '1-kanal-corner-house-madina-town'
on conflict do nothing;

insert into seo_page_pins (seo_page_id, property_id, sort_order)
select sp.id, p.id, 2
from seo_pages sp, properties p
where sp.slug = 'faisalabad/madina-town/houses-for-sale'
  and p.slug = '10-marla-spanish-house-madina-town'
on conflict do nothing;
