-- The agency covers Faisalabad only. Other cities and their listings are removed.
-- Areas are Faisalabad colonies and societies.

delete from properties where city <> 'Faisalabad';
delete from projects where city <> 'Faisalabad';
delete from seo_pages where city <> 'Faisalabad';
delete from property_alerts where city is not null and city <> 'Faisalabad';
delete from areas where city <> 'Faisalabad';
delete from cities where name <> 'Faisalabad';

update cities set is_primary = true where name = 'Faisalabad';

insert into areas (name, city, h1, meta_title, meta_desc, description) values
  ('D Ground', 'Faisalabad', 'Property in D Ground, Faisalabad', 'D Ground Faisalabad property', 'Houses and commercial property around D Ground, Faisalabad.', '<p>D Ground is a central commercial and residential pocket beside Peoples Colony.</p>'),
  ('Civil Lines', 'Faisalabad', 'Property in Civil Lines, Faisalabad', 'Civil Lines Faisalabad property', 'Houses for sale and rent in Civil Lines, Faisalabad.', '<p>Civil Lines is an older residential area close to the city centre.</p>'),
  ('Ghulam Muhammad Abad', 'Faisalabad', 'Property in Ghulam Muhammad Abad, Faisalabad', 'Ghulam Muhammad Abad property', 'Houses and plots in Ghulam Muhammad Abad, Faisalabad.', '<p>Ghulam Muhammad Abad is a large residential colony in west Faisalabad.</p>'),
  ('Samanabad', 'Faisalabad', 'Property in Samanabad, Faisalabad', 'Samanabad Faisalabad property', 'Houses for sale and rent in Samanabad, Faisalabad.', '<p>Samanabad is a dense residential colony with local markets and schools.</p>'),
  ('Batala Colony', 'Faisalabad', 'Property in Batala Colony, Faisalabad', 'Batala Colony Faisalabad property', 'Houses in Batala Colony, Faisalabad.', '<p>Batala Colony is a central residential colony near the older city.</p>'),
  ('Nishatabad', 'Faisalabad', 'Property in Nishatabad, Faisalabad', 'Nishatabad Faisalabad property', 'Houses and commercial property in Nishatabad, Faisalabad.', '<p>Nishatabad sits on the industrial and residential edge of the city.</p>'),
  ('Millat Town', 'Faisalabad', 'Property in Millat Town, Faisalabad', 'Millat Town Faisalabad property', 'Houses and plots in Millat Town, Faisalabad.', '<p>Millat Town is a planned residential society on the Canal Road side of Faisalabad.</p>'),
  ('Amin Town', 'Faisalabad', 'Property in Amin Town, Faisalabad', 'Amin Town Faisalabad property', 'Houses for sale and rent in Amin Town, Faisalabad.', '<p>Amin Town is a residential colony with family houses and local shops.</p>'),
  ('Model Town', 'Faisalabad', 'Property in Model Town, Faisalabad', 'Model Town Faisalabad property', 'Houses in Model Town, Faisalabad.', '<p>Model Town is an established residential colony in Faisalabad.</p>'),
  ('Iqbal Town', 'Faisalabad', 'Property in Iqbal Town, Faisalabad', 'Iqbal Town Faisalabad property', 'Houses and plots in Iqbal Town, Faisalabad.', '<p>Iqbal Town is a residential colony on the eastern side of Faisalabad.</p>'),
  ('Satiana Road', 'Faisalabad', 'Property on Satiana Road, Faisalabad', 'Satiana Road Faisalabad property', 'Houses and plots on Satiana Road, Faisalabad.', '<p>Satiana Road is a main southern corridor with houses, plazas, and schemes.</p>'),
  ('Jaranwala Road', 'Faisalabad', 'Property on Jaranwala Road, Faisalabad', 'Jaranwala Road Faisalabad property', 'Houses and commercial property on Jaranwala Road, Faisalabad.', '<p>Jaranwala Road links the city centre with eastern housing and industry.</p>'),
  ('Samundri Road', 'Faisalabad', 'Property on Samundri Road, Faisalabad', 'Samundri Road Faisalabad property', 'Houses and plots on Samundri Road, Faisalabad.', '<p>Samundri Road is a southern approach road with mixed housing and commercial use.</p>')
on conflict (city, name) do nothing;
