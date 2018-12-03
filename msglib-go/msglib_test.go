package msglib

import (
	"bytes"
	"testing"
)

func TestMsglibIO(t *testing.T) {
	buff := &bytes.Buffer{}

	proto := NewBinaryProto()

	//
	{
		i32 := int32(224048)
		if err := proto.WriteI32(buff, i32); err != nil {
			t.Fatalf("write int32 failure: %+v", err)
		}
		t.Logf("bytes for %v = %v", i32, buff.Bytes())
		if val, err := proto.ReadI32(buff); err != nil || val != i32 {
			t.Fatalf("read int32 failure or not match: %+v", err)
		}
	}

	// float32
	{
		f32 := float32(21295.29)
		if err := proto.WriteFloat32(buff, f32); err != nil {
			t.Fatalf("write float32 failure: %+v", err)
		}
		t.Logf("bytes for %f = %v", f32, buff.Bytes())
		if val, err := proto.ReadFloat32(buff); err != nil || val != f32 {
			t.Fatalf("read float32 failure or not match: %+v", err)
		}
	}

	// float64
	{
		f64 := float64(80245295.24829)
		if err := proto.WriteFloat64(buff, f64); err != nil {
			t.Fatalf("write float64 failure: %+v", err)
		}
		t.Logf("bytes for %f = %v", f64, buff.Bytes())
		if val, err := proto.ReadFloat64(buff); err != nil || val != f64 {
			t.Fatalf("read float64 failure or not match: %+v", err)
		}
	}
}
