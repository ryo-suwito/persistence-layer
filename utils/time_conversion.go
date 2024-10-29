package utils

import (
    "time"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// Convert time.Time to *timestamppb.Timestamp
func ToTimestamp(t time.Time) *timestamppb.Timestamp {
    if t.IsZero() {
        return nil
    }
    return timestamppb.New(t)
}

// Convert *timestamppb.Timestamp to time.Time
func ToTime(ts *timestamppb.Timestamp) time.Time {
    if ts == nil {
        return time.Time{}
    }
    return ts.AsTime()
}
