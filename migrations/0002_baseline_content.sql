insert into cities (name, is_primary) values
  ('Faisalabad', true),
  ('Lahore', false),
  ('Islamabad', false),
  ('Rawalpindi', false),
  ('Multan', false),
  ('Sialkot', false),
  ('Gujranwala', false)
on conflict (name) do update set is_primary = excluded.is_primary;

insert into site_settings (
  id, name, phone, whatsapp, email, hero_title, hero_sub, hero_image_url,
  exit_intent_enabled, whatsapp_float_enabled
) values (
  true,
  '786 Real Estate',
  '+923000000000',
  '923000000000',
  'info@786realestate.example',
  'Verified houses, plots and approved projects in Faisalabad',
  'Trusted property advisors for buyers, renters, sellers and overseas Pakistanis.',
  '',
  true,
  true
)
on conflict (id) do update set
  name = excluded.name,
  phone = excluded.phone,
  whatsapp = excluded.whatsapp,
  email = excluded.email,
  hero_title = excluded.hero_title,
  hero_sub = excluded.hero_sub,
  hero_image_url = excluded.hero_image_url,
  exit_intent_enabled = excluded.exit_intent_enabled,
  whatsapp_float_enabled = excluded.whatsapp_float_enabled;

insert into currency_rates (code, rate_to_pkr) values
  ('GBP', 370.000000),
  ('USD', 280.000000),
  ('AED', 76.000000)
on conflict (code) do update set rate_to_pkr = excluded.rate_to_pkr;

insert into homepage_sections (key, label, sort_order, is_visible) values
  ('hero', 'Hero search', 10, true),
  ('featuredProperties', 'Featured properties', 20, true),
  ('projects', 'Projects', 30, true),
  ('cities', 'Cities', 40, true),
  ('areas', 'Popular areas', 50, true),
  ('overseas', 'Overseas buyers', 60, true),
  ('testimonials', 'Testimonials', 70, true),
  ('blog', 'Blog', 80, true),
  ('faqs', 'FAQs', 90, true),
  ('contact', 'Contact', 100, true)
on conflict (key) do update set
  label = excluded.label,
  sort_order = excluded.sort_order,
  is_visible = excluded.is_visible;

insert into legal_documents (key, title, description) values
  ('privacy', 'Privacy Policy', 'How 786 Real Estate collects and uses enquiry information.'),
  ('terms', 'Terms and Conditions', 'Website terms for buyers, sellers, renters and visitors.'),
  ('cookies', 'Cookie Policy', 'How this website may use essential and analytics cookies.'),
  ('disclaimer', 'Property Disclaimer', 'Listings, prices and availability should be confirmed with an advisor.')
on conflict (key) do update set
  title = excluded.title,
  description = excluded.description,
  last_updated_at = now();

insert into legal_document_sections (document_key, heading, body, sort_order)
select key, 'Overview', description, 10
from legal_documents d
where not exists (
  select 1 from legal_document_sections s where s.document_key = d.key
);

insert into faqs (entity_type, entity_id, question, answer, sort_order) values
  ('home', null, 'How do I contact an advisor?', 'Submit a form or call the office number shown on the website. An advisor will follow up with suitable options.', 10),
  ('home', null, 'Are listings verified?', 'Listings should be checked by the agency before publication, but final price and availability should be confirmed before any decision.', 20)
on conflict do nothing;
