package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestScanTimelineMediaAcceptsNullableRenditionMetadata(t *testing.T) {
	scanner := timelineScannerWithNullRenditionMetadata{}

	item, err := scanTimelineMedia(scanner)
	if err != nil {
		t.Fatalf("scan timeline media: %v", err)
	}
	if item.Display.Width != 0 || item.Display.Height != 0 || item.Display.DurationMillis != 0 || item.Display.ErrorMessage != "" {
		t.Fatalf("expected nullable display metadata to become zero values, got %#v", item.Display)
	}
	if item.Thumbnail.Width != 0 || item.Thumbnail.Height != 0 || item.Thumbnail.DurationMillis != 0 || item.Thumbnail.ErrorMessage != "" {
		t.Fatalf("expected nullable thumbnail metadata to become zero values, got %#v", item.Thumbnail)
	}
}

type timelineScannerWithNullRenditionMetadata struct{}

func (timelineScannerWithNullRenditionMetadata) Scan(dest ...any) error {
	if _, ok := dest[16].(*sql.NullInt64); !ok {
		return fmt.Errorf("display width must scan through sql.NullInt64")
	}
	if _, ok := dest[17].(*sql.NullInt64); !ok {
		return fmt.Errorf("display height must scan through sql.NullInt64")
	}
	if _, ok := dest[18].(*sql.NullInt64); !ok {
		return fmt.Errorf("display duration must scan through sql.NullInt64")
	}
	if _, ok := dest[20].(*sql.NullString); !ok {
		return fmt.Errorf("display error must scan through sql.NullString")
	}
	if _, ok := dest[27].(*sql.NullInt64); !ok {
		return fmt.Errorf("thumbnail width must scan through sql.NullInt64")
	}
	if _, ok := dest[28].(*sql.NullInt64); !ok {
		return fmt.Errorf("thumbnail height must scan through sql.NullInt64")
	}
	if _, ok := dest[29].(*sql.NullInt64); !ok {
		return fmt.Errorf("thumbnail duration must scan through sql.NullInt64")
	}
	if _, ok := dest[31].(*sql.NullString); !ok {
		return fmt.Errorf("thumbnail error must scan through sql.NullString")
	}

	assignInt64(dest[0], 1)
	assignInt64(dest[1], 1)
	assignInt64(dest[2], 1)
	assignString(dest[3], MediaTypePhoto)
	assignString(dest[4], MediaStatusActive)
	assignString(dest[5], RenditionStatusReady)
	assignTime(dest[7], time.Date(2026, 6, 13, 15, 0, 0, 0, time.UTC))
	assignString(dest[9], "妈妈")
	assignInt64(dest[10], 2)
	assignInt64(dest[11], 1)
	assignString(dest[12], RenditionTypeDisplayImage)
	assignString(dest[13], "previews/media/1/display.jpg")
	assignString(dest[14], "image/jpeg")
	assignInt64(dest[15], 12345)
	assignString(dest[19], RenditionStatusReady)
	assignNullableInt64(dest[21], 3)
	assignNullableInt64(dest[22], 1)
	assignNullableString(dest[23], RenditionTypeThumbnail)
	assignNullableString(dest[24], "previews/media/1/thumbnail.jpg")
	assignNullableString(dest[25], "image/jpeg")
	assignNullableInt64(dest[26], 6789)
	assignNullableString(dest[30], RenditionStatusReady)
	return nil
}

func assignInt64(dest any, value int64) {
	*(dest.(*int64)) = value
}

func assignString(dest any, value string) {
	*(dest.(*string)) = value
}

func assignTime(dest any, value time.Time) {
	*(dest.(*time.Time)) = value
}

func assignNullableInt64(dest any, value int64) {
	typed := dest.(*sql.NullInt64)
	typed.Int64 = value
	typed.Valid = true
}

func assignNullableString(dest any, value string) {
	typed := dest.(*sql.NullString)
	typed.String = value
	typed.Valid = true
}
