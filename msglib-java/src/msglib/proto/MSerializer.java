package msglib.proto;

import java.lang.reflect.Array;
import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.TreeMap;

import msglib.annotation.MsgField;
import msglib.proto.io.IDataBuffer;
import msglib.proto.io.MBinaryProto;
import msglib.proto.io.MException;
import msglib.proto.io.MProto;
import msglib.proto.io.MTextProto;
import msglib.proto.io.MemoryDataBuffer;
import msglib.proto.io.MemoryTextBuffer;
import msglib.proto.types.MField;
import msglib.proto.types.MList;
import msglib.proto.types.MMap;
import msglib.proto.types.MStruct;
import msglib.proto.types.MType;

public class MSerializer {

    private static MSerializer instance = new MSerializer();

    public static MSerializer inst() {
        return instance;
    }

    public static String SerializeAsText(Object data) throws MException {
        MemoryTextBuffer buffer = new MemoryTextBuffer();
        MProto target = new MTextProto(buffer);
        instance.Write(target, data);
        return buffer.getText();
    }

    public static <T> T Deserialize(String text, Class<T> clazz) throws MException {
        MemoryTextBuffer buffer = new MemoryTextBuffer(text);
        MProto source = new MTextProto(buffer);
        return (T) instance.Read(source, clazz);
    }

    public static byte[] Serialize(Object data) throws MException {
        MemoryDataBuffer buffer = new MemoryDataBuffer();
        MProto target = new MBinaryProto(buffer);
        instance.Write(target, data);
        return buffer.toArray();
    }

    public static void Serialize(Object msg, IDataBuffer buffer) throws MException {
        MProto target = new MBinaryProto(buffer);
        instance.Write(target, msg);
    }

    public static <T> T Deserialize(IDataBuffer buffer, Class<T> clazz) throws MException {
        MProto source = new MBinaryProto(buffer);
        return (T) instance.Read(source, clazz);
    }

    public static <T> T Deserialize(byte[] data, Class<T> clazz) throws MException {
        MemoryDataBuffer buffer = new MemoryDataBuffer(data, false);
        MProto source = new MBinaryProto(buffer);
        return (T) instance.Read(source, clazz);
    }

    public static <T> T Deserialize(byte[] data, int offset, int length, Class<T> clazz) throws MException {
        MemoryDataBuffer buffer = new MemoryDataBuffer(data, offset, length, false);
        MProto source = new MBinaryProto(buffer);
        return (T) instance.Read(source, clazz);
    }

    public static <T> T DecodeNoThrow(IDataBuffer buffer, Class<T> clazz) {
        try {
            return Deserialize(buffer, clazz);
        }
        catch (MException e) {
            e.printStackTrace();
            return null;
        }
    }

    public static <T> T DecodeNoThrow(byte[] data, Class<T> clazz) {
        try {
            return Deserialize(data, clazz);
        }
        catch (MException e) {
            e.printStackTrace();
            return null;
        }
    }

    public void Write(MProto target, Object data) throws MException {
        Write(target, data, null);
    }

    public void Write(MProto target, Object data, Class<?> type) throws MException {
        Write(target, data, type, null);
    }

