package scm

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"sort"
)

const (
	recordSize  = 168
	nameOffset  = 0x24
	mapSateDKey = "map-SateD"
)

// Record represents one 168-byte satellite program entry.
// We keep raw bytes to stay non-destructive for future editing.

type Record struct {
	SlotIndex int
	LCN       uint16
	Name      string
	Raw       [recordSize]byte
}

// ReadSatelliteRecords opens a Samsung .scm archive, extracts map-SateD, and parses fixed-size records.
func ReadSatelliteRecords(scmArchive string) ([]Record, error) {
	z, err := zip.OpenReader(scmArchive)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	var f *zip.File
	for _, ff := range z.File {
		if equalFold(ff.Name, mapSateDKey) || hasSuffixFold(ff.Name, "/"+mapSateDKey) || hasSuffixFold(ff.Name, mapSateDKey) {
			f = ff
			break
		}
	}
	if f == nil {
		return nil, fmt.Errorf("%s not found in %s", mapSateDKey, scmArchive)
	}

	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(buf)%recordSize != 0 {
		return nil, errors.New("unexpected map-SateD size; not multiple of 168 bytes")
	}

	n := len(buf) / recordSize
	recs := make([]Record, 0, n)
	for i := 0; i < n; i++ {
		var raw [recordSize]byte
		copy(raw[:], buf[i*recordSize:(i+1)*recordSize])
		lcn := uint16(raw[0]) | uint16(raw[1])<<8 // little-endian
		name := decodeUTF16BEZeroTerm(raw[nameOffset:])
		recs = append(recs, Record{SlotIndex: i, LCN: lcn, Name: name, Raw: raw})
	}
	return recs, nil
}

// SortTVOrder returns records with non-empty names sorted by LCN then slot index (stable).
func SortTVOrder(recs []Record) []Record {
	out := make([]Record, 0, len(recs))
	for _, r := range recs {
		if r.Name != "" {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LCN == out[j].LCN {
			return out[i].SlotIndex < out[j].SlotIndex
		}
		return out[i].LCN < out[j].LCN
	})
	return out
}
