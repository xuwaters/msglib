(function(G){ 'use strict';
	if (G.msglib != null) {
		return;
	}

	var msglib = {};
	G.msglib = msglib;

	// types
	msglib.t_Null = 1;
	msglib.t_Bool = 2;
	msglib.t_Byte = 3;
	msglib.t_I16 = 4;
	msglib.t_I32 = 5;
	msglib.t_I64 = 6;
	msglib.t_Float = 7;
	msglib.t_Double = 8;
	msglib.t_Binary = 9;
	msglib.t_String = 10;
	msglib.t_Struct = 11;
	msglib.t_Map = 12;
	msglib.t_List = 13;
	msglib.t_Set = 14;

	msglib.Mio = function(type) {
		var m = {
			type : type
		};
		return m;
	};

	msglib.MioField = function(id, mio) {
		return {
			id : id,
			mio : mio
		};
	};

	msglib.MioInt32 = msglib.Mio(msglib.t_I32);
	msglib.MioInt64 = msglib.Mio(msglib.t_I64);
	msglib.MioDouble = msglib.Mio(msglib.t_Double);
	msglib.MioString = msglib.Mio(msglib.t_String);

	msglib.MioList = function(elem) {
		var mio = msglib.Mio(msglib.t_List);
		mio.elem = elem;
		return mio;
	};
	msglib.MioStruct = function(fields) {
		var mio = msglib.Mio(msglib.t_Struct);
		mio.fields = fields;
		// id => field_key
		mio.fieldsIndex = {};
		for ( var k in fields) {
			var f = fields[k];
			mio.fieldsIndex[f.id] = k;
		}
		return mio;
	};
	msglib.MioMap = function(k, v) {
		var mio = msglib.Mio(msglib.t_Map);
		mio.k = k;
		mio.v = v;
		return mio;
	};

	// helpers
	var TextBuffer = function(text) {

		function parseText(text) {
			if (text == null) {
				return [];
			}
			var list = [];
			var start = 0;
			var end = 0;
			while (start < text.length) {
				var curr = [];
				for (end = start; end < text.length;) {
					var c = text.charAt(end);
					if (c == ';') {
						list.push(curr.join(""));
						++end;
						break;
					}
					//
					if (c == '\\') {
						if (end + 1 >= text.length) {
							curr.push(c);
							++end;
							break;
						}
						// next
						var n = text.charAt(end + 1);
						if (n == '\\' || n == ';') {
							++end;
							curr.push(n);
						} else {
							curr.push(c);
						}
					} else {
						curr.push(c);
					}

					// 
					++end;
					if (end >= text.length) {
						list.push(curr.join(""));
						break;
					}
				}
				//
				start = end;
			}
			return list;
		}

		function joinText(list) {
			if (list == null) {
				return "";
			}
			var sb = [];
			for (var i = 0; i < list.length; i++) {
				var val = list[i];
				if (val == null) {
					val = "";
				} else {
					val = val.replace(/\\/g, "\\\\");
					val = val.replace(/;/g, "\\;");
				}
				if (i > 0) {
					sb.push(';');
				}
				sb.push(val);
			}
			var str = sb.join("");
			return str;
		}

		// TextBuffer
		var me = this;

		//
		me.strlist = parseText(text);
		me.readIndex = 0;

		//
		me.Read = function() {
			var ret = null;
			if (me.readIndex >= 0 && me.readIndex < me.strlist.length) {
				var s = me.strlist[me.readIndex];
				if (s == null) {
					s = "";
				}
				++me.readIndex;
				ret = s;
			}
			return ret;
		};
		me.Write = function(txt) {
			if (txt == null) {
				txt = "";
			}
			me.strlist.push(txt);
		};
		me.getText = function() {
			return joinText(me.strlist);
		};
	};

	// Proto

	function MList(elemtype, count) {
		this.ElementType = elemtype;
		this.Count = count;
	}
	function MMap(keytype, valuetype, count) {
		this.KeyType = keytype;
		this.ValueType = valuetype;
		this.Count = count;
	}
	function MStruct(name) {
		this.Name = name;
	}
	function MField(name, type, id) {
		this.Name = name;
		this.Type = type;
		this.ID = id;
	}

	/**
	 * @param {TextBuffer}
	 *            buffer
	 */
	var MTextProto = function(buffer) {
		var me = this;
		me.buffer = buffer;
		//

		function readBuffer() {
			var str = me.buffer.Read();
			if (str == null) {
				throw new Error('EOF');
			}
			return str;
		}

		function writeBuffer(text) {
			me.buffer.Write(text);
		}

		me.ReadStructBegin = function() {
			var str = new MStruct("");
			return str;
		};

		me.WriteStructBegin = function(struc) {
		};

		me.ReadFieldBegin = function() {
			var field = new MField();
			var idAndType = me.ReadVarint32();
			field.Type = (idAndType & 0x0F);
			field.ID = (idAndType >>> 4);
			return field;
		};

		me.WriteFieldBegin = function(field) {
			var id = field.ID;
			var type = field.Type;
			var idAndType = ((id << 4) | (type & 0x0F));
			me.WriteVarint32(idAndType);
		};

		me.WriteFieldStop = function() {
			me.WriteFieldBegin(new MField("", msglib.t_Null, 0));
		};

		me.ReadMapBegin = function() {
			var map = new MMap();
			map.Count = me.ReadVarint32();
			var type = me.ReadByte();
			map.KeyType = (type & 0x0F);
			map.ValueType = (type >>> 4);
			return map;
		};

		me.WriteMapBegin = function(map) {
			me.WriteVarint32(map.Count);
			var val = map.ValueType;
			var key = map.KeyType;
			var type = (((val & 0x0F) << 4) | (key & 0x0F));
			me.WriteByte(type);
		};

		me.ReadListBegin = function() {
			var list = new MList();
			var countAndType = me.ReadVarint32();
			list.ElementType = (countAndType & 0x0F);
			list.Count = (countAndType >> 4);
			return list;
		};

		me.WriteListBegin = function(list) {
			var count = list.Count;
			var type = list.ElementType;
			var countAndType = ((count << 4) | (type & 0x0F));
			me.WriteVarint32(countAndType);
		};

		me.ReadSetBegin = function() {
			var list = me.ReadListBegin();
			return list;
		};

		me.WriteSetBegin = function(mlist) {
			me.WriteListBegin(mlist);
		};

		me.ReadBool = function() {
			var val = me.ReadByte();
			return val == 1;
		};

		me.WriteBool = function(b) {
			var val = (b ? 1 : 0);
			me.WriteByte(val);
		};

		me.ReadByte = function() {
			return me.ReadVarint32() & 0xFF;
		};

		me.WriteByte = function(b) {
			me.WriteVarint32(b);
		};

		me.ReadI16 = function() {
			return me.ReadVarint32();
		};

		me.WriteI16 = function(i16) {
			me.WriteVarint32(i16);
		};

		me.ReadI32 = function() {
			return me.ReadVarint32();
		};
		me.WriteI32 = function(i32) {
			me.WriteVarint32(i32);
		};

		me.ReadI64 = function() {
			return me.ReadVarint64();
		};

		me.WriteI64 = function(i64) {
			me.WriteVarint64(i64);
		};

		me.ReadFloat = function() {
			return me.ReadDouble();
		};

		me.WriteFloat = function(f) {
			me.WriteDouble(f);
		};

		me.ReadDouble = function() {
			var str = me.ReadString();
			return parseFloat(str);
		};

		me.WriteDouble = function(d) {
			me.WriteString(d.toString());
		};

		me.ReadBinary = function() {
			throw new Error("not support");
		};

		me.WriteBinary = function(binary, offset, length) {
			throw new Error("not suppport");
		};

		me.ReadString = function() {
			return readBuffer();
		};

		me.WriteString = function(str) {
			writeBuffer(str);
		};

		me.WriteVarint64 = function(n) {
			writeBuffer(n.toString());
		};

		me.WriteVarint32 = function(n) {
			writeBuffer(n.toString());
		};

		me.ReadVarint64 = function() {
			var str = readBuffer();
			return parseFloat(str);
		};

		me.ReadVarint32 = function() {
			var str = readBuffer();
			return parseInt(str);
		};
	};

	// decode / encode

	function mapsize(map) {
		var c = 0;
		for ( var k in map) {
			if (map.hasOwnProperty(k)) {
				c++;
			}
		}
		return c;
	}

	var io = {};

	io.SkipList = function(proto, mlist) {
		for (var i = 0; i < mlist.Count; i++) {
			io.Skip(proto, mlist.ElementType);
		}
	};

	io.SkipMap = function(proto, mmap) {
		for (var i = 0; i < mmap.Count; i++) {
			io.Skip(proto, mmap.KeyType);
			io.Skip(proto, mmap.ValueType);
		}
	};

	io.SkipStruct = function(proto, mstruc) {
		while (true) {
			var field = proto.ReadFieldBegin();
			if (field.Type == msglib.t_Null)
				break;
			io.Skip(proto, field.Type);
		}
	};

	io.Skip = function(proto, type) {
		switch (type) {
		case msglib.t_Binary:
			proto.ReadBinary();
			break;
		case msglib.t_Bool:
			proto.ReadBool();
			break;
		case msglib.t_Byte:
			proto.ReadByte();
			break;
		case msglib.t_Double:
			proto.ReadDouble();
			break;
		case msglib.t_Float:
			proto.ReadFloat();
			break;
		case msglib.t_I16:
			proto.ReadI16();
			break;
		case msglib.t_I32:
			proto.ReadI32();
			break;
		case msglib.t_I64:
			proto.ReadI64();
			break;
		case msglib.t_List:
			io.SkipList(proto, proto.ReadListBegin());
			break;
		case msglib.t_Map:
			io.SkipMap(proto, proto.ReadMapBegin());
			break;
		case msglib.t_Null:
			break;
		case msglib.t_Set:
			io.SkipList(proto, proto.ReadListBegin());
			break;
		case msglib.t_String:
			proto.ReadString();
			break;
		case msglib.t_Struct:
			io.SkipStruct(proto, proto.ReadStructBegin());
			break;
		}
	};

	io.WriteList = function(proto, meta, data) {
		if (data == null) {
			proto.WriteListBegin(new MList(meta.elem.type, 0));
			return;
		}
		var list = new MList(meta.elem.type, data.length);
		proto.WriteListBegin(list);
		for ( var i in data) {
			var elem = data[i];
			io.Write(proto, meta.elem, elem);
		}
	};

	io.WriteMap = function(proto, meta, data) {
		if (data == null) {
			proto.WriteMapBegin(new MMap(meta.k.type, meta.v.type, 0));
			return;
		}
		var sz = mapsize(data);
		var map = new MMap(meta.k.type, meta.v.type, sz);
		proto.WriteMapBegin(map);
		for ( var key in data) {
			var value = data[key];
			io.Write(proto, meta.Key, key);
			io.Write(proto, meta.Value, value);
		}
	};

	io.WriteStruct = function(proto, meta, data) {
		if (data == null) {
			proto.WriteFieldStop();
			return;
		}
		var struc = new MStruct();
		proto.WriteStructBegin(struc);
		for ( var key in meta.fields) {
			var field = meta.fields[key];
			var obj = data[key];
			if (obj == null)
				continue;
			var f = new MField("", field.mio.type, field.id);
			proto.WriteFieldBegin(f);
			io.Write(proto, field.mio, obj);
		}
		proto.WriteFieldStop();
	};

	io.ReadList = function(proto, meta) {
		if (meta == null)
			return null;
		var mlist = proto.ReadListBegin();
		if (mlist.ElementType != meta.elem.type) {
			io.SkipList(proto, mlist);
			return null;
		}
		var res = [];
		for (var i = 0; i < mlist.Count; i++) {
			var obj = io.Read(proto, meta.elem);
			res.push(obj);
		}
		return res;
	};

	io.ReadMap = function(proto, meta) {
		if (meta == null)
			return null;
		var map = proto.ReadMapBegin();
		if (map.KeyType != meta.k.type || map.ValueType != meta.v.type) {
			io.SkipMap(proto, map);
			return null;
		}
		var res = {};
		for (var i = 0; i < map.Count; i++) {
			var key = io.Read(proto, meta.k);
			var val = io.Read(proto, meta.v);
			res[key] = val;
		}
		return res;
	};

	io.ReadStruct = function(proto, meta) {
		if (meta == null)
			return null;
		var struc = proto.ReadStructBegin();
		var res = {};
		while (true) {
			var field = proto.ReadFieldBegin();
			if (field.Type == msglib.t_Null) {
				break;
			}
			if (!meta.fieldsIndex.hasOwnProperty(field.ID)) {
				io.Skip(proto, field.Type);
				continue;
			}
			var key = meta.fieldsIndex[field.ID];
			var miofield = meta.fields[key];
			if (miofield.mio.type != field.Type) {
				io.Skip(proto, field.Type);
				continue;
			}
			var obj = io.Read(proto, miofield.mio);
			res[key] = obj;
		}
		return res;
	};

	io.Write = function(proto, mio, data) {
		if (mio == null)
			return;
		switch (mio.type) {
		case msglib.t_Binary:
			proto.WriteBinary(data);
			break;
		case msglib.t_Bool:
			proto.WriteBool(data);
			break;
		case msglib.t_Byte:
			proto.WriteByte(data);
			break;
		case msglib.t_Double:
			proto.WriteDouble(data);
			break;
		case msglib.t_Float:
			proto.WriteFloat(data);
			break;
		case msglib.t_I16:
			proto.WriteI16(data);
			break;
		case msglib.t_I32:
			proto.WriteI32(data);
			break;
		case msglib.t_I64:
			proto.WriteI64(data);
			break;
		case msglib.t_List:
			io.WriteList(proto, mio, data);
			break;
		case msglib.t_Map:
			io.WriteMap(proto, mio, data);
			break;
		case msglib.t_Null:
			break;
		case msglib.t_Set:
			io.WriteList(proto, mio, data);
			break;
		case msglib.t_String:
			proto.WriteString(data);
			break;
		case msglib.t_Struct:
			io.WriteStruct(proto, mio, data);
			break;
		}
	};

	io.Read = function(proto, mio) {
		if (mio == null) {
			return null;
		}
		switch (mio.type) {
		case msglib.t_Binary:
			return proto.ReadBinary();
		case msglib.t_Bool:
			return proto.ReadBool();
		case msglib.t_Byte:
			return proto.ReadByte();
		case msglib.t_Double:
			return proto.ReadDouble();
		case msglib.t_Float:
			return proto.ReadFloat();
		case msglib.t_I16:
			return proto.ReadI16();
		case msglib.t_I32:
			return proto.ReadI32();
		case msglib.t_I64:
			return proto.ReadI64();
		case msglib.t_List:
			return io.ReadList(proto, mio);
		case msglib.t_Map:
			return io.ReadMap(proto, mio);
		case msglib.t_Null:
			return null;
		case msglib.t_Set:
			return io.ReadList(proto, mio);
		case msglib.t_String:
			return proto.ReadString();
		case msglib.t_Struct:
			return io.ReadStruct(proto, mio);
		}
		return null;
	};

	// exports

	msglib.decodeData = function(mio, text) {
		var buffer = new TextBuffer(text);
		var proto = new MTextProto(buffer);
		var res = io.Read(proto, mio);
		return res;
	};

	msglib.encodeData = function(mio, data) {
		var buffer = new TextBuffer();
		var proto = new MTextProto(buffer);
		io.Write(proto, mio, data);
		return buffer.getText();
	};

})(this);
