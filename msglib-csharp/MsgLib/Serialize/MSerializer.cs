using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Reflection;

namespace MsgLib
{
    internal class MSerializer
    {
        private static MSerializer instance = null;
        internal static MSerializer Default 
        {
            get
            {
                if (instance == null)
                {
                    instance = new MSerializer();
                }
                return instance;
            }
        }

        internal T Deserialize<T>(byte[] bytes)
        {
            MemoryStream ms = new MemoryStream(bytes, false);
            return Deserialize<T>(ms);
        }

        internal T Deserialize<T>(MemoryStream ms)
        {
            MemoryDataBuffer buffer = new MemoryDataBuffer(ms);
            MProto proto = new MBinaryProto(buffer);
            object res = Read(proto, typeof(T));
            return (T)res;
        }

        internal void Serialize(MemoryStream stream, object data)
        {
            MemoryDataBuffer buffer = new MemoryDataBuffer(stream);
            MProto proto = new MBinaryProto(buffer);
            Write(proto, data);
        }
		internal void Write(MProto target, object data)
		{
			Write(target, data, null, null);
		}
		internal void Write(MProto target, object data, Type type)
		{
			Write(target, data, type, null);
		}
        internal void Write(MProto target, object data, Type type , MMeta meta )
        {
            if (type == null)
                type = data.GetType();
            if (meta == null)
                meta = MMeta.GetMMeta(type);
            if (meta == null || MType.Null == meta.MType)
            {
                throw new MException("Unsupported Type: " + type);
            }
            else if (MType.Bool == meta.MType)
            {
                target.WriteBool((bool)data);
            }
            else if (MType.Byte == meta.MType)
            {
                target.WriteByte((sbyte)data);
            }
            else if (MType.I16 == meta.MType)
            {
                target.WriteI16((short)data);
            }
            else if (MType.I32 == meta.MType)
            {
                target.WriteI32((int)data);
            }
            else if (MType.I64 == meta.MType)
            {
                target.WriteI64((long)data);
            }
            else if (MType.Float == meta.MType)
            {
                target.WriteFloat((float)data);
            }
            else if (MType.Double == meta.MType)
            {
                target.WriteDouble((double)data);
            }
            else if (MType.Binary == meta.MType)
            {
                if (meta.MainType.IsArray)
                {
                    target.WriteBinary((byte[])data);
                }
                else
                {
                    var seg = (ArraySegment<byte>)data;
                    target.WriteBinary(seg.Array, seg.Offset, seg.Count);
                }
            }
            else if (MType.String == meta.MType)
            {
                target.WriteString((string)data);
            }
            else if (MType.Struct == meta.MType)
            {
                MStruct struc = new MStruct();
                struc.Name = type.Name;
                var fields = GetStructFields(type);
                target.WriteStructBegin(struc);
                // 
                foreach (var entry in fields)
                {
                    var proptype = entry.Value.PropertyType;
                    var fieldmeta = MMeta.GetMMeta(proptype);
                    var value = entry.Value.GetValue(data, null);
                    if (value == null)
                        continue;
                    var field = new MField(entry.Value.Name, fieldmeta.MType, entry.Key);
                    target.WriteFieldBegin(field);
                    Write(target, value, proptype, fieldmeta);
                }
                target.WriteFieldStop();
            }
            else if (MType.Map == meta.MType)
            {
                var keymeta = MMeta.GetMMeta(meta.KeyType);
                var valmeta = MMeta.GetMMeta(meta.ValueType);
                var dict = data as IDictionary;
                var map = new MMap(keymeta.MType, valmeta.MType, dict.Count);
                target.WriteMapBegin(map);
                foreach (var key in dict.Keys)
                {
                    Write(target, key, meta.KeyType, keymeta);
                    var value = dict[key];
                    Write(target, value, meta.ValueType, valmeta);
                }
            }
            else if (MType.List == meta.MType)
            {
                var valmeta = MMeta.GetMMeta(meta.ValueType);
                var arr = data as IEnumerable;
                int count = 0;
                if (meta.MainType.IsArray)
                {
                    count = (data as Array).Length;
                }
                else
                {
                    count = (data as IList).Count;
                }
                //
                var list = new MList(valmeta.MType, count);
                target.WriteListBegin(list);
                foreach (var entry in arr)
                {
                    Write(target, entry, meta.ValueType, valmeta);
                }
            }
            else if (MType.Set == meta.MType)
            {
                var valmeta = MMeta.GetMMeta(meta.ValueType);
                int count = (int)meta.ContainerType.GetProperty("Count").GetValue(data, null);
                var arr = data as IEnumerable; 
                var list = new MList(valmeta.MType, count);
                target.WriteListBegin(list);
                foreach (var entry in arr)
                {
                    Write(target, entry, meta.ValueType, valmeta);
                }
            }
        }

