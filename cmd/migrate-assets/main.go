package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"example.com/786-real-estate/backend/internal/config"
	"example.com/786-real-estate/backend/internal/database"
	"example.com/786-real-estate/backend/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type asset struct {
	table, column, id, slug, source string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	db, err := database.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	r2, err := storage.NewR2(cfg)
	if err != nil || r2 == nil {
		return fmt.Errorf("R2 is not fully configured")
	}

	assets, err := databaseAssets(ctx, db)
	if err != nil {
		return err
	}
	var failures []string
	for _, item := range assets {
		if err := migrateRemoteAsset(ctx, db, r2, item); err != nil {
			log.Printf("skipping %s/%s: %v", item.table, item.id, err)
			failures = append(failures, item.table+"/"+item.id)
		}
	}
	if err := migrateProjectPDFs(ctx, db, r2); err != nil {
		return err
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d seeded remote assets could not be migrated: %s", len(failures), strings.Join(failures, ", "))
	}
	return nil
}

func databaseAssets(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}) ([]asset, error) {
	rows, err := db.Query(ctx, `
		select 'areas', 'cover_image_url', id::text, lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g')), cover_image_url
		from areas where cover_image_url <> ''
		union all
		select 'property_media', 'url', pm.id::text, p.slug::text, pm.url
		from property_media pm join properties p on p.id = pm.property_id
		where pm.url <> ''
		union all
		select 'project_media', 'url', pm.id::text, p.slug::text, pm.url
		from project_media pm join projects p on p.id = pm.project_id
		where pm.url <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assets []asset
	for rows.Next() {
		var item asset
		if err := rows.Scan(&item.table, &item.column, &item.id, &item.slug, &item.source); err != nil {
			return nil, err
		}
		assets = append(assets, item)
	}
	return assets, rows.Err()
}

func migrateRemoteAsset(ctx context.Context, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, r2 *storage.R2, item asset) error {
	if strings.Contains(item.source, ".r2.dev/") || strings.Contains(item.source, "cdn.786realestates.com/") {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.source, nil)
	if err != nil {
		return fmt.Errorf("download %s: %w", item.source, err)
	}
	res, err := (&http.Client{Timeout: 45 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", item.source, err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if res.StatusCode == http.StatusNotFound {
			res.Body.Close()
			fallback := "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80"
			log.Printf("using fallback demo image for unavailable source %s", item.source)
			return migrateRemoteAsset(ctx, db, r2, asset{table: item.table, column: item.column, id: item.id, slug: item.slug, source: fallback})
		}
		return fmt.Errorf("download %s: HTTP %s", item.source, res.Status)
	}
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		contentType = "image/jpeg"
	}
	name := filepath.Base(parsedPath(item.source))
	if name == "." || name == "/" || name == "" {
		name = "image.jpg"
	}
	key := "properties/" + item.slug + "/images/" + safeName(name)
	if item.table == "project_media" {
		key = "projects/" + item.slug + "/images/" + safeName(name)
	} else if item.table == "areas" {
		key = "areas/" + item.slug + "/images/" + safeName(name)
	}
	publicURL, err := r2.Put(ctx, key, res.Body, res.ContentLength, contentType)
	if err != nil {
		return err
	}
	var query string
	switch item.table {
	case "areas":
		query = "update areas set cover_image_url=$1 where id=$2"
	case "property_media":
		query = "update property_media set url=$1 where id=$2"
	case "project_media":
		query = "update project_media set url=$1 where id=$2"
	default:
		return fmt.Errorf("unsupported asset table %s", item.table)
	}
	_, err = db.Exec(ctx, query, publicURL, item.id)
	return err
}

func migrateProjectPDFs(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, r2 *storage.R2) error {
	root := filepath.Join("..", "786-real-estate-frontend", "public", "documents", "projects")
	rows, err := db.Query(ctx, "select id::text, slug::text from projects where slug in ('canal-vista-residencia','grand-avenue-heights','citi-executive-enclave')")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return err
		}
		source := filepath.Join(root, slug+"-demo-guide.pdf")
		file, err := os.Open(source)
		key := "projects/" + slug + "/pdf/demo-guide.pdf"
		var publicURL string
		if err == nil {
			stat, statErr := file.Stat()
			if statErr != nil {
				file.Close()
				return statErr
			}
			publicURL, err = r2.Put(ctx, key, file, stat.Size(), "application/pdf")
			file.Close()
			if err != nil {
				return err
			}
		} else if os.IsNotExist(err) {
			publicURL = r2.URL(key)
		} else {
			return fmt.Errorf("open demo PDF %s: %w", source, err)
		}
		if _, err := db.Exec(ctx, `update projects set brochure_url=$1, plan_pdf_url=$1, price_list_url=$1, master_pdf_url=$1, gate_downloads=false where id=$2`, publicURL, id); err != nil {
			return err
		}
	}
	return rows.Err()
}

func parsedPath(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "image.jpg"
	}
	return path.Base(parsed.Path)
}

func safeName(raw string) string {
	raw = strings.ToLower(raw)
	raw = strings.NewReplacer(" ", "-", "_", "-", "?", "-", "&", "-").Replace(raw)
	return raw
}
