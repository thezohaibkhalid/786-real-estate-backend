update projects p set address = a.name || ', ' || p.city::text || ', Pakistan',
 description = p.description || ' This is a demonstration listing with sample specifications and prices.', updated_at=now()
from areas a where a.id=p.area_id and p.slug in ('canal-vista-residencia','grand-avenue-heights','citi-executive-enclave');
update project_payment_stages set label='Quarterly installments (4 years)', count=16
where project_id=(select id from projects where slug='canal-vista-residencia') and count=14;
