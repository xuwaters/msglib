package msglib.proto.io;

import java.io.IOException;
import java.nio.charset.Charset;

import msglib.proto.types.MField;
import msglib.proto.types.MList;
import msglib.proto.types.MMap;
import msglib.proto.types.MSet;
import msglib.proto.types.MStruct;

public class MBinaryProto extends MProto {

    public static Charset UTF8 = Charset.forName("UTF8");

    protected IDataBuffer databuffer;

    public IDataBuffer getDataBuffer() {
        return this.databuffer;
    }

    public MBinaryProto(IDataBuffer buffer) {
        this.databuffer = buffer;
    }

    protected void ReadAll(byte[] buffer, int offset, int length) throws MException {
        try {
            int len = databuffer.Read(buffer, offset, length);
            if (len < length)
                throw new MException("Read EOF");
        }
        catch (IOException e) {
            throw new MException("ReadAll IO Error", e);
        }
    }

    protected void WriteAll(byte[] buffer, int offset, int length) throws MException {
        try {
            databuffer.Write(buffer, offset, length);
        }
        catch (IOException e) {
            throw new MException("WriteAll IO Error", e);
        }
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

    byte[] byteRawBuf = new byte[1];

    public byte ReadByte() throws MException {
        ReadAll(byteRawBuf, 0, 1);
        return (byte) byteRawBuf[0];
    }

    public void WriteByte(byte b) throws MException {
        byteRawBuf[0] = (byte) b;
        WriteAll(byteRawBuf, 0, 1);
    }

    public short ReadI16() throws MException {
        return (short) zigzagToInt(ReadVarint32());
    }

    public void WriteI16(short i16) throws MException {
        WriteVarint32(intToZigZag(i16));
    }

    public int ReadI32() throws MException {
        return zigzagToInt(ReadVarint32());
    }

    public void WriteI32(int i32) throws MException {
        WriteVarint32(intToZigZag(i32));
    }

    public long ReadI64() throws MException {
        return zigzagToLong(ReadVarint64());
    }

    public void WriteI64(long i64) throws MException {
        WriteVarint64(longToZigzag(i64));
    }

    byte[] floatBuffer = new byte[4];

    public float ReadFloat() throws MException {
        ReadAll(floatBuffer, 0, floatBuffer.length);
        int val = bytesToInt(floatBuffer);
        float res = Float.intBitsToFloat(val);
        return res;
    }

    public void WriteFloat(float f) throws MException {
        int val = Float.floatToIntBits(f);
        intToBytes(val, floatBuffer);
        WriteAll(floatBuffer, 0, floatBuffer.length);
    }

    byte[] doubleBuffer = new byte[8];

    public double ReadDouble() throws MException {
        ReadAll(doubleBuffer, 0, 8);
        long val = bytesToLong(doubleBuffer);
        double res = Double.longBitsToDouble(val);
        return res;
    }

    public void WriteDouble(double d) throws MException {
        long val = Double.doubleToLongBits(d);
        longToBytes(val, doubleBuffer);
        WriteAll(doubleBuffer, 0, doubleBuffer.length);
    }

    public byte[] ReadBinary() throws MException {
        int len = (int) ReadVarint32();
        if (len == 0)
            return new byte[0];
        byte[] res = new byte[len];
        ReadAll(res, 0, len);
        return res;
    }

    public void WriteBinary(byte[] binary, int offset, int length) throws MException {
        WriteVarint32((int) length);
        WriteAll(binary, offset, length);
    }

    @Override
    public String ReadString() throws MException {
        byte[] bytes = ReadBinary();
        return new String(bytes, UTF8);
    }

    public void WriteString(String str) throws MException {
        byte[] buffer = str.getBytes(UTF8);
        WriteBinary(buffer);
    }

    private void intToBytes(int n, byte[] buf) {
        buf[0] = (byte) (n & 0xff);
        buf[1] = (byte) ((n >> 8) & 0xff);
        buf[2] = (byte) ((n >> 16) & 0xff);
        buf[3] = (byte) ((n >> 24) & 0xff);
    }

    private int bytesToInt(byte[] bytes) {
        return ((bytes[3] & 0xff) << 24) | ((bytes[2] & 0xff) << 16) | ((bytes[1] & 0xff) << 8) | ((bytes[0] & 0xff));
    }

    private void longToBytes(long n, byte[] buf) {
        buf[0] = (byte) (n & 0xff);
        buf[1] = (byte) ((n >> 8) & 0xff);
        buf[2] = (byte) ((n >> 16) & 0xff);
        buf[3] = (byte) ((n >> 24) & 0xff);
        buf[4] = (byte) ((n >> 32) & 0xff);
        buf[5] = (byte) ((n >> 40) & 0xff);
        buf[6] = (byte) ((n >> 48) & 0xff);
        buf[7] = (byte) ((n >> 56) & 0xff);
    }

    private long bytesToLong(byte[] bytes) {
        return ((bytes[7] & 0xffL) << 56) | ((bytes[6] & 0xffL) << 48) | ((bytes[5] & 0xffL) << 40)
                | ((bytes[4] & 0xffL) << 32) | ((bytes[3] & 0xffL) << 24) | ((bytes[2] & 0xffL) << 16)
                | ((bytes[1] & 0xffL) << 8) | ((bytes[0] & 0xffL));
    }

    byte[] varint64out = new byte[10];

    private void WriteVarint64(long n) throws MException {
        // byte[] varint64out = local_varint64out.get();
        int idx = 0;
        while (true) {
            if ((n & ~(long) 0x7FL) == 0) {
                varint64out[idx++] = (byte) n;
                break;
            }
            else {
                varint64out[idx++] = ((byte) ((n & 0x7F) | 0x80));
                n >>>= 7;
            }
        }
        WriteAll(varint64out, 0, idx);
    }

    byte[] i32buf = new byte[5];

    private void WriteVarint32(int n) throws MException {
        // byte[] i32buf = local_i32buf.get();
        int idx = 0;
        while (true) {
            if ((n & ~0x7F) == 0) {
                i32buf[idx++] = (byte) n;
                break;
            }
            else {
                i32buf[idx++] = (byte) ((n & 0x7F) | 0x80);
                n >>>= 7;
            }
        }
        WriteAll(i32buf, 0, idx);
    }

    private long ReadVarint64() throws MException {
        int shift = 0;
        long result = 0;
        while (true) {
            byte b = (byte) ReadByte();
            result |= (long) (b & 0x7f) << shift;
            if ((b & 0x80) != 0x80)
                break;
            shift += 7;
        }

        return result;
    }

    private int ReadVarint32() throws MException {
        int result = 0;
        int shift = 0;
        while (true) {
            byte b = (byte) ReadByte();
            result |= (int) (b & 0x7f) << shift;
            if ((b & 0x80) != 0x80)
                break;
            shift += 7;
        }
        return result;
    }

    private long longToZigzag(long n) {
        return (n << 1) ^ (n >> 63);
    }

    private long zigzagToLong(long n) {
        return (n >>> 1) ^ (-(n & 1));
    }

    private int intToZigZag(int n) {
        return (n << 1) ^ (n >> 31);
    }

    private int zigzagToInt(int n) {
        return (n >>> 1) ^ (-(n & 1));
    }
}