        private IDictionary<int, PropertyInfo> GetStructFields(Type type)
        {
            var fields = new SortedDictionary<int, PropertyInfo>();
            var allprops = type.GetTypeProperties();
            foreach (var prop in allprops)
            {
                var attrs = prop.GetCustomAttributes(typeof(MFieldAttribute), false);
                if (attrs.Length == 0)
                {
                    continue;
                }
                int id = (attrs[0] as MFieldAttribute).ID;
                fields.Add(id, prop);
            }
            return fields;
        }
		internal object Read(MProto source, Type type)
		{
			return Read(source, type, null);
		}
        internal object Read(MProto source, Type type, MMeta meta )
        {
            if (meta == null)
                meta = MMeta.GetMMeta(type);
            if (meta == null || MType.Null == meta.MType)
            {
                throw new MException("Unsupported Type: " + type);
            }
            else if (MType.Bool == meta.MType)
            {
                return source.ReadBool();
            }
            else if (MType.Byte == meta.MType)
            {
                return source.ReadByte();
            }
            else if (MType.I16 == meta.MType)
            {
                return source.ReadI16();
            }
            else if (MType.I32 == meta.MType)
            {
                return source.ReadI32();
            }
            else if (MType.I64 == meta.MType)
            {
                return source.ReadI64();
            }
            else if (MType.Float == meta.MType)
            {
                return source.ReadFloat();
            }
            else if (MType.Double == meta.MType)
            {
                return source.ReadDouble();
            }
            else if (MType.Binary == meta.MType)
            {
                byte[] bytes = source.ReadBinary();
                if (meta.MainType.IsArray)
                {
                    return bytes;
                }
                else
                {
                    return new ArraySegment<byte>(bytes);
                }
            }
            else if (MType.String == meta.MType)
            {
                return source.ReadString();
            }
            else if (MType.Struct == meta.MType)
            {
                MStruct struc = source.ReadStructBegin();
                var data = Activator.CreateInstance(type, true);
                var fields = GetStructFields(type);
                while (true)
                {
                    MField field = source.ReadFieldBegin();
                    if (field.Type == MType.Null)
                        break;
                    if (fields.ContainsKey(field.ID))
                    {
                        var prop = fields[field.ID];
                        var propmeta = MMeta.GetMMeta(prop.PropertyType);
						if (propmeta.MType != field.Type)
						{
							throw new MException("FieldErr, field = " + type.Name + ":"
                                + prop.Name + ", got type = " + field.Type);
						}
						else
						{
	                        var val = Read(source, prop.PropertyType, propmeta);
	                        prop.SetValue(data, val, null);
						}
                    }
                    else
                    {
                        Skip(source, field.Type);
                    }
                }
                return data;
            }
            else if (MType.Map == meta.MType)
            {
                var map = source.ReadMapBegin();
                object data = null;
                if (type.IsInterface)
                {
                    var instType = typeof(Dictionary<,>);
                    instType = instType.MakeGenericType(meta.KeyType, meta.ValueType);
                    data = Activator.CreateInstance(instType);
                }
                else
                {
                    data = Activator.CreateInstance(type);
                }
                var dict = data as IDictionary;
                for (int i = 0; i < map.Count; i++)
                {
                    var key = Read(source, meta.KeyType);
                    var val = Read(source, meta.ValueType);
                    dict.Add(key, val);
                }
                return data;
            }
            else if (MType.List == meta.MType)
            {
                var list = source.ReadListBegin();
                object data = null;
                if (type.IsArray)
                {
                    data = Array.CreateInstance(meta.ValueType, list.Count);
                    var arr = data as Array;
                    for (int i = 0; i < list.Count; i++)
                    {
                        var value = Read(source, meta.ValueType);
                        arr.SetValue(value, i);
                    }
                }
                else
                {
                    if (type.IsInterface)
                    {
                        var instType = typeof(List<>);
                        instType = instType.MakeGenericType(meta.ValueType);
                        data = Activator.CreateInstance(instType);
                    }
                    else
                    {
                        data = Activator.CreateInstance(type);
                    }
                    var arr = data as IList;
                    for (int i = 0; i < list.Count; i++)
                    {
                        var value = Read(source, meta.ValueType);
                        arr.Add(value);
                    }
                }
                return data;
            }
            else if (MType.Set == meta.MType)
            {
                var list = source.ReadListBegin();
                object data = null;
                var instType = type;
                if (type.IsInterface)
                {
                    instType = typeof(HashSet<>).MakeGenericType(meta.ValueType);
                }
                else if (type.IsGenericTypeDefinition)
                {
                    instType = type.MakeGenericType(meta.ValueType);
                }
                data = Activator.CreateInstance(instType);
                var addMethod = type.GetMethod("Add");
                for (int i = 0; i < list.Count; i++)
                {
                    var value = Read(source, meta.ValueType);
                    addMethod.Invoke(data, new object[] { value });
                }
                return data;
            }
            throw new MException("Should Not Reach Here");
        }

