package models

type DataRecord struct {
	UserId        string  `csv:"user_id"`
	Name          string  `csv:"name"`
	FileName      string  `csv:"gpx_file"`
	Distance      float32 `csv:"distance"`
	Duration      float32 `csv:"duration"`
	Ascent        float32 `csv:"ascent"`
	Descent       float32 `csv:"descent"`
	ElevationDiff float32 `csv:"elevation_diff"`
	Trails        string  `csv:"trails"`
	RecordedAt    string  `csv:"recorded_at"`
}

type CSVFile struct {
	FileName  string
	SHA512Sum string
	Data      []*DataRecord
}
