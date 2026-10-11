alter table site_settings add column hero_config jsonb not null default
'{"city":"","eyebrow":"","areaMode":"all","areaIds":[],"featuredAreaIds":[],"showAreaChips":true,"filters":{"purpose":true,"area":true,"type":true,"budget":true}}'::jsonb;

update site_settings set hero_config = hero_config || jsonb_build_object(
  'city', coalesce((select name::text from cities order by is_primary desc, name limit 1), ''),
  'featuredAreaIds', coalesce((select jsonb_agg(id order by listing_count desc, name) from (
    select a.id, a.name, coalesce(s.listing_count,0) as listing_count from areas a
    left join v_area_stats s on s.area_id=a.id
    where a.city=(select name from cities order by is_primary desc, name limit 1)
    order by listing_count desc,a.name limit 4
  ) selected), '[]'::jsonb)
);

-- Update only the original seeded copy; retain headings entered by the owner.
update site_settings set hero_title=E'Prime Locations.\nBrighter Opportunities.',
hero_sub='Explore verified houses, plots and projects in the most sought-after areas of Faisalabad.'
where hero_title='Verified houses, plots and approved projects in Faisalabad';
