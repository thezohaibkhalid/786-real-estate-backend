package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func TestPublicManagedPagesAndHomepageQueries(t *testing.T) {
	url:=os.Getenv("TEST_DATABASE_URL")
	if url=="" { t.Skip("set TEST_DATABASE_URL to verify managed public pages") }
	ctx:=context.Background()
	conn,err:=pgx.Connect(ctx,url); if err!=nil { t.Fatal(err) }; defer conn.Close(ctx)
	for _,table:=range []string{"areas","properties","agents","property_media","projects","project_media","seo_pages","seo_page_pins","faqs"} {
		if _,err:=conn.Exec(ctx,`create temporary table `+table+` (like public.`+table+` including all)`);err!=nil { t.Fatal(err) }
	}
	_,err=conn.Exec(ctx,`
create temporary view v_area_stats as select a.id as area_id,count(p.id) filter(where p.status='Published') as listing_count from areas a left join properties p on p.area_id=a.id group by a.id;
create temporary view v_seo_page_status as
select s.id as seo_page_id,count(p.id) as matching_listing_count,
case when not s.published then 'Unpublished' when count(p.id)<3 then 'Hidden (< 3 listings)' else 'Indexed' end as display_status
from seo_pages s left join properties p on p.status='Published' and p.city=s.city and (s.area_id is null or p.area_id=s.area_id)
and (s.type='All' or s.type::text=p.type::text) and p.purpose=s.purpose
and (s.min_price is null or p.price>=s.min_price) and (s.max_price is null or p.price<=s.max_price)
group by s.id;
`);if err!=nil { t.Fatal(err) }
	var areaID,pageID,firstProperty int64
	if err:=conn.QueryRow(ctx,`insert into areas(name,city,description) values('Test Area','Faisalabad','Real area guide') returning id`).Scan(&areaID);err!=nil { t.Fatal(err) }
	if err:=conn.QueryRow(ctx,`insert into seo_pages(h1,slug,meta_title,city,area_id,type,purpose,published) values('Test search','faisalabad/test-area/houses-for-sale','{count} live houses','Faisalabad',$1,'House','sale',true) returning id`,areaID).Scan(&pageID);err!=nil { t.Fatal(err) }
	for _,entity:=range []struct{kind string;id int64}{{"area",areaID},{"seo_page",pageID}} {
		if _,err:=conn.Exec(ctx,`insert into faqs(entity_type,entity_id,question,answer) values($1::faq_entity_type,$2,'Can I visit?','Yes, contact our advisor.')`,entity.kind,entity.id);err!=nil { t.Fatal(err) }
	}
	app:=fiber.New();RegisterPublicRoutes(app,conn)
	get:=func(path string,status int) (map[string]json.RawMessage,map[string]json.RawMessage) {
		t.Helper();response,err:=app.Test(httptest.NewRequest("GET",path,nil),-1);if err!=nil { t.Fatal(err) };defer response.Body.Close()
		var body map[string]json.RawMessage;if err:=json.NewDecoder(response.Body).Decode(&body);err!=nil { t.Fatal(err) }
		if response.StatusCode!=status { t.Fatalf("%s status %d: %s",path,response.StatusCode,body["error"]) }
		var item,meta map[string]json.RawMessage
		_ = json.Unmarshal(body["data"],&item);_ = json.Unmarshal(body["meta"],&meta)
		return item,body
	}
	// Published content with low inventory is reachable; index status stays hidden.
	item,_:=get("/seo-pages/faisalabad%2Ftest-area%2Fhouses-for-sale",200)
	var status string;_ = json.Unmarshal(item["displayStatus"],&status)
	if status!="Hidden (< 3 listings)" { t.Fatal("low-inventory page should not be indexed") }
	area,_:=get("/areas/Faisalabad/test-area",200)
	var faqs []any;_ = json.Unmarshal(area["faqs"],&faqs);if len(faqs)!=1 { t.Fatal("area FAQs not exposed") }
	for i:=0;i<7;i++ {
		var id int64
		if err:=conn.QueryRow(ctx,`insert into properties(title,city,area_id,type,purpose,price,size,unit,status,slug,home,home_sort_order) values($1,'Faisalabad',$2,'House','sale',1000000,5,'Marla','Published',$3,$4,$5) returning id`,fmt.Sprintf("Live property %d",i),areaID,fmt.Sprintf("live-%d",i),i<6,6-i).Scan(&id);err!=nil { t.Fatal(err) }
		if i==0 { firstProperty=id }
	}
	if _,err:=conn.Exec(ctx,`insert into seo_page_pins(seo_page_id,property_id,sort_order) values($1,$2,0)`,pageID,firstProperty);err!=nil { t.Fatal(err) }
	item,_=get("/seo-pages/faisalabad/test-area/houses-for-sale?perPage=2",200)
	var properties []map[string]any;_ = json.Unmarshal(item["properties"],&properties)
	var count int;_ = json.Unmarshal(item["matchingListingCount"],&count)
	if count!=7 || len(properties)!=2 || properties[0]["id"]!=fmt.Sprint(firstProperty) { t.Fatalf("wrong count, pagination or pin order: %v",properties) }
	if properties[0]["area"]!="Test Area" || properties[0]["size"]!=float64(5) { t.Fatal("SEO property cards lack full listing data") }
	_ = json.Unmarshal(item["faqs"],&faqs);if len(faqs)!=1 { t.Fatal("SEO FAQs not exposed") }
	item,_=get("/seo-pages/faisalabad/test-area/houses-for-sale?perPage=2&page=2",200)
	_ = json.Unmarshal(item["properties"],&properties);if len(properties)!=2 || properties[0]["id"]==fmt.Sprint(firstProperty) { t.Fatal("SEO pagination repeats pinned listing") }
	_,body:=get("/properties?purpose=sale&featured=true&sort=featured&perPage=2",200)
	_ = json.Unmarshal(body["data"],&properties)
	if len(properties)!=2 || properties[0]["homeSortOrder"]!=float64(1) { t.Fatal("homepage selection ordering was not enforced by the API") }
	for i:=0;i<2;i++ { if _,err:=conn.Exec(ctx,`insert into projects(name,city,area_id,slug,status,from_price,active,home) values($1,'Faisalabad',$2,$3,'Launching',1000000,true,$4)`,fmt.Sprintf("Project %d",i),areaID,fmt.Sprintf("managed-project-%d",i),i==0);err!=nil { t.Fatal(err) } }
	_,body=get("/projects?home=true",200)
	var projects []any;_ = json.Unmarshal(body["data"],&projects);if len(projects)!=1 { t.Fatal("homepage query returned unselected projects") }
	if _,err:=conn.Exec(ctx,`update seo_pages set published=false where id=$1`,pageID);err!=nil { t.Fatal(err) }
	get("/seo-pages/faisalabad%2Ftest-area%2Fhouses-for-sale",404)
	get("/areas/Faisalabad/missing-area",404)
}
