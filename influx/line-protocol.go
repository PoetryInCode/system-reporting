package influx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
)

type LineBuilder struct {
	Measurement string
	Tags        *orderedmap.OrderedMap[string, string]
	Values      *orderedmap.OrderedMap[string, string]
	Timestamp   time.Time
}

func (lb LineBuilder) AddTag(key, value string) {
	lb.Tags.Set(key, value)
}

func (lb LineBuilder) Add(key, value string) {
	lb.Values.Set(key, value)
}

func (lb LineBuilder) Encode() string {
	sb := strings.Builder{}

	sb.WriteString(lb.Measurement)
	for key, val := range lb.Tags.AllFromFront() {
		sb.WriteString(fmt.Sprintf(",%s=%s", key, val))
	}
	sb.WriteRune(' ')

	i := 0
	total := lb.Values.Len()
	for key, val := range lb.Values.AllFromFront() {
		sb.WriteString(fmt.Sprintf("%s=%s", key, val))

		i++
		if i < total {
			sb.WriteRune(',')
		}
	}

	sb.WriteRune(' ')
	sb.WriteString(strconv.FormatInt(lb.Timestamp.UTC().UnixNano(), 10))

	return sb.String()
}

func (lb *LineBuilder) UpdateTime() {
	lb.Timestamp = time.Now()
}

func NewLineBuilder(measurement string) LineBuilder {
	return LineBuilder{
		Measurement: measurement,
		Tags:        orderedmap.NewOrderedMap[string, string](),
		Values:      orderedmap.NewOrderedMap[string, string](),
		Timestamp:   time.Now(),
	}
}