    public void Write(MProto target, Object data, Class<?> type, MMeta meta) throws MException {
        if (type == null)
            type = data.getClass();
        if (meta == null)
            meta = MMeta.GetClassMeta(type);

        if (meta == null || MType.Null == meta.Type) {
            throw new MException("Unsupported Type: " + type);
        }
        if (MType.Bool == meta.Type) {
            target.WriteBool((Boolean) data);
            return;
        }
        if (MType.Byte == meta.Type) {
            target.WriteByte((Byte) data);
            return;
        }
        if (MType.I16 == meta.Type) {
            target.WriteI16((Short) data);
            return;
        }
        if (MType.I32 == meta.Type) {
            target.WriteI32((Integer) data);
            return;
        }
        if (MType.I64 == meta.Type) {
            target.WriteI64((Long) data);
            return;
        }
        if (MType.Float == meta.Type) {
            target.WriteFloat((Float) data);
            return;
        }
        if (MType.Double == meta.Type) {
            target.WriteDouble((Double) data);
            return;
        }
        if (MType.Binary == meta.Type) {
            if (meta.MainType.isArray()) {
                target.WriteBinary((byte[]) data);
                return;
            }
            throw new MException("Binary should be byte[]");
        }
        if (MType.String == meta.Type) {
            target.WriteString((String) data);
            return;
        }
        if (MType.Struct == meta.Type) {
            MStruct struc = new MStruct();
            struc.Name = type.getName();
            Map<Integer, Field> fields = GetStructFields(type);
            target.WriteStructBegin(struc);
            //
            for (Entry<Integer, Field> entry : fields.entrySet()) {
                Field fieldinfo = entry.getValue();
                if (null != fieldinfo.getAnnotation(Deprecated.class))
                    continue;
                MMeta fieldmeta = MMeta.GetFieldMeta(fieldinfo);
                MField field = new MField(fieldinfo.getName(), fieldmeta.Type, entry.getKey());
                //
                Object value = null;
                try {
                    value = fieldinfo.get(data);
                }
                catch (Exception e) {
                    throw new MException("Field Access Error: " + type.getName() + "." + fieldinfo.getName(), e);
                }
                if (value == null)
                    continue;
                target.WriteFieldBegin(field);
                Write(target, value, fieldinfo.getType(), fieldmeta);
            }
            target.WriteFieldStop();

            return;
        }

        if (MType.Map == meta.Type) {
            MMeta keymeta = MMeta.GetClassMeta(meta.KeyType);
            MMeta valmeta = MMeta.GetClassMeta(meta.ValueType);
            Map dict = (Map) data;
            MMap map = new MMap(keymeta.Type, valmeta.Type, dict.size());
            target.WriteMapBegin(map);
            for (Object obj : dict.entrySet()) {
                Entry entry = (Entry) obj;
                Object key = entry.getKey();
                Object value = entry.getValue();
                Write(target, key, meta.KeyType, keymeta);
                Write(target, value, meta.ValueType, valmeta);
            }
            return;
        }

        if (MType.List == meta.Type) {
            MMeta valmeta = MMeta.GetClassMeta(meta.ValueType);
            if (meta.MainType.isArray()) {
                int count = Array.getLength(data);
                //
                MList list = new MList(valmeta.Type, count);
                target.WriteListBegin(list);
                for (int i = 0; i < count; i++) {
                    Object entry = Array.get(data, i);
                    Write(target, entry, meta.ValueType, valmeta);
                }
                return;
            }

            if (List.class.isAssignableFrom(meta.MainType)) {
                List arr = (List) data;
                MList list = new MList(valmeta.Type, arr.size());
                target.WriteListBegin(list);
                for (Object entry : arr) {
                    Write(target, entry, meta.ValueType, valmeta);
                }
                return;
            }

            throw new MException("unsupported list type: " + meta.MainType.getName());
        }

        throw new MException("Unsupported Type: " + meta.Type);
    }

    /**
     * TODO: [optimization] cache field info of given type
     * @param type
     * @return
     * @throws MException
     */
    private Map<Integer, Field> GetStructFields(Class<?> type) throws MException {
        Map<Integer, Field> fields = new TreeMap<Integer, Field>();
        for (Field field : type.getFields()) {
            MsgField msgfield = field.getAnnotation(MsgField.class);
            if (msgfield == null) {
                continue;
            }
            if (fields.containsKey(msgfield.Id())) {
                throw new MException("Duplicate FieldId, struct = " + type.getCanonicalName() + ", field = "
                        + msgfield.Id() + ", name = " + field.getName());
            }
            fields.put(msgfield.Id(), field);
        }
        return fields;
    }

    public Object Read(MProto source, Class<?> type) throws MException {
        return Read(source, type, null);
    }

