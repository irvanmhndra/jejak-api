package service

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/irvanmhndra/jejak-api/internal/model"
)

// ---- GPX parsing ----

type gpxFile struct {
	XMLName xml.Name  `xml:"gpx"`
	Tracks  []gpxTrk  `xml:"trk"`
	Meta    gpxMeta   `xml:"metadata"`
}

type gpxMeta struct {
	Time string `xml:"time"`
}

type gpxTrk struct {
	Name string      `xml:"name"`
	Type string      `xml:"type"`
	Segs []gpxTrkSeg `xml:"trkseg"`
}

type gpxTrkSeg struct {
	Points []gpxTrkPt `xml:"trkpt"`
}

type gpxTrkPt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  *float64 `xml:"ele"`
	Time string  `xml:"time"`
}

// ParsedGPX is the result of reading a GPX file.
type ParsedGPX struct {
	Name      string
	Type      string
	StartedAt time.Time
	Points    []model.TrackPoint
}

func parseGPX(data []byte) (*ParsedGPX, error) {
	var f gpxFile
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("gpx tidak valid: %w", err)
	}
	res := &ParsedGPX{Points: []model.TrackPoint{}}
	for _, trk := range f.Tracks {
		if res.Name == "" {
			res.Name = trk.Name
		}
		if res.Type == "" {
			res.Type = trk.Type
		}
		for _, seg := range trk.Segs {
			for _, p := range seg.Points {
				pt := model.TrackPoint{Lat: p.Lat, Lng: p.Lon, Ele: p.Ele}
				if p.Time != "" {
					if t, err := time.Parse(time.RFC3339, p.Time); err == nil {
						pt.T = t
					}
				}
				res.Points = append(res.Points, pt)
			}
		}
	}
	if len(res.Points) == 0 {
		return nil, fmt.Errorf("gpx tidak memiliki titik track")
	}
	// Fall back to metadata time or first point time for the start.
	if res.Points[0].T.IsZero() {
		base := time.Now()
		if f.Meta.Time != "" {
			if t, err := time.Parse(time.RFC3339, f.Meta.Time); err == nil {
				base = t
			}
		}
		for i := range res.Points {
			res.Points[i].T = base.Add(time.Duration(i) * time.Second)
		}
	}
	res.StartedAt = res.Points[0].T
	return res, nil
}

// ---- GPX building ----

func buildGPX(a *model.Activity, points []model.TrackPoint) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<gpx version="1.1" creator="jejak" xmlns="http://www.topografix.com/GPX/1/1">` + "\n")
	b.WriteString("  <metadata><time>" + a.StartedAt.UTC().Format(time.RFC3339) + "</time></metadata>\n")
	b.WriteString("  <trk>\n")
	b.WriteString("    <name>" + xmlEscape(a.Title) + "</name>\n")
	b.WriteString("    <type>" + a.Type + "</type>\n")
	b.WriteString("    <trkseg>\n")
	for _, p := range points {
		b.WriteString(fmt.Sprintf(`      <trkpt lat="%.6f" lon="%.6f">`, p.Lat, p.Lng))
		if p.Ele != nil {
			b.WriteString(fmt.Sprintf("<ele>%.1f</ele>", *p.Ele))
		}
		if !p.T.IsZero() {
			b.WriteString("<time>" + p.T.UTC().Format(time.RFC3339) + "</time>")
		}
		b.WriteString("</trkpt>\n")
	}
	b.WriteString("    </trkseg>\n  </trk>\n</gpx>\n")
	return b.String()
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
