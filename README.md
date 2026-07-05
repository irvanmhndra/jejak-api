# jejak-api

Backend for **Jejak** — a self-hosted, Strava-style activity tracker (running,
cycling, walking, hiking). Go + Echo v5 + Postgres, layered architecture
(handler → service → repository), token sessions. Companion apps: `jejak-mobile`
(Flutter, GPS recording) and `jejak-web` (React dashboard).

## Run

```bash
cp .env.example .env          # adjust POSTGRES_* if needed
createdb jejak_db             # or use `make docker-up`
make run                      # migrations run automatically on boot
```

Server listens on `:8087`. Response envelope: `{ success, message?, data?, error_code? }`.

## Data model

- **activities** — one workout: type, title, source (`gps`/`manual`/`gpx`),
  `started_at`, and server-computed summary (distance, elapsed/moving seconds,
  avg/max speed, elevation gain/loss, encoded `polyline`, `point_count`).
- **activity_tracks** — the full GPS sample stream per activity (`points` JSONB),
  loaded only on the detail endpoint.
- **personal_records** — best-effort PRs per (type, record_type): longest
  distance, longest duration, most elevation, fastest average speed.
- **settings** — per-user units (`metric`/`imperial`).

Stats (distance, moving time, elevation, avg/max speed, cumulative per-point
distance, polyline) are all computed **server-side** from the uploaded GPS
points, so every client sees consistent numbers.

## Endpoints (`/api/v1`)

| Method | Path | Notes |
|--------|------|-------|
| POST | `/auth/register` `/auth/login` `/auth/refresh` | public |
| POST | `/auth/logout` · GET `/me` | auth |
| POST | `/activities` | create from GPS track `{type,title?,notes?,points:[{lat,lng,ele?,t}]}` |
| POST | `/activities/manual` | `{type,title?,started_at,elapsed_seconds,distance_meters}` |
| POST | `/activities/import?type=` | import a GPX file (raw body or multipart `file`) |
| GET | `/activities?type=&limit=&offset=` | paginated list (summary) |
| GET | `/activities/:id` | detail incl. full track `points` |
| PUT | `/activities/:id` | edit `type`/`title`/`notes` |
| DELETE | `/activities/:id` | |
| GET | `/activities/:id/gpx` | export GPX |
| GET | `/stats/summary` | totals for week/month/year/all-time, by type |
| GET | `/records` | personal records |
| GET/PUT | `/settings` | units preference |

All non-auth routes require `Authorization: Bearer <access_token>`.
