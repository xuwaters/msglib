## Msglib

A simple serialization library inspired by [Apache Thrift](https://thrift.apache.org/) and [Protocol Buffers](https://developers.google.com/protocol-buffers).


A compiler is provided ```msglib-tools/msglibc```. It can compile Protocol Buffers ```.proto``` file v2 into source code of java and csharp. The serialization wire format is inspired by thrift.

There are two modes for serialization: Text Mode and Binary Mode.
Text mode is not efficient, by now, javascript only supports text mode, maybe a binary mode will be added in the future.

Although a ```msglibc``` compiler is provided for code generation, it's not neccessary.
Since data definition and serialization is seperated, it's also easy to define serializable structure in code directly.


e.g.

Javascript
```javascript
    msg.MsgPlayer = m.MioStruct({
        playerid : m.MioField(1, m.MioInt32),
        name : m.MioField(2, m.MioString),
        uid : m.MioField(3, m.MioInt64),
        skill : m.MioField(4, m.MioString)
    });

```

Go
```go

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

```
