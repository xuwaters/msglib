package msglib.proto;

import java.lang.reflect.Field;
import java.util.List;
import java.util.Map;

import msglib.annotation.MsgList;
import msglib.annotation.MsgMap;
import msglib.annotation.MsgStruct;
import msglib.proto.types.MType;

public class MMeta {
    public int Type;

    public Class<?> MainType;
    public Class<?> KeyType;
    public Class<?> ValueType;
    public Class<?> ContainerType; // List or Map

    public MMeta() {
        this.Type = MType.Null;
    }

    public MMeta(int type) {
        this.Type = type;
    }

    public static MMeta GetFieldMeta(Field field) {
        Class<?> fieldType = field.getType();
        MMeta meta = GetClassMeta(fieldType);

        if (meta.Type != MType.Null)
            return meta;

        MsgList list = field.getAnnotation(MsgList.class);
        if (list != null) {
            if (list.ElemClass() != null && List.class.isAssignableFrom(fieldType)) {
                meta.Type = MType.List;
                meta.ValueType = list.ElemClass();
            }
            return meta;
        }

        MsgMap map = field.getAnnotation(MsgMap.class);
        if (map != null) {
            if (map.KeyClass() != null && map.ValueClass() != null && Map.class.isAssignableFrom(fieldType)) {
                meta.Type = MType.Map;
                meta.KeyType = map.KeyClass();
                meta.ValueType = map.ValueClass();
            }
            return meta;
        }

        return meta;
    }

    public static MMeta GetClassMeta(Class<?> type) {
        MMeta meta = new MMeta();
        meta.MainType = type;
        if (type == String.class) {
            meta.Type = MType.String;
            return meta;
        }
        if (type == Boolean.class || type == boolean.class) {
            meta.Type = MType.Bool;
            return meta;
        }
        if (type == Byte.class || type == byte.class) {
            meta.Type = MType.Byte;
            return meta;
        }
        if (type == Short.class || type == short.class) {
            meta.Type = MType.I16;
            return meta;
        }
        if (type == Integer.class || type == int.class) {
            meta.Type = MType.I32;
            return meta;
        }
        if (type == Long.class || type == long.class) {
            meta.Type = MType.I64;
            return meta;
        }
        if (type == Float.class || type == float.class) {
            meta.Type = MType.Float;
            return meta;
        }
        if (type == Double.class || type == double.class) {
            meta.Type = MType.Double;
            return meta;
        }
        if (type == byte[].class) {
            meta.Type = MType.Binary;
            return meta;
        }

        if (type.isArray()) {
            meta.Type = MType.List;
            Class<?> elemType = type.getComponentType();
            meta.ValueType = elemType;
            return meta;
        }

        MsgStruct struct = type.getAnnotation(MsgStruct.class);
        if (struct != null) {
            meta.Type = MType.Struct;
            return meta;
        }

        return meta;
    }
}
