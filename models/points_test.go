package models_test

import (
	"reflect"
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/testutils"

	"google.golang.org/protobuf/proto"
)

func TestPointToProto(t *testing.T) {
	point := testutils.Point(testutils.ParseUUID("7f753007-0eda-4aec-8d25-de6ac96220fc"), testutils.ParseTime("2024-09-08T17:39:10Z"), 0)

	startProtoPoint := point.ToProto()
	data, err := proto.Marshal(startProtoPoint)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expected := []byte{
		13, 205, 204, 140, 63, 21, 205, 204, 140, 63, 24, 2, 37, 205, 204, 140, 63, 45, 205, 204, 140, 63, 48, 2, 56, 2, 64, 2, 72, 2, 85, 205, 204, 140, 63,
		93, 205, 204, 140, 63, 101, 205, 204, 140, 63, 109, 205, 204, 140, 63, 117, 205, 204, 140, 63, 125, 205, 204, 140, 63, 133, 1, 205, 204, 140, 63, 141,
		1, 205, 204, 140, 63, 149, 1, 205, 204, 140, 63, 157, 1, 205, 204, 140, 63, 165, 1, 205, 204, 140, 63, 173, 1, 0, 64, 156, 69, 176, 1, 4,
	}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected %v got %v", expected, data)
	}

	var protoPoint models.ApiPoint
	err = proto.Unmarshal(data, &protoPoint)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	if !proto.Equal(startProtoPoint, &protoPoint) {
		t.Errorf("expected %v got %v", startProtoPoint, &protoPoint)
	}
}
