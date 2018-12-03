package msglib

import (
	"golang.org/x/net/websocket"
)

func msglibMarshal(v interface{}) (msg []byte, payloadType byte, err error) {
	msg, err = Serialize(v)
	return msg, websocket.BinaryFrame, err
}

func msglibUnmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	return Deserialize(msg, v)
}

/*
MsgCodec is a codec to send/receive MsgLib data in a frame from a WebSocket connection.

Example:

	type T struct {
		Msg string `msglib:"1"`
		Count int `msglib:"2"`
	}

	// receive Msglib type T
	var data T
	msglib.MsgCodec.Receive(ws, &data)

	// send MsgLib type T
	msglib.MsgCodec.Send(ws, data)
*/
var MsgCodec = websocket.Codec{msglibMarshal, msglibUnmarshal}
