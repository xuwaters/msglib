using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;

namespace MsgLib
{
    internal class Mio
    {
        internal static readonly Mio Null = new Mio(MType.Null);
        internal static readonly Mio Bool = new Mio(MType.Bool);
        internal static readonly Mio Byte = new Mio(MType.Byte);
        internal static readonly Mio I16 = new Mio(MType.I16);
        internal static readonly Mio I32 = new Mio(MType.I32);
        internal static readonly Mio I64 = new Mio(MType.I64);
        internal static readonly Mio Float = new Mio(MType.Float);
        internal static readonly Mio Double = new Mio(MType.Double);
        internal static readonly Mio Binary = new Mio(MType.Binary);
        internal static readonly Mio String = new Mio(MType.String);

        // 
        internal MType Type;
        internal Mio(MType type) { this.Type = type; }
    }

    internal struct MioField
    {
        internal int ID;
        internal Mio Meta;
        internal MioField(int id, Mio meta) { this.ID = id; this.Meta = meta; }
    }

    internal class MioStruct : Mio
    {
        internal SortedList<int, MioField> Fields;
        internal MioStruct(MioField[] fields) : base(MType.Struct) 
        {
            this.Fields = new SortedList<int,MioField>();
            foreach (var field in fields)
            {
                this.Fields.Add(field.ID, field);
            }
        }
    }

    internal class MioList : Mio
    {
        internal Mio Element;
        internal MioList(Mio element) : base(MType.List) { this.Element = element; }
    }

    internal class MioMap : Mio
    {
        internal Mio Key;
        internal Mio Value;
        internal MioMap(Mio key, Mio val) : base(MType.Map) { this.Key = key; this.Value = val; }
    }

    internal class MioObject
    {
        internal readonly SortedList<int, object> Fields = new SortedList<int, object>();

        internal bool HasField(int idx) { return Fields.ContainsKey(idx); }

        internal object GetObject(int idx) { object val; return Fields.TryGetValue(idx, out val) ? val : null; }
        internal void SetObject(int idx, object obj) { Fields[idx] = obj; }

        internal byte GetByte(int idx) { var obj = GetObject(idx); return obj == null ? (byte)0 : Convert.ToByte(obj); }
        internal bool GetBool(int idx) { var obj = GetObject(idx); return obj == null ? false: Convert.ToBoolean(obj); }
        internal short GetShort(int idx) { var obj = GetObject(idx); return obj == null ? (short)0 : Convert.ToInt16(obj); }
        internal int GetInt(int idx) { var obj = GetObject(idx); return obj == null ? 0 : Convert.ToInt32(obj); }
        internal long GetLong(int idx) { var obj = GetObject(idx); return obj == null ? 0L : Convert.ToInt64(obj); }
        internal float GetFloat(int idx) { var obj = GetObject(idx); return obj == null ? 0.0f : Convert.ToSingle(obj); }
        internal double GetDouble(int idx) { var obj = GetObject(idx); return obj == null ? 0.0 : Convert.ToDouble(obj); }
        internal string GetString(int idx) { var obj = GetObject(idx); return obj == null ? null : Convert.ToString(obj); }
        internal byte[] GetBytes(int idx) { var obj = GetObject(idx); return obj == null ? null : (byte[])obj; }
    }

    internal class MSimpleIO
    {
        internal static readonly MSimpleIO i = new MSimpleIO();

        //  interfaces

        internal void Serialize(MemoryStream stream, Mio meta, object data)
        {
            MemoryDataBuffer buffer = new MemoryDataBuffer(stream);
            MProto proto = new MBinaryProto(buffer);
            Write(proto, meta, data);
        }

        internal object Deserialize(byte[] bytes, Mio meta)
        {
            MemoryStream ms = new MemoryStream(bytes, false);
            return Deserialize(ms, meta);
        }

        internal object Deserialize(MemoryStream ms, Mio meta)
        {
            MemoryDataBuffer buffer = new MemoryDataBuffer(ms);
            MProto proto = new MBinaryProto(buffer);
            object res = Read(proto, meta);
            return res;
        }

        // Writes

        internal void Write(MProto proto, Mio meta, object data)
        {
            if (meta == null)
                return;
            switch (meta.Type)
            {
                case MType.Binary: proto.WriteBinary((byte[])data); break;
                case MType.Bool: proto.WriteBool(Convert.ToBoolean(data)); break;
                case MType.Byte: proto.WriteByte((sbyte)Convert.ToByte(data)); break;
                case MType.Double: proto.WriteDouble(Convert.ToDouble(data)); break;
                case MType.Float: proto.WriteFloat(Convert.ToSingle(data)); break;
                case MType.I16: proto.WriteI16(Convert.ToInt16(data)); break;
                case MType.I32: proto.WriteI32(Convert.ToInt32(data)); break;
                case MType.I64: proto.WriteI64(Convert.ToInt64(data)); break;
                case MType.List: WriteList(proto, meta as MioList, data as List<object>); break;
                case MType.Map: WriteMap(proto, meta as MioMap, data as Dictionary<object,object>); break;
                case MType.Null: break;
                case MType.Set: WriteList(proto, meta as MioList, data as List<object>); break;
                case MType.String: proto.WriteString(Convert.ToString(data)); break;
                case MType.Struct: WriteStruct(proto, meta as MioStruct, data as MioObject); break;
            }
        }