        private void Skip(MProto source, MType mType)
        {
            if (MType.Null == mType)
            {
                throw new MException("Unsupported Type: " + mType);
            }
            else if (MType.Bool == mType)
            {
                source.ReadBool();
            }
            else if (MType.Byte == mType)
            {
                source.ReadByte();
            }
            else if (MType.I16 == mType)
            {
                source.ReadI16();
            }
            else if (MType.I32 == mType)
            {
                source.ReadI32();
            }
            else if (MType.I64 == mType)
            {
                source.ReadI64();
            }
            else if (MType.Float == mType)
            {
                source.ReadFloat();
            }
            else if (MType.Double == mType)
            {
                source.ReadDouble();
            }
            else if (MType.Binary == mType)
            {
                source.ReadBinary();
            }
            else if (MType.String == mType)
            {
                source.ReadString();
            }
            else if (MType.Struct == mType)
            {
                MStruct struc = source.ReadStructBegin();
                while (true)
                {
                    MField field = source.ReadFieldBegin();
                    if (field.Type == MType.Null)
                        break;
                    Skip(source, field.Type);
                }
            }
            else if (MType.Map == mType)
            {
                var map = source.ReadMapBegin();
                for (int i = 0; i < map.Count; i++)
                {
                    Skip(source, map.KeyType);
                    Skip(source, map.ValueType);
                }
            }
            else if (MType.List == mType)
            {
                var list = source.ReadListBegin();
                for (int i = 0; i < list.Count; i++)
                {
                    Skip(source, list.ElementType);
                }
            }
            else if (MType.Set == mType)
            {
                var list = source.ReadListBegin();
                for (int i = 0; i < list.Count; i++)
                {
                    Skip(source, list.ElementType);
                }
            }
            else
            {
                throw new MException("Should Not Reach Here, type = " + mType);
            }
        }
    }

    internal class MMeta
    {
        internal MType MType { get; set; }

        internal Type MainType { get; set; }
        internal Type KeyType { get; set; }
        internal Type ValueType { get; set; }
        internal Type ContainerType { get; set; } // IDictonary or IList
		
		internal MMeta() : this (MType.Null)
		{
		}
        internal MMeta(MType type )
        {
            this.MType = type;
        }

        internal static MMeta GetMMeta(Type type)
        {
            MMeta meta = new MMeta();
            meta.MainType = type;
            if (type == typeof(String))
            {
                meta.MType = MType.String;
            }
            else if (type == typeof(Boolean))
            {
                meta.MType = MType.Bool;
            }
            else if (type == typeof(Byte) || type == typeof(SByte))
            {
                meta.MType = MType.Byte;
            }
            else if (type == typeof(Int16) || type == typeof(UInt16))
            {
                meta.MType = MType.I16;
            }
            else if (type == typeof(Int32) || type == typeof(UInt32))
            {
                meta.MType = MType.I32;
            }
            else if (type == typeof(Int64) || type == typeof(UInt64))
            {
                meta.MType = MType.I64;
            }
            else if (type == typeof(Single))
            {
                meta.MType = MType.Float;
            }
            else if (type == typeof(Double))
            {
                meta.MType = MType.Double;
            }
            else if (type == typeof(byte[]) || type == typeof(ArraySegment<byte>))
            {
                meta.MType = MType.Binary;
            }
            else if (type.IsArray)
            {
                if (type.GetArrayRank() == 1)
                {
                    meta.MType = MType.List;
                    var elemType = type.GetElementType();
                    meta.ValueType = elemType;
                }
            }
            else if (type.IsGenericType)
            {
                if (!FillGenericeType(meta, type))
                {
                    foreach (var itype in type.GetInterfaces())
                    {
                        if (FillGenericeType(meta, itype))
                            break;
                    }
                }
            }
            else
            {
                var structAttrs = type.GetCustomAttributes(typeof(MStructAttribute), false);
                if (structAttrs.Length > 0)
                {
                    meta.MType = MType.Struct;
                }
            }

            return meta;
        }

        private static bool FillGenericeType(MMeta meta, Type itype)
        {
            if (!itype.IsGenericType)
                return false;
            if (!itype.IsGenericType)
                return false;
            var gtype = itype.GetGenericTypeDefinition();
            var targs = itype.GetGenericArguments();
            if (gtype == typeof(IDictionary<,>))
            {
                gtype = gtype.MakeGenericType(targs);
                meta.MType = MType.Map;
                meta.KeyType = targs[0];
                meta.ValueType = targs[1];
                meta.ContainerType = gtype;
                return true;
            }
            else if (gtype == typeof(IList<>))
            {
                gtype = gtype.MakeGenericType(targs);
                meta.MType = MType.List;
                meta.ValueType = targs[0];
                meta.ContainerType = gtype;
                return true;
            }
            return false;
        }
    }

}
