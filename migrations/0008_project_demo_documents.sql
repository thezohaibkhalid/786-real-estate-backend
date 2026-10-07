-- These frontend-served PDFs are labeled demonstration documents throughout.
update projects set brochure_url = '/documents/projects/' || slug::text || '-demo-guide.pdf',
  plan_pdf_url = '/documents/projects/' || slug::text || '-demo-guide.pdf',
  price_list_url = '/documents/projects/' || slug::text || '-demo-guide.pdf',
  master_pdf_url = '/documents/projects/' || slug::text || '-demo-guide.pdf',
  gate_downloads = false, updated_at = now()
where slug in ('canal-vista-residencia', 'grand-avenue-heights', 'citi-executive-enclave');
