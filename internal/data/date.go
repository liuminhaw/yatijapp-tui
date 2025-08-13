package data

import (
	"errors"
	"strconv"
	"time"
)

var ErrInvalidDateFormat = errors.New("invalid date format, expected YYYY-mm-dd")

var acceptedFormats = []string{
	"2006-01-02",
}

// type Date time.Time
type Date time.Time

// Implement a UnmarshalJSON method on the InputTime type so that it satisfies
// the json.Unmarshaler interface.
func (d *Date) UnmarshalJSON(jsonValue []byte) error {
	if string(jsonValue) == "null" {
		return nil
	}

	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidDateFormat
	}

	if unquotedJSONValue == "" {
		return nil
	}

	// var err error
	for _, format := range acceptedFormats {
		parsedTime, err := time.Parse(format, unquotedJSONValue)
		if err == nil {
			*d = Date(parsedTime)
			return nil
		}
	}

	return ErrInvalidDateFormat
}

func (d *Date) MarshalJSON() ([]byte, error) {
	// nullTime := sql.NullTime(d)
	// if !nullTime.Valid {
	// 	return []byte(""), nil
	// }
	if d == nil {
		return []byte(""), nil
	}

	// Format the time as RFC3339
	// formattedTime := time.Time(it).Format(time.RFC3339)
	// formattedTime := nullTime.Time.Format(time.RFC3339)
	// formattedTime := nullTime.Time.Format("2006-01-02")
	// return []byte(strconv.Quote(formattedTime)), nil
	formattedTime := time.Time(*d).Format("2006-01-02")
	return []byte(strconv.Quote(formattedTime)), nil
}

func (d Date) GetTime() time.Time {
	// return sql.NullTime(d).Time
	return time.Time(d)
}

func (d Date) String() string {
	// nullTime := sql.NullTime(d)
	// if !nullTime.Valid {
	// 	return ""
	// }
	//
	// // return nullTime.Time.String()
	// return nullTime.Time.Format("2006-01-02")
	return time.Time(d).Format("2006-01-02")
}
