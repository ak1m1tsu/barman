package logger

import (
	"bytes"
	"encoding/json"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// JSONFormatter outputs logrus entries as JSON with a fixed field order:
// time → level → msg → (remaining fields, sorted alphabetically).
type JSONFormatter struct {
	TimestampFormat string
}

func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	tf := f.TimestampFormat
	if tf == "" {
		tf = time.RFC3339
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	writeField := func(key string, value any, comma bool) error {
		if comma {
			buf.WriteByte(',')
		}
		k, err := json.Marshal(key)
		if err != nil {
			return err
		}
		v, err := json.Marshal(value)
		if err != nil {
			return err
		}
		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(v)
		return nil
	}

	if err := writeField("time", entry.Time.Format(tf), false); err != nil {
		return nil, err
	}
	if err := writeField("level", entry.Level.String(), true); err != nil {
		return nil, err
	}
	if err := writeField("msg", entry.Message, true); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := entry.Data[k]
		if err, ok := v.(error); ok {
			v = err.Error()
		}
		if err := writeField(k, v, true); err != nil {
			return nil, err
		}
	}

	buf.WriteString("}\n")
	return buf.Bytes(), nil
}