    public Object Read(MProto source, Class<?> type, MMeta meta) throws MException {
        if (meta == null)
            meta = MMeta.GetClassMeta(type);
        if (meta == null || MType.Null == meta.Type) {
            throw new MException("Unsupported Type: " + type);
        }
        else if (MType.Bool == meta.Type) {
            return source.ReadBool();
        }
        else if (MType.Byte == meta.Type) {
            return source.ReadByte();
        }
        else if (MType.I16 == meta.Type) {
            return source.ReadI16();
        }
        else if (MType.I32 == meta.Type) {
            return source.ReadI32();
        }
        else if (MType.I64 == meta.Type) {
            return source.ReadI64();
        }
        else if (MType.Float == meta.Type) {
            return source.ReadFloat();
        }
        else if (MType.Double == meta.Type) {
            return source.ReadDouble();
        }
        else if (MType.Binary == meta.Type) {
            byte[] bytes = source.ReadBinary();
            if (meta.MainType.isArray()) {
                return bytes;
            }
            else {
                throw new MException("Binary shoud be byte[]");
            }
        }
        else if (MType.String == meta.Type) {
            return source.ReadString();
        }
        else if (MType.Struct == meta.Type) {
            MStruct struc = source.ReadStructBegin();
            Object data = null;
            try {
                data = type.getConstructor().newInstance();
            }
            catch (Exception e) {
                throw new MException("Error to Create Struct " + type.getName(), e);
            }

            Map<Integer, Field> fields = GetStructFields(type);
            while (true) {
                MField field = source.ReadFieldBegin();
                if (field.Type == MType.Null)
                    break;
                if (fields.containsKey(field.ID)) {
                    Field prop = fields.get(field.ID);
                    MMeta propmeta = MMeta.GetFieldMeta(prop);
                    Object value = Read(source, prop.getType(), propmeta);
                    try {
                        prop.set(data, value);
                    }
                    catch (Exception e) {
                        throw new MException("Error Writing Field " + type.getName() + "." + prop.getName(), e);
                    }
                }
                else {
                    Skip(source, field.Type);
                }
            }
            return data;
        }
        else if (MType.Map == meta.Type) {
            MMap map = source.ReadMapBegin();

            if (!Map.class.isAssignableFrom(type))
                throw new MException("Map type should be assignable to java.util.Map");

            Object data = null;
            if (type.isInterface()) {
                data = new HashMap();
            }
            else {
                try {
                    data = type.getConstructor().newInstance();
                }
                catch (Exception e) {
                    throw new MException("Failed to create Map " + type.getName(), e);
                }
            }

            Map dict = (Map) data;
            for (int i = 0; i < map.Count; i++) {
                Object key = Read(source, meta.KeyType);
                Object value = Read(source, meta.ValueType);
                dict.put(key, value);
            }

            return data;
        }
        else if (MType.List == meta.Type) {
            MList list = source.ReadListBegin();
            Object data = null;
            if (type.isArray()) {
                data = Array.newInstance(meta.ValueType, list.Count);
                for (int i = 0; i < list.Count; i++) {
                    Object value = Read(source, meta.ValueType);
                    Array.set(data, i, value);
                }
            }
            else {
                if (!List.class.isAssignableFrom(type))
                    throw new MException("List type should be assignable to java.util.List");

                if (type.isInterface()) {
                    data = new ArrayList();
                }
                else {
                    try {
                        data = type.getConstructor().newInstance();
                    }
                    catch (Exception e) {
                        throw new MException("Error Creating List " + type.getName(), e);
                    }
                }

                List arr = (List) data;
                for (int i = 0; i < list.Count; i++) {
                    Object value = Read(source, meta.ValueType);
                    arr.add(value);
                }
            }
            return data;
        }
        else {
            throw new MException("Unsupported MsgType " + type);
        }
    }

    private void Skip(MProto source, int type) throws MException {
        if (MType.Null == type) {
            throw new MException("Unsupported Type: " + type);
        }
        else if (MType.Bool == type) {
            source.ReadBool();
        }
        else if (MType.Byte == type) {
            source.ReadByte();
        }
        else if (MType.I16 == type) {
            source.ReadI16();
        }
        else if (MType.I32 == type) {
            source.ReadI32();
        }
        else if (MType.I64 == type) {
            source.ReadI64();
        }
        else if (MType.Float == type) {
            source.ReadFloat();
        }
        else if (MType.Double == type) {
            source.ReadDouble();
        }
        else if (MType.Binary == type) {
            source.ReadBinary();
        }
        else if (MType.String == type) {
            source.ReadString();
        }
        else if (MType.Struct == type) {
            MStruct struc = source.ReadStructBegin();
            while (true) {
                MField field = source.ReadFieldBegin();
                if (field.Type == MType.Null)
                    break;
                Skip(source, field.Type);
            }
        }
        else if (MType.Map == type) {
            MMap map = source.ReadMapBegin();
            for (int i = 0; i < map.Count; i++) {
                Skip(source, map.KeyType);
                Skip(source, map.ValueType);
            }
        }
        else if (MType.List == type) {
            MList list = source.ReadListBegin();
            for (int i = 0; i < list.Count; i++) {
                Skip(source, list.ElementType);
            }
        }
        else {
            throw new MException("Unsupported MsgType " + type);
        }
    }

}
