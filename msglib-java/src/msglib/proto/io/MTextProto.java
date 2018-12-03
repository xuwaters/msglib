package msglib.proto.io;

import java.io.IOException;

import msglib.proto.types.MField;
import msglib.proto.types.MList;
import msglib.proto.types.MMap;
import msglib.proto.types.MSet;
import msglib.proto.types.MStruct;

public class MTextProto extends MProto {

    private final ITextBuffer buffer;

    public MTextProto(ITextBuffer buff) {
        this.buffer = buff;
    }

    public MStruct ReadStructBegin() throws MException {
        MStruct str = new MStruct();
        return str;
    }

    public void WriteStructBegin(MStruct struc) throws MException {
    }

    public MField ReadFieldBegin() throws MException {
        MField field = new MField();
        int idAndType = ReadVarint32();
        field.Type = (int) (idAndType & 0x0F);
        field.ID = (int) (idAndType >> 4);
        return field;
    }

    public void WriteFieldBegin(MField field) throws MException {
        int id = (int) field.ID;
        int type = (int) field.Type;
        int idAndType = ((id << 4) | (type & 0x0F));
        WriteVarint32(idAndType);
    }

    public MMap ReadMapBegin() throws MException {
        MMap map = new MMap();
        map.Count = (int) ReadVarint32();
        byte type = (byte) ReadByte();
        map.KeyType = (int) (type & 0x0F);
        map.ValueType = (int) (type >> 4);
        return map;
    }

    public void WriteMapBegin(MMap map) throws MException {
        WriteVarint32((int) map.Count);
        byte val = (byte) map.ValueType;
        byte key = (byte) map.KeyType;
        byte type = (byte) (((val & 0x0F) << 4) | (key & 0x0F));
        WriteByte((byte) type);
    }

    public MList ReadListBegin() throws MException {
        MList list = new MList();
        int countAndType = ReadVarint32();
        list.ElementType = (int) (countAndType & 0x0F);
        list.Count = (int) (countAndType >> 4);
        return list;
    }

    public void WriteListBegin(MList list) throws MException {
        int count = (int) list.Count;
        int type = (int) list.ElementType;
        int countAndType = ((count << 4) | (type & 0x0F));
        WriteVarint32(countAndType);
    }

    public MSet ReadSetBegin() throws MException {
        MList list = ReadListBegin();
        MSet set = new MSet(list.ElementType, list.Count);
        return set;
    }

    public void WriteSetBegin(MSet set) throws MException {
        WriteListBegin(new MList(set.ElementType, set.Count));
    }

    public boolean ReadBool() throws MException {
        byte val = ReadByte();
        return val == 1;
    }

    public void WriteBool(boolean b) throws MException {
        byte val = (byte) (b ? 1 : 0);
        WriteByte(val);
    }

    public byte ReadByte() throws MException {
        return (byte) ReadVarint32();
    }

    public void WriteByte(byte b) throws MException {
        WriteVarint32(b);
    }

    public short ReadI16() throws MException {
        return (short) ReadVarint32();
    }

    public void WriteI16(short i16) throws MException {
        WriteVarint32(i16);
    }

    public int ReadI32() throws MException {
        return ReadVarint32();
    }

    public void WriteI32(int i32) throws MException {
        WriteVarint32(i32);
    }

    public long ReadI64() throws MException {
        return ReadVarint64();
    }

    public void WriteI64(long i64) throws MException {
        WriteVarint64(i64);
    }

    public float ReadFloat() throws MException {
        return (float) ReadDouble();
    }

    public void WriteFloat(float f) throws MException {
        WriteDouble(f);
    }

    public double ReadDouble() throws MException {
        return Double.parseDouble(readText());
    }

    public void WriteDouble(double d) throws MException {
        writeText(String.valueOf(d));
    }

    public byte[] ReadBinary() throws MException {
        throw new MException("Not Implemented");
    }

    public void WriteBinary(byte[] binary, int offset, int length) throws MException {
        // WriteVarint32((int) length);
        // WriteAll(binary, offset, length);
        throw new MException("Not Implemented");
    }

    @Override
    public String ReadString() throws MException {
        return readText();
    }

    @Override
    public void WriteString(String str) throws MException {
        writeText(str);
    }

    private void WriteVarint64(long n) throws MException {
        writeText(String.valueOf(n));
    }

    private void WriteVarint32(int n) throws MException {
        writeText(String.valueOf(n));
    }

    private long ReadVarint64() throws MException {
        String text = readText();
        try {
            return Long.parseLong(text);
        }
        catch (NumberFormatException e) {
            throw new MException("", e);
        }
    }

    private int ReadVarint32() throws MException {
        String text = readText();
        try {
            return Integer.parseInt(text);
        }
        catch (NumberFormatException e) {
            throw new MException("", e);
        }
    }

    private String readText() throws MException {
        try {
            String str = buffer.Read();
            if (str == null) {
                throw new MException("End of Buffer");
            }
            return str;
        }
        catch (IOException e) {
            throw new MException("", e);
        }
    }

    private void writeText(String text) throws MException {
        try {
            buffer.Write(text);
        }
        catch (IOException e) {
            throw new MException("", e);
        }
    }

}
