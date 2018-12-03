package msglib.proto.io;

import msglib.proto.types.MField;
import msglib.proto.types.MList;
import msglib.proto.types.MMap;
import msglib.proto.types.MSet;
import msglib.proto.types.MStruct;
import msglib.proto.types.MType;

public abstract class MProto {

    public abstract MStruct ReadStructBegin() throws MException;

    public abstract MField ReadFieldBegin() throws MException;

    public abstract MMap ReadMapBegin() throws MException;

    public abstract MList ReadListBegin() throws MException;

    public abstract MSet ReadSetBegin() throws MException;

    public abstract boolean ReadBool() throws MException;

    public abstract byte ReadByte() throws MException;

    public abstract short ReadI16() throws MException;

    public abstract int ReadI32() throws MException;

    public abstract long ReadI64() throws MException;

    public abstract float ReadFloat() throws MException;

    public abstract double ReadDouble() throws MException;

    public abstract byte[] ReadBinary() throws MException;

    public abstract String ReadString() throws MException ;

    public abstract void WriteStructBegin(MStruct struc) throws MException;

    public abstract void WriteFieldBegin(MField field) throws MException;

    public void WriteFieldStop() throws MException {
        WriteFieldBegin(new MField("", MType.Null, 0));
    }

    public abstract void WriteMapBegin(MMap map) throws MException;

    public abstract void WriteListBegin(MList list) throws MException;

    public abstract void WriteSetBegin(MSet set) throws MException;

    public abstract void WriteBool(boolean b) throws MException;

    public abstract void WriteByte(byte b) throws MException;

    public abstract void WriteI16(short i16) throws MException;

    public abstract void WriteI32(int i32) throws MException;

    public abstract void WriteI64(long i64) throws MException;

    public abstract void WriteFloat(float f) throws MException;

    public abstract void WriteDouble(double d) throws MException;

    public abstract void WriteBinary(byte[] binary, int offset, int length) throws MException;

    public void WriteBinary(byte[] binary) throws MException {
        WriteBinary(binary, 0, binary.length);
    }

    public abstract void WriteString(String str) throws MException;
}
