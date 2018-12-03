package msglib

import (
	"fmt"
	"testing"
)

type msgTest1_Tag struct {
	Hash []byte `msglib:"1"`
	Val  int32  `msglib:"2"`
}

type msgTest1 struct {
	Name    string          `msglib:"1"`
	Age     int32           `msglib:"2"`
	Tag     *msgTest1_Tag   `msglib:"3"`
	TagList []*msgTest1_Tag `msglib:"4"`
}

type msgData struct {
	Command int32  `msglib:"1"`
	Data    []byte `msglib:"2"`
}

func TestMsglibCodec2(t *testing.T) {
	obj := &msgData{}
	obj.Command = 1
	obj.Data = []byte("hello")

	bytes, err := Serialize(obj)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	var obj2 msgData
	err = Deserialize(bytes, &obj2)
	if err != nil {
		t.Fatalf("err decode = %v", err)
	}
}

func TestMsglibCodec(t *testing.T) {
	// data
	obj := new(msgTest1)
	obj.Name = "xixi"
	obj.Age = 2820
	obj.Tag = &msgTest1_Tag{Val: 242, Hash: []byte("hello")}
	obj.TagList = make([]*msgTest1_Tag, 2)
	obj.TagList[0] = &msgTest1_Tag{Val: 2824, Hash: []byte("xixi_0")}
	obj.TagList[1] = &msgTest1_Tag{Val: 2825, Hash: []byte("xixi_1")}

	// encode
	bytes, err := Serialize(obj)
	if err != nil {
		t.Fatalf("serialize object failure: %+v", err)
	}
	t.Logf("serialize bytes = %+v", bytes)
	fmt.Printf("serialize bytes = [")
	for _, k := range bytes {
		fmt.Printf(", (byte)%v", k)
	}
	fmt.Printf("]\n")

	// decode
	var obj2 msgTest1
	err = Deserialize(bytes, &obj2)
	if err != nil {
		t.Fatalf("deserialize object failure: %+v", err)
	}
	if obj2.Name != obj.Name {
		t.Fatalf("data err, expect %v, got %v", obj.Name, obj2.Name)
	}
	if obj2.Age != obj.Age {
		t.Fatalf("data err, expect %v, got %v", obj.Age, obj2.Age)
	}
	if obj2.Tag.Val != obj.Tag.Val {
		t.Fatalf("data err, expect %v, got %v", obj.Tag.Val, obj2.Tag.Val)
	}
	if len(obj2.Tag.Hash) != len(obj.Tag.Hash) {
		t.Fatalf("data err, expect %v, got %v", obj.Tag.Hash, obj2.Tag.Hash)
	}
	if len(obj2.TagList) != len(obj.TagList) {
		t.Fatalf("data err, expect %v, got %v", len(obj.TagList), len(obj2.TagList))
	}
	t.Logf("msglib decode/encode struct ok")
}
