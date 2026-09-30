-- Complete three existing demo projects without changing unrelated catalog entries.
update projects set address = area_id::text || ', ' || city::text || ', Pakistan',
  dev_about = developer || ' is the developer for this demonstration project. All prices, progress and specifications are sample data for testing.',
  dev_projects = 3, dev_years = 8, dev_families = 240,
  progress_date = '2026-09-30', price_note = 'Sample PKR prices. Confirm availability and charges with an advisor.',
  approval = 'Demo — approval verification pending', noc = 'Demo record; no official NOC supplied',
  plan_years = case when slug = 'citi-executive-enclave' then 0 else 4 end,
  down_pct = case when slug = 'citi-executive-enclave' then 100 else 20 end,
  updated_at = now()
where slug in ('canal-vista-residencia', 'grand-avenue-heights', 'citi-executive-enclave');

insert into project_units(project_id, category, name, size, price, sort_order)
select p.id, u.* from projects p cross join (values
 ('Residential Plots', '5 Marla Developed Plot', '5 Marla', 11500000::numeric, 1),
 ('Villas', '10 Marla Family Villa', '2,800 sq. ft.', 28500000::numeric, 2),
 ('Villas', '1 Kanal Executive Villa', '4,500 sq. ft.', 48000000::numeric, 3)
) u(category,name,size,price,sort_order) where p.slug='citi-executive-enclave';

insert into project_floor_plans(project_id,label,covered_area,beds,baths,balcony,price,sort_order)
select p.id, f.* from projects p cross join (values
 ('1 Bed Executive','620 sq. ft.',1,1,1,6800000::numeric,1),
 ('2 Bed Panoramic','1,120 sq. ft.',2,2,1,11500000::numeric,2),
 ('3 Bed Penthouse','1,850 sq. ft.',3,3,2,18000000::numeric,3)
) f(label,covered_area,beds,baths,balcony,price,sort_order) where p.slug='grand-avenue-heights';
insert into project_floor_plans(project_id,label,covered_area,beds,baths,balcony,price,sort_order)
select p.id, f.* from projects p cross join (values
 ('10 Marla Villa','2,800 sq. ft.',4,4,2,28500000::numeric,1),
 ('1 Kanal Villa','4,500 sq. ft.',5,6,3,48000000::numeric,2)
) f(label,covered_area,beds,baths,balcony,price,sort_order) where p.slug='citi-executive-enclave';

insert into project_payment_stages(project_id,label,pct,count,sort_order)
select p.id, s.* from projects p cross join (values
 ('Booking',20::numeric,1,1),('Quarterly installments',65::numeric,16,2),('Possession',15::numeric,1,3)
) s(label,pct,count,sort_order) where p.slug='grand-avenue-heights';
insert into project_payment_stages(project_id,label,pct,count,sort_order)
select id,'Full payment on transfer',100,1,1 from projects where slug='citi-executive-enclave';

insert into project_progress_stages(project_id,label,pct,note,sort_order)
select p.id, s.label, case when p.slug='citi-executive-enclave' then 100 else s.pct end,
 case when p.slug='citi-executive-enclave' then 'Completed (demo)' else 'Sample construction update' end,s.sort_order
from projects p cross join (values ('Infrastructure',25,1),('Structure',10,2),('Utilities',5,3)) s(label,pct,sort_order)
where p.slug in ('grand-avenue-heights','citi-executive-enclave');

insert into project_amenities(project_id,amenity)
select p.id,a.amenity from projects p cross join (values ('Gated security'),('Community park'),('Backup power'),('Visitor parking')) a(amenity)
where p.slug='citi-executive-enclave' on conflict do nothing;

insert into project_nearby_places(project_id,place,minutes,sort_order)
select p.id,n.* from projects p cross join (values ('Shopping district',10,1),('Hospital',15,2),('School',8,3)) n(place,minutes,sort_order)
where p.slug in ('grand-avenue-heights','citi-executive-enclave');

insert into faqs(entity_type,entity_id,question,answer,sort_order)
select 'project',p.id,f.question,f.answer,f.sort_order from projects p cross join (values
 ('Is this a real listing?','This is a seeded demonstration project. Verify developer documents and availability before booking.',1),
 ('How do I request a site visit?','Use Schedule a Visit. An advisor will contact you to confirm the appointment.',2),
 ('What charges are included?','Prices are sample base prices in PKR. Taxes, transfer and development charges must be confirmed separately.',3)
) f(question,answer,sort_order) where p.slug in ('canal-vista-residencia','grand-avenue-heights','citi-executive-enclave');