        internal void WriteList(MProto proto, MioList meta, List<object> data)
        {
            if (data == null)
            {
                proto.WriteListBegin(new MList(meta.Element.Type, 0));
                return;
            }
            var list = new MList(meta.Element.Type, data.Count);
            proto.WriteListBegin(list);
            foreach (var elem in data)
            {
                Write(proto, meta.Element, elem);
            }
        }

        internal void WriteMap(MProto proto, MioMap meta, Dictionary<object, object> data)
        {
            if (data == null)
            {
                proto.WriteMapBegin(new MMap(meta.Key.Type, meta.Value.Type, 0));
                return;
            }
            var map = new MMap(meta.Key.Type, meta.Value.Type, data.Count);
            proto.WriteMapBegin(map);
            foreach (var entry in data)
            {
                Write(proto, meta.Key, entry.Key);
                Write(proto, meta.Value, entry.Value);
            }
        }

        internal void WriteStruct(MProto proto, MioStruct meta, MioObject data)
        {
            if (data == null)
            {
                proto.WriteFieldStop();
                return;
            }
            var struc = new MStruct();
            proto.WriteStructBegin(struc);
            foreach (var field in meta.Fields.Values)
            {
                var obj = data.GetObject(field.ID);
                if (obj == null)
                    continue;
                var f = new MField(string.Empty, field.Meta.Type, field.ID);
                proto.WriteFieldBegin(f);
                Write(proto, field.Meta, obj);
            }
            proto.WriteFieldStop();
        }

        // Reads

        internal object Read(MProto proto, Mio meta)
        {
            if (meta == null)
                return null;
            switch (meta.Type)
            {
                case MType.Binary: return proto.ReadBinary();
                case MType.Bool: return proto.ReadBool();
                case MType.Byte: return proto.ReadByte();
                case MType.Double: return proto.ReadDouble();
                case MType.Float: return proto.ReadFloat();
                case MType.I16: return proto.ReadI16();
                case MType.I32: return proto.ReadI32();
                case MType.I64: return proto.ReadI64();
                case MType.List: return ReadList(proto, meta as MioList);
                case MType.Map: return ReadMap(proto, meta as MioMap);
                case MType.Null: return null;
                case MType.Set: return ReadList(proto, meta as MioList);
                case MType.String: return proto.ReadString();
                case MType.Struct: return ReadStruct(proto, meta as MioStruct);
            }
            return null;
        }

        internal List<object> ReadList(MProto proto, MioList meta)
        {
            if (meta == null)
                return null;
            var list = proto.ReadListBegin();
            if (list.ElementType != meta.Element.Type)
            {
                SkipList(proto, list);
                return null;
            }
            var res = new List<object>();
            for (int i = 0; i < list.Count; i++)
            {
                var obj = Read(proto, meta.Element);
                res.Add(obj);
            }
            return res;
        }

        internal Dictionary<object, object> ReadMap(MProto proto, MioMap meta)
        {
            if (meta == null)
                return null;
            var map = proto.ReadMapBegin();
            if (map.KeyType != meta.Key.Type || map.ValueType != meta.Value.Type)
            {
                SkipMap(proto, map);
                return null;
            }
            var res = new Dictionary<object, object>();
            for (int i = 0; i < map.Count; i++)
            {
                var key = Read(proto, meta.Key);
                var val = Read(proto, meta.Value);
                res.Add(key, val);
            }
            return res;
        }

        internal MioObject ReadStruct(MProto proto, MioStruct meta)
        {
            if (meta == null)
                return null;
            var struc = proto.ReadStructBegin();
            var res = new MioObject();
            while (true)
            {
                var field = proto.ReadFieldBegin();
                if (field.Type == MType.Null)
                    break;
                if (!meta.Fields.ContainsKey(field.ID))
                {
                    Skip(proto, field.Type);
                    continue;
                }
                var miofield = meta.Fields[field.ID];
                if (miofield.Meta.Type != field.Type)
                {
                    Skip(proto, field.Type);
                    continue;
                }
                var obj = Read(proto, miofield.Meta);
                res.SetObject(miofield.ID, obj);
            }
            return res;
        }

        // Skips

        internal void Skip(MProto proto, MType type)
        {
            switch (type)
            {
                case MType.Binary: proto.ReadBinary(); break;
                case MType.Bool: proto.ReadBool(); break;
                case MType.Byte: proto.ReadByte(); break;
                case MType.Double: proto.ReadDouble(); break;
                case MType.Float: proto.ReadFloat(); break;
                case MType.I16: proto.ReadI16(); break;
                case MType.I32: proto.ReadI32(); break;
                case MType.I64: proto.ReadI64(); break;
                case MType.List: SkipList(proto, proto.ReadListBegin()); break;
                case MType.Map: SkipMap(proto, proto.ReadMapBegin()); break;
                case MType.Null: break;
                case MType.Set: SkipList(proto, proto.ReadListBegin()); break;
                case MType.String: proto.ReadString(); break;
                case MType.Struct: SkipStruct(proto, proto.ReadStructBegin()); break;
            }
        }

        internal void SkipList(MProto proto, MList list)
        {
            for (int i = 0; i < list.Count; i++)
            {
                Skip(proto, list.ElementType);
            }
        }

        internal void SkipMap(MProto proto, MMap map)
        {
            for (int i = 0; i < map.Count; i++)
            {
                Skip(proto, map.KeyType);
                Skip(proto, map.ValueType);
            }
        }

        internal void SkipStruct(MProto proto, MStruct struc)
        {
            while (true)
            {
                var field = proto.ReadFieldBegin();
                if (field.Type == MType.Null)
                    break;
                Skip(proto, field.Type);
            }
        }
    }
}
