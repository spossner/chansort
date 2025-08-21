package scm

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const (
	recordSize  = 168
	nameOffset  = 0x24
	mapSateDKey = "map-SateD"
)

// Record represents one 168-byte satellite program entry.
// We keep raw bytes to stay non-destructive for future editing.
type Channel struct {
	ID      int
	OrderId uint16
	Name    string
	Raw     []byte
}

// ReadSatelliteRecords opens a Samsung .scm archive, extracts map-SateD, and parses fixed-size records.
func ReadSatelliteRecords(scmArchive string) ([]Channel, error) {
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
	recs := make([]Channel, 0, n)
	for i := 0; i < n; i++ {
		raw := make([]byte, recordSize)
		copy(raw[:], buf[i*recordSize:(i+1)*recordSize])
		lcn := uint16(raw[0]) | uint16(raw[1])<<8 // little-endian
		name := decodeUTF16BEZeroTerm(raw[nameOffset:])
		recs = append(recs, Channel{ID: i, OrderId: lcn, Name: name, Raw: raw})
	}
	return recs, nil
}

// WriteSatelliteRecords copies the content of a given scm archive with a new sattelite records list into a new output archive. The function clones the original archive and only replaces the eventually modified list of sattelite channels in a probably new order
func WriteSatelliteRecords(scmArchive string, recs []Channel, outputArchive string) error {
	// Open the original zip archive
	zipReader, err := zip.OpenReader(scmArchive)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Create a new zip archive for the output
	outFile, err := os.Create(outputArchive)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// Copy all files from the original archive to the new archive
	for _, f := range zipReader.File {
		if f.Name == mapSateDKey {
			// Write the modified satellite records to the new archive
			f, err := zipWriter.Create(mapSateDKey)
			if err != nil {
				return err
			}
			for _, rec := range recs {
				if _, err := f.Write(rec.Raw[:]); err != nil {
					return err
				}
			}
		} else {
			// Copy the original file
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			fw, err := zipWriter.Create(f.Name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(fw, rc); err != nil {
				return err
			}
		}
	}
	return nil
}

func WriteOrderIdsBackIntoChannelData(recs []Channel) {
	for _, rec := range recs {
		rec.Raw[0] = byte(rec.OrderId & 0xFF)        // LSB
		rec.Raw[1] = byte((rec.OrderId >> 8) & 0xFF) // MSB

		// Recalculate and update checksum (sum of first 167 bytes mod 256)
		var checksum byte
		for i := range recordSize - 1 {
			checksum += rec.Raw[i]
		}
		rec.Raw[recordSize-1] = checksum
	}
}

// SortTVOrder returns records with non-empty names sorted by LCN then slot index (stable).
func SortTVOrder(recs []Channel) []Channel {
	out := make([]Channel, 0, len(recs))
	for _, r := range recs {
		if r.Name != "" {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OrderId == out[j].OrderId {
			return out[i].ID < out[j].ID
		}
		return out[i].OrderId < out[j].OrderId
	})
	return out
}
